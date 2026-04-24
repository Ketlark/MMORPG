package game

import (
	"context"

	"mmorpg/server/internal/game/domain"

	pb "mmorpg/server/api/proto/gen"
)

// EventBroadcaster sends game events to connected players.
type EventBroadcaster interface {
	Broadcast(event *pb.GameEvent, excludePlayerID string)
}

// PositionStore handles player position persistence.
type PositionStore interface {
	LoadPosition(ctx context.Context, playerID string) (x, y int32, err error)
	SavePosition(ctx context.Context, playerID string, x, y int32) error
}

// FighterData is the port-level DTO for creating a combat fighter.
type FighterData struct {
	ID        string
	Username  string
	Team      int
	X, Y      int
	Health    int
	MaxHealth int
	AP, MP    int
}

// CellData is the port-level DTO for combat grid cells.
type CellData struct {
	X, Y     int
	Walkable bool
}

// CombatStarter is the hexagonal port for initiating combat from the game world.
type CombatStarter interface {
	CreateCombatWithGrid(fighters []FighterData, cells []CellData, width, height int) string
	RunMobAI(combatID, mobID string, behavior domain.MobBehavior)
}
