package models

import (
	"time"
)

// Account represents a player account in the system.
type Account struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Character represents a player character in the game world.
type Character struct {
	ID            string `json:"id"`
	AccountID     string `json:"account_id"`
	Name          string `json:"name"`
	Level         int    `json:"level"`
	Experience    int64  `json:"experience"`
	Health        int32  `json:"health"`
	MaxHealth     int32  `json:"max_health"`
	ActionPoints  int32  `json:"action_points"`
	MovementPoints int32 `json:"movement_points"`
	X             int32  `json:"x"`
	Y             int32  `json:"y"`
}

// NewCharacter creates a Character with default values for a new player.
func NewCharacter(id, accountID, name string, x, y int32) Character {
	return Character{
		ID:            id,
		AccountID:     accountID,
		Name:          name,
		Level:         1,
		Experience:    0,
		Health:        100,
		MaxHealth:     100,
		ActionPoints:  6,
		MovementPoints: 3,
		X:             x,
		Y:             y,
	}
}

// ElementType represents the elemental affinity of a spell.
type ElementType int

const (
	ElementNeutral ElementType = 0
	ElementFire    ElementType = 1
	ElementWater   ElementType = 2
	ElementEarth   ElementType = 3
	ElementWind    ElementType = 4
)

// Spell represents a combat spell that a character can cast.
type Spell struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	APCost      int32       `json:"ap_cost"`
	Damage      int32       `json:"damage"`
	Range       int32       `json:"range"`
	ElementType ElementType  `json:"element_type"`
}

// TerrainType represents the type of terrain on a map cell.
type TerrainType int

const (
	TerrainGrass TerrainType = 0
	TerrainWater TerrainType = 1
	TerrainWall  TerrainType = 2
	TerrainRoad  TerrainType = 3
)

// MapCell represents a single cell on the game map.
type MapCell struct {
	X           int32       `json:"x"`
	Y           int32       `json:"y"`
	TerrainType TerrainType `json:"terrain_type"`
	Walkable    bool        `json:"walkable"`
	Elevation   int32       `json:"elevation"`
}
