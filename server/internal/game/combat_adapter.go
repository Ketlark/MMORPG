package game

import (
	"mmorpg/server/internal/combat"
	"mmorpg/server/internal/game/domain"

	pb "mmorpg/server/api/proto/gen"
)

// CombatAdapter is the hexagonal adapter that bridges the game domain port
// CombatStarter with the concrete combat.CombatService.
type CombatAdapter struct {
	svc *combat.CombatService
}

// NewCombatAdapter creates a new adapter wrapping the combat service.
func NewCombatAdapter(svc *combat.CombatService) *CombatAdapter {
	return &CombatAdapter{svc: svc}
}

// CreateCombatWithGrid implements CombatStarter.
func (a *CombatAdapter) CreateCombatWithGrid(
	fighters []FighterData,
	cells []CellData,
	width, height int,
) string {
	cf := make([]*combat.CombatFighter, len(fighters))
	for i, f := range fighters {
		cf[i] = &combat.CombatFighter{
			PlayerID:  f.ID,
			Username:  f.Username,
			Team:      f.Team,
			X:         f.X,
			Y:         f.Y,
			Health:    f.Health,
			MaxHealth: f.MaxHealth,
			AP:        f.AP,
			MP:        f.MP,
		}
	}

	protoCells := make([]*pb.CombatCell, len(cells))
	for i, c := range cells {
		protoCells[i] = &pb.CombatCell{
			X:        int32(c.X),
			Y:        int32(c.Y),
			Walkable: c.Walkable,
		}
	}

	return a.svc.CreateCombatWithGrid(cf, protoCells, width, height)
}

// RunMobAI implements CombatStarter.
func (a *CombatAdapter) RunMobAI(combatID, mobID string, behavior domain.MobBehavior) {
	a.svc.RunMobAI(combatID, mobID, behavior)
}
