package game

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	pb "mmorpg/server/api/proto/gen"

	"mmorpg/server/internal/auth"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PlayerState holds the in-memory state for a connected player.
type PlayerState struct {
	PlayerID      string
	Username      string
	CharacterName string
	X             int32
	Y             int32
	Health        int32
	MaxHealth     int32
	AP            int32
	MP            int32
}

const (
	defaultHealth    = 100
	defaultMaxHealth = 100
	defaultAP        = 6
	defaultMP        = 3
	spawnX           = 9
	spawnY           = 14
)

// playerStream wraps a gRPC server stream for broadcasting events to a player.
type playerStream struct {
	stream grpc.ServerStreamingServer[pb.GameEvent]
}

// Service implements the GameService gRPC server.
type Service struct {
	pb.UnimplementedGameServiceServer

	rdb *redis.Client
	mu  sync.RWMutex

	players map[string]*PlayerState
	streams map[string]*playerStream
	mapData *pb.MapData
}

// NewService creates a new game service instance.
func NewService(rdb *redis.Client) *Service {
	return &Service{
		rdb:     rdb,
		players: make(map[string]*PlayerState),
		streams: make(map[string]*playerStream),
		mapData: GenerateTestMap(),
	}
}

// Connect is a server-side streaming RPC. When a player connects:
//  1. Their character position is loaded from Redis (or a default spawn point is chosen).
//  2. They are registered in the in-memory player map.
//  3. They receive the MapData event.
//  4. They receive PlayerConnected events for every already-connected player.
//  5. All other connected players receive a PlayerConnected event for the newcomer.
//  6. The stream stays open until the client disconnects (context cancelled).
func (s *Service) Connect(req *pb.ConnectRequest, stream grpc.ServerStreamingServer[pb.GameEvent]) error {
	ctx := stream.Context()

	playerID := auth.GetUserID(ctx)
	if playerID == "" {
		return status.Error(codes.Unauthenticated, "not authenticated")
	}

	characterName := req.CharacterName
	if characterName == "" {
		characterName = fmt.Sprintf("Hero-%s", playerID[len(playerID)-6:])
	}
	username := characterName

	// Load or create position from Redis.
	posKey := fmt.Sprintf("player:%s:pos", playerID)
	x, y, err := s.loadPosition(ctx, posKey)
	if err != nil {
		log.Printf("[Game] Redis position load failed for %s: %v, using default spawn", playerID, err)
		// Find a walkable spawn cell near the village entrance (row 14, col 9).
		x, y = FindNearestWalkableCell(s.mapData, spawnX, spawnY)
	}

	// Build the player state.
	player := &PlayerState{
		PlayerID:      playerID,
		Username:      username,
		CharacterName: characterName,
		X:             x,
		Y:             y,
		Health:        defaultHealth,
		MaxHealth:     defaultMaxHealth,
		AP:            defaultAP,
		MP:            defaultMP,
	}

	// Register the player and their stream.
	s.mu.Lock()
	s.players[playerID] = player
	s.streams[playerID] = &playerStream{stream: stream}
	s.mu.Unlock()

	// Ensure cleanup on any exit path.
	defer func() {
		s.mu.Lock()
		delete(s.players, playerID)
		delete(s.streams, playerID)
		s.mu.Unlock()

		// Broadcast disconnection to remaining players.
		s.broadcast(&pb.GameEvent{
			Event: &pb.GameEvent_PlayerDisconnected{
				PlayerDisconnected: &pb.PlayerDisconnected{
					PlayerId: playerID,
				},
			},
		}, playerID)

		log.Printf("[Game] Player %s (%s) disconnected", playerID, characterName)
	}()

	// 1. Send MapData to the newly connected player.
	if err := stream.Send(&pb.GameEvent{
		Event: &pb.GameEvent_MapData{
			MapData: s.mapData,
		},
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send map data: %v", err)
	}

	// 2. Send PlayerConnected events for all existing players.
	s.mu.RLock()
	existingPlayers := make([]*PlayerState, 0, len(s.players)-1)
	for _, p := range s.players {
		if p.PlayerID != playerID {
			existingPlayers = append(existingPlayers, p)
		}
	}
	s.mu.RUnlock()

	for _, p := range existingPlayers {
		if err := stream.Send(&pb.GameEvent{
			Event: &pb.GameEvent_PlayerConnected{
				PlayerConnected: &pb.PlayerConnected{
					PlayerId: p.PlayerID,
					Username: p.Username,
					X:        p.X,
					Y:        p.Y,
				},
			},
		}); err != nil {
			log.Printf("[Game] Failed to send existing player %s to %s: %v", p.PlayerID, playerID, err)
		}
	}

	// 3. Send this player's own info to the connecting client.
	if err := stream.Send(&pb.GameEvent{
		Event: &pb.GameEvent_PlayerConnected{
			PlayerConnected: &pb.PlayerConnected{
				PlayerId: playerID,
				Username: username,
				X:        x,
				Y:        y,
			},
		},
	}); err != nil {
		log.Printf("[Game] Failed to send own PlayerConnected to %s: %v", playerID, err)
	}

	// 4. Broadcast this player's connection to everyone else.
	s.broadcast(&pb.GameEvent{
		Event: &pb.GameEvent_PlayerConnected{
			PlayerConnected: &pb.PlayerConnected{
				PlayerId: playerID,
				Username: username,
				X:        x,
				Y:        y,
			},
		},
	}, playerID)

	log.Printf("[Game] Player %s (%s) connected at (%d,%d)", playerID, characterName, x, y)

	// 4. Keep the stream open, waiting for context cancellation (client disconnect).
	<-ctx.Done()
	return nil
}

// Move handles a player movement request. It validates the target cell is
// within bounds and walkable, checks the Manhattan distance against available
// movement points, updates position in Redis and memory, then broadcasts the
// move event to all connected players.
func (s *Service) Move(ctx context.Context, req *pb.MoveRequest) (*pb.MoveResponse, error) {
	playerID := auth.GetUserID(ctx)
	if playerID == "" {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	s.mu.RLock()
	player, ok := s.players[playerID]
	s.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "player %s not connected", playerID)
	}

	targetX := req.TargetX
	targetY := req.TargetY

	// Validate map bounds.
	if targetX < 0 || targetX >= s.mapData.Width || targetY < 0 || targetY >= s.mapData.Height {
		return &pb.MoveResponse{
			Success:        false,
			X:              player.X,
			Y:              player.Y,
			ActionPoints:   player.AP,
			MovementPoints: player.MP,
		}, nil
	}

	// Validate walkable terrain.
	if !IsWalkable(s.mapData, targetX, targetY) {
		return &pb.MoveResponse{
			Success:        false,
			X:              player.X,
			Y:              player.Y,
			ActionPoints:   player.AP,
			MovementPoints: player.MP,
		}, nil
	}

	if targetX == player.X && targetY == player.Y {
		return &pb.MoveResponse{
			Success:        false,
			X:              player.X,
			Y:              player.Y,
			ActionPoints:   player.AP,
			MovementPoints: player.MP,
		}, nil
	}

	// Compute A* path.
	path := FindPath(s.mapData, player.X, player.Y, targetX, targetY)
	if path == nil {
		return &pb.MoveResponse{
			Success:        false,
			X:              player.X,
			Y:              player.Y,
			ActionPoints:   player.AP,
			MovementPoints: player.MP,
		}, nil
	}

	finalNode := path[len(path)-1]
	pathCost := int32(len(path) - 1)

	s.mu.Lock()
	player.X = finalNode.X
	player.Y = finalNode.Y
	s.mu.Unlock()

	// Persist new position to Redis.
	posKey := fmt.Sprintf("player:%s:pos", playerID)
	s.persistPosition(ctx, posKey, finalNode.X, finalNode.Y)

	// Broadcast move event with full path.
	s.broadcast(&pb.GameEvent{
		Event: &pb.GameEvent_PlayerMoved{
			PlayerMoved: &pb.PlayerMoved{
				PlayerId: playerID,
				X:        finalNode.X,
				Y:        finalNode.Y,
				Path:     path,
			},
		},
	}, "")

	return &pb.MoveResponse{
		Success:        true,
		X:              finalNode.X,
		Y:              finalNode.Y,
		ActionPoints:   player.AP,
		MovementPoints: player.MP,
		Path:           path,
		PathCost:       pathCost,
	}, nil
}

