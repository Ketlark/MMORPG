package combat

import (
	"time"

	"mmorpg/server/internal/game/domain"

	pb "mmorpg/server/api/proto/gen"
)

const (
	mobAITurnPollInterval = 300 * time.Millisecond
	mobAIActionDelay      = 800 * time.Millisecond
)

// RunMobAI runs the AI loop for a mob in a combat.
// It polls until it's the mob's turn, then executes actions based on MobBehavior.
func (s *CombatService) RunMobAI(combatID, mobID string, behavior domain.MobBehavior) {
	for {
		time.Sleep(mobAITurnPollInterval)

		s.mu.RLock()
		combat, exists := s.combats[combatID]
		s.mu.RUnlock()

		if !exists || !combat.Active {
			return
		}

		combat.mu.Lock()
		if len(combat.TurnOrder) == 0 || combat.TurnOrder[combat.CurrentTurn] != mobID {
			combat.mu.Unlock()
			continue
		}

		fighter := combat.Fighters[mobID]
		if fighter == nil || fighter.Health <= 0 {
			combat.mu.Unlock()
			return
		}

		// Build domain CombatView snapshot.
		view := combat.toDomainView(fighter)
		combat.mu.Unlock()

		action := behavior.Decide(view)

		switched := s.executeMobAction(combat, fighter, action)
		if !switched {
			// Mob ended its turn or couldn't act.
			time.Sleep(mobAIActionDelay)
			s.advanceMobTurn(combat)
		}
	}
}

// executeMobAction executes one action and returns true if the turn was advanced.
func (s *CombatService) executeMobAction(combat *Combat, fighter *CombatFighter, action domain.MobAction) bool {
	switch action.Type {
	case domain.ActionAttack:
		time.Sleep(mobAIActionDelay)
		combat.mu.Lock()
		result, event, endEvent := s.handleSpellLocked(combat, fighter, &pb.SpellAction{
			SpellId: "mob_attack",
			TargetX: int32(action.TargetX),
			TargetY: int32(action.TargetY),
		})
		combat.mu.Unlock()

		if result.Success {
			if event != nil {
				broadcastCombatEvent(combat, event)
			}
			if endEvent != nil {
				broadcastCombatEvent(combat, endEvent)
				return true // Combat ended.
			}
			// Attack was the action, end turn.
			s.advanceMobTurn(combat)
			return true
		}

	case domain.ActionMove:
		time.Sleep(mobAIActionDelay)
		combat.mu.Lock()
		result, event := s.handleMoveLocked(combat, fighter, &pb.MoveAction{
			TargetX: int32(action.TargetX),
			TargetY: int32(action.TargetY),
		})
		combat.mu.Unlock()

		if result.Success && event != nil {
			broadcastCombatEvent(combat, event)
			// Only move once per turn unless more MP remain.
			// TofuBehavior checks remaining MP on next poll, but for simplicity end turn.
			s.advanceMobTurn(combat)
			return true
		}

	case domain.ActionEndTurn:
		time.Sleep(mobAIActionDelay)
		s.advanceMobTurn(combat)
		return true
	}

	return false
}

// advanceMobTurn advances the turn after the mob's action.
func (s *CombatService) advanceMobTurn(combat *Combat) {
	combat.mu.Lock()
	turnEndedEvent, turnStartedEvent := s.advanceTurnLocked(combat)
	combat.mu.Unlock()

	broadcastCombatEvent(combat, turnEndedEvent)
	if turnStartedEvent != nil {
		broadcastCombatEvent(combat, turnStartedEvent)
	}
}

// toDomainView builds a domain CombatView from the combat state.
// Must be called with combat.mu held.
func (c *Combat) toDomainView(fighter *CombatFighter) domain.CombatView {
	fighters := make([]domain.FighterView, 0, len(c.Fighters))
	for _, f := range c.Fighters {
		fighters = append(fighters, domain.FighterView{
			ID:       f.PlayerID,
			Team:     f.Team,
			Position: domain.Position{X: int32(f.X), Y: int32(f.Y)},
			Health:   f.Health,
		})
	}

	return domain.CombatView{
		FighterPos: domain.Position{X: int32(fighter.X), Y: int32(fighter.Y)},
		FighterAP:  fighter.AP,
		FighterMP:  fighter.MP,
		Fighters:   fighters,
		IsWalkable: func(x, y int) bool {
			return isCellWalkable(c, x, y)
		},
	}
}
