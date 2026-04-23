package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pb "mmorpg/server/api/proto/gen"
	"mmorpg/server/internal/auth"
	"mmorpg/server/internal/combat"
	"mmorpg/server/internal/config"
	"mmorpg/server/internal/database"
	"mmorpg/server/internal/game"
)

func loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		var addr string
		if p, ok := peer.FromContext(ctx); ok {
			addr = p.Addr.String()
		}
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		log.Printf("[RPC] %s from %s - %v - %v", info.FullMethod, addr, duration, err)
		return resp, err
	}
}

func streamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		var addr string
		if p, ok := peer.FromContext(ss.Context()); ok {
			addr = p.Addr.String()
		}
		err := handler(srv, ss)
		duration := time.Since(start)
		log.Printf("[RPC-Stream] %s from %s - %v - %v", info.FullMethod, addr, duration, err)
		return err
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	visits   map[string][]time.Time
	limit    int
	window   time.Duration
	cleanup  time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visits:  make(map[string][]time.Time),
		limit:   limit,
		window:  window,
		cleanup: window * 2,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	visits := rl.visits[key]
	filtered := visits[:0]
	for _, t := range visits {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	rl.visits[key] = filtered

	if len(filtered) >= rl.limit {
		return false
	}
	rl.visits[key] = append(rl.visits[key], now)
	return true
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for k, v := range rl.visits {
			filtered := v[:0]
			for _, t := range v {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.visits, k)
			} else {
				rl.visits[k] = filtered
			}
		}
		rl.mu.Unlock()
	}
}

func rateLimitInterceptor(rl *rateLimiter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var addr string
		if p, ok := peer.FromContext(ctx); ok {
			addr = p.Addr.String()
		}
		if !rl.allow(addr) {
			return nil, statusRateLimited()
		}
		return handler(ctx, req)
	}
}

func statusRateLimited() error {
	return status.Error(codes.ResourceExhausted, "rate limit exceeded")
}

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	rdb, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	authInterceptor := auth.AuthInterceptor(cfg.JWTSecret)
	rl := newRateLimiter(30, time.Second) // 30 req/s per IP

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(),
			rateLimitInterceptor(rl),
			authInterceptor,
		),
		grpc.ChainStreamInterceptor(
			streamLoggingInterceptor(),
			auth.AuthStreamInterceptor(cfg.JWTSecret),
		),
	)

	pb.RegisterAuthServiceServer(grpcServer, auth.NewAuthService(db, cfg.JWTSecret))
	pb.RegisterGameServiceServer(grpcServer, game.NewService(rdb))
	pb.RegisterCombatServiceServer(grpcServer, combat.NewCombatService())

	listener, err := net.Listen("tcp", cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", cfg.ServerPort, err)
	}

	grpcWebServer := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return cfg.IsOriginAllowed(origin)
		}),
	)

	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if grpcWebServer.IsGrpcWebRequest(r) || grpcWebServer.IsAcceptableGrpcCorsRequest(r) {
				grpcWebServer.ServeHTTP(w, r)
				return
			}
			grpcServer.ServeHTTP(w, r)
		}),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("gRPC-Web server listening on %s", cfg.ServerPort)
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, stopping server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)
	log.Println("Server stopped gracefully")
}
