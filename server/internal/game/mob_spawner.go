package game

import (
	"sync"
	"time"

	"mmorpg/server/internal/game/domain"

	pb "mmorpg/server/api/proto/gen"
)

// MobSpawner manages the lifecycle of mobs on the world map.
type MobSpawner struct {
	mu            sync.RWMutex
	mobs          map[string]*domain.Mob
	factory       *domain.MobFactory
	broadcaster   EventBroadcaster
	combatStarter CombatStarter
	mapData       *pb.MapData
}

// NewMobSpawner creates a new spawner with the given dependencies.
func NewMobSpawner(
	factory *domain.MobFactory,
	broadcaster EventBroadcaster,
	combatStarter CombatStarter,
	mapData *pb.MapData,
) *MobSpawner {
	return &MobSpawner{
		mobs:          make(map[string]*domain.Mob),
		factory:       factory,
		broadcaster:   broadcaster,
		combatStarter: combatStarter,
		mapData:       mapData,
	}
}

// SpawnDefaultMobs creates the initial set of mobs for the world map.
func (s *MobSpawner) SpawnDefaultMobs() {
	s.spawnMob(domain.MobTypeTofu, domain.Position{X: 3, Y: 2})
}

func (s *MobSpawner) spawnMob(mobType domain.MobType, pos domain.Position) *domain.Mob {
	mob := s.factory.Create(mobType, pos)

	s.mu.Lock()
	s.mobs[mob.ID] = mob
	s.mu.Unlock()

	if s.broadcaster != nil {
		s.broadcaster.Broadcast(&pb.GameEvent{
			Event: &pb.GameEvent_MobSpawned{
				MobSpawned: &pb.MobSpawned{
					MobId: mob.ID,
					Name:  mob.Name,
					X:     mob.Position.X,
					Y:     mob.Position.Y,
				},
			},
		}, "")
	}

	return mob
}

// GetMobAt returns the mob at the given grid position, or nil if none.
func (s *MobSpawner) GetMobAt(x, y int32) *domain.Mob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, mob := range s.mobs {
		if mob.Position.X == x && mob.Position.Y == y {
			return mob
		}
	}
	return nil
}

// RemoveMob removes a mob from the world map by ID.
func (s *MobSpawner) RemoveMob(mobID string) *domain.Mob {
	s.mu.Lock()
	mob, ok := s.mobs[mobID]
	if ok {
		delete(s.mobs, mobID)
	}
	s.mu.Unlock()

	if ok && s.broadcaster != nil {
		s.broadcaster.Broadcast(&pb.GameEvent{
			Event: &pb.GameEvent_MobDespawned{
				MobDespawned: &pb.MobDespawned{MobId: mobID},
			},
		}, "")
	}

	return mob
}

// ScheduleRespawn schedules the mob to respawn after the given delay.
func (s *MobSpawner) ScheduleRespawn(mob *domain.Mob, delay time.Duration) {
	time.AfterFunc(delay, func() {
		s.spawnMob(mob.Type, mob.Position)
	})
}

// GetAllMobs returns a snapshot of all currently spawned mobs.
func (s *MobSpawner) GetAllMobs() []*domain.Mob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*domain.Mob, 0, len(s.mobs))
	for _, mob := range s.mobs {
		result = append(result, mob)
	}
	return result
}

// BuildCombatGrid converts the world map data to combat CellData.
func (s *MobSpawner) BuildCombatGrid() ([]CellData, int, int) {
	w := int(s.mapData.Width)
	h := int(s.mapData.Height)
	cells := make([]CellData, 0, w*h)

	for y := int32(0); y < s.mapData.Height; y++ {
		for x := int32(0); x < s.mapData.Width; x++ {
			idx := y*s.mapData.Width + x
			cells = append(cells, CellData{
				X:        int(x),
				Y:        int(y),
				Walkable: s.mapData.Cells[idx].Walkable,
			})
		}
	}

	return cells, w, h
}