// Chat broadcasts a chat message from a player to all connected players.
func (s *Service) Chat(ctx context.Context, req *pb.ChatRequest) (*emptypb.Empty, error) {
	playerID := auth.GetUserID(ctx)
	if playerID == "" {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	s.mu.RLock()
	player, ok := s.players[playerID]
	s.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "player %s not connected", playerID)
	}

	msg := sanitizeChatMessage(req.Message)
	if len(msg) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "message cannot be empty")
	}

	s.broadcast(&pb.GameEvent{
		Event: &pb.GameEvent_ChatMessage{
			ChatMessage: &pb.ChatMessage{
				PlayerId: playerID,
				Username: player.Username,
				Message:  msg,
			},
		},
	}, "")

	return &emptypb.Empty{}, nil
}

// broadcast sends a GameEvent to all connected players, optionally excluding
// the player with excludeID. An empty excludeID sends to everyone.
func (s *Service) broadcast(event *pb.GameEvent, excludeID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for pid, ps := range s.streams {
		if pid == excludeID {
			continue
		}
		if err := ps.stream.Send(event); err != nil {
			log.Printf("[Game] broadcast to %s failed: %v", pid, err)
		}
	}
}

// loadPosition reads a player's last known position from Redis.
// Returns (x, y, nil) on success. Returns an error if no position is stored.
func (s *Service) loadPosition(ctx context.Context, key string) (int32, int32, error) {
	result, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	if len(result) == 0 {
		return 0, 0, fmt.Errorf("no position stored for key %s", key)
	}

	x, err := strconv.ParseInt(result["x"], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid x in Redis: %w", err)
	}
	y, err := strconv.ParseInt(result["y"], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid y in Redis: %w", err)
	}

	return int32(x), int32(y), nil
}

// persistPosition writes a player's position to Redis as a hash.
func (s *Service) persistPosition(ctx context.Context, key string, x, y int32) {
	if err := s.rdb.HSet(ctx, key, "x", x, "y", y).Err(); err != nil {
		log.Printf("[Game] Failed to persist position for %s: %v", key, err)
	}
}

// Compile-time interface check.
var _ pb.GameServiceServer = (*Service)(nil)

const maxChatLength = 500

func sanitizeChatMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if len(msg) > maxChatLength {
		msg = msg[:maxChatLength]
	}
	var b strings.Builder
	b.Grow(len(msg))
	for _, r := range msg {
		if r < 0x20 && r != '\n' && r != '\t' {
			continue
		}
		switch r {
		case '<', '>', '&', '"', '\'':
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
