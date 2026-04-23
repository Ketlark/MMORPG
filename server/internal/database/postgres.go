package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"mmorpg/server/internal/config"
)

// Connect opens a connection to PostgreSQL and verifies connectivity.
func Connect(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}

// Migrate creates the required tables if they do not already exist.
func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id UUID PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS characters (
			id UUID PRIMARY KEY,
			account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			name VARCHAR(50) UNIQUE NOT NULL,
			level INTEGER NOT NULL DEFAULT 1,
			experience BIGINT NOT NULL DEFAULT 0,
			health INTEGER NOT NULL DEFAULT 100,
			max_health INTEGER NOT NULL DEFAULT 100,
			action_points INTEGER NOT NULL DEFAULT 6,
			movement_points INTEGER NOT NULL DEFAULT 3,
			x INTEGER NOT NULL DEFAULT 0,
			y INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS spells (
			id UUID PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			ap_cost INTEGER NOT NULL DEFAULT 0,
			damage INTEGER NOT NULL DEFAULT 0,
			spell_range INTEGER NOT NULL DEFAULT 1,
			element_type INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_characters_account_id ON characters(account_id)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute migration: %w\nquery: %s", err, q)
		}
	}

	log.Println("Database migration completed")
	return nil
}
