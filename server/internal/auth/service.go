package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	pb "mmorpg/server/api/proto/gen"
)

// AuthService implements the gRPC AuthServiceServer.
type AuthService struct {
	pb.UnimplementedAuthServiceServer

	db        *sql.DB
	jwtSecret string
}

// NewAuthService creates a new AuthService with the given database connection and JWT secret.
func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register handles user registration: validates input, hashes the password,
// and inserts a new account into the database.
func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validate input
	username := strings.TrimSpace(req.GetUsername())
	email := strings.TrimSpace(req.GetEmail())
	password := req.GetPassword()

	if username == "" {
		return &pb.RegisterResponse{
			Success: false,
			Error:   "username is required",
		}, nil
	}
	if email == "" {
		return &pb.RegisterResponse{
			Success: false,
			Error:   "email is required",
		}, nil
	}
	if password == "" {
		return &pb.RegisterResponse{
			Success: false,
			Error:   "password is required",
		}, nil
	}
	if len(password) < 6 {
		return &pb.RegisterResponse{
			Success: false,
			Error:   "password must be at least 6 characters",
		}, nil
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Error:   "failed to hash password",
		}, nil
	}

	// Generate UUID for the new account
	accountID := uuid.New().String()

	// Insert into accounts table
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO accounts (id, username, email, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`,
		accountID, username, email, string(hashedPassword), time.Now().UTC(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return &pb.RegisterResponse{
				Success: false,
				Error:   "username or email already exists",
			}, nil
		}
		return &pb.RegisterResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create account: %v", err),
		}, nil
	}

	return &pb.RegisterResponse{
		Success: true,
	}, nil
}

// Login handles user authentication: looks up the account by username,
// verifies the password, and returns a signed JWT token on success.
func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	username := strings.TrimSpace(req.GetUsername())
	password := req.GetPassword()

	if username == "" {
		return &pb.LoginResponse{
			Success: false,
			Error:   "username is required",
		}, nil
	}
	if password == "" {
		return &pb.LoginResponse{
			Success: false,
			Error:   "password is required",
		}, nil
	}

	// Find account by username
	var id string
	var passwordHash string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, password_hash FROM accounts WHERE username = $1`,
		username,
	).Scan(&id, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.LoginResponse{
				Success: false,
				Error:   "invalid username or password",
			}, nil
		}
		return &pb.LoginResponse{
			Success: false,
			Error:   fmt.Sprintf("database error: %v", err),
		}, nil
	}

	// Verify bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Error:   "invalid username or password",
		}, nil
	}

	// Generate JWT token with user_id claim and 24h expiry
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"user_id": id,
		"iat":     now.Unix(),
		"exp":     now.Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Error:   "failed to generate token",
		}, nil
	}

	return &pb.LoginResponse{
		Success: true,
		Token:   tokenString,
	}, nil
}
