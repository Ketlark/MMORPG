package game

import (
	"context"

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
