package combat

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	pb "mmorpg/server/api/proto/gen"

	"mmorpg/server/internal/auth"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CombatFighter struct {
	PlayerID   string
	Username   string
	Team       int
	X, Y       int
	Health     int
	MaxHealth  int
	AP         int
	MP         int
	Stream     pb.CombatService_JoinCombatServer
	CancelFunc context.CancelFunc
}

type Combat struct {
	ID          string
	Fighters    map[string]*CombatFighter
	Cells       []*pb.CombatCell
	CellLookup  map[[2]int]bool
	Width       int
	Height      int
	TurnOrder   []string
	CurrentTurn int
	Active      bool
	mu          sync.Mutex
}

type CombatService struct {
	pb.UnimplementedCombatServiceServer
	combats map[string]*Combat
	mu      sync.RWMutex
}

const (
	defaultCombatAP     = 6
	defaultCombatMP     = 3
	spellAPCost         = 3
	spellBaseDamage     = 10
	combatGridWidth     = 14
	combatGridHeight    = 18
	winExperience       = 50
	winGold             = 25
)

func NewCombatService() *CombatService {
	return &CombatService{
		combats: make(map[string]*Combat),
	}
}

func (s *CombatService) CreateCombat(fighters []*CombatFighter) string {
	combatID := uuid.New().String()
	width, height := combatGridWidth, combatGridHeight
	cells := generateCombatGrid(width, height)

	combat := &Combat{
		ID:         combatID,
		Fighters:   make(map[string]*CombatFighter),
		Cells:      cells,
		CellLookup: buildCellLookup(cells),
		Width:      width,
		Height:     height,
		Active:     true,
	}

	var turnOrder []string
	for _, f := range fighters {
		combat.Fighters[f.PlayerID] = f
		turnOrder = append(turnOrder, f.PlayerID)
	}
	combat.TurnOrder = turnOrder
	combat.CurrentTurn = 0

	s.mu.Lock()
	s.combats[combatID] = combat
	s.mu.Unlock()

	started := &pb.CombatEvent{
		Event: &pb.CombatEvent_CombatStarted{
			CombatStarted: &pb.CombatStarted{
				CombatId: combatID,
				Fighters: buildCombatFighters(combat),
				Cells:    cells,
			},
		},
	}
	for _, f := range fighters {
		if err := f.Stream.Send(started); err != nil {
			log.Printf("Failed to send combat start to %s: %v", f.PlayerID, err)
		}
	}

	s.startTurn(combat)
	return combatID
}

func (s *CombatService) JoinCombat(req *pb.JoinCombatRequest, stream pb.CombatService_JoinCombatServer) error {
	ctx := stream.Context()

	s.mu.RLock()
	combat, exists := s.combats[req.CombatId]
	s.mu.RUnlock()

	if !exists {
		return status.Error(codes.NotFound, "combat not found")
	}

	playerID := auth.GetUserID(ctx)
	if playerID == "" {
		return status.Error(codes.Unauthenticated, "not authenticated")
	}

	combat.mu.Lock()
	fighter, exists := combat.Fighters[playerID]
	if !exists {
		combat.mu.Unlock()
		return status.Error(codes.NotFound, "fighter not found in combat")
	}
	fighter.Stream = stream
	combat.mu.Unlock()

	<-ctx.Done()

	combat.mu.Lock()
	delete(combat.Fighters, playerID)
	combat.mu.Unlock()

	log.Printf("Player %s left combat %s", playerID, req.CombatId)
	return nil
}

func (s *CombatService) PerformAction(ctx context.Context, req *pb.CombatAction) (*pb.ActionResult, error) {
	playerID := auth.GetUserID(ctx)
	if playerID == "" {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	s.mu.RLock()
	combat, exists := s.combats[req.CombatId]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "combat not found")
	}

	combat.mu.Lock()

	if len(combat.TurnOrder) == 0 || combat.TurnOrder[combat.CurrentTurn] != playerID {
		combat.mu.Unlock()
		return nil, status.Error(codes.FailedPrecondition, "not your turn")
	}

	fighter := combat.Fighters[playerID]

	var result *pb.ActionResult
	var event *pb.CombatEvent
	var endEvent *pb.CombatEvent

	switch action := req.Action.(type) {
	case *pb.CombatAction_Move:
		result, event = s.handleMoveLocked(combat, fighter, action.Move)
	case *pb.CombatAction_Spell:
		result, event, endEvent = s.handleSpellLocked(combat, fighter, action.Spell)
	case *pb.CombatAction_Skip:
		result, event = s.handleSkipLocked(fighter)
	default:
		combat.mu.Unlock()
		return nil, status.Error(codes.InvalidArgument, "unknown action type")
	}

	combat.mu.Unlock()

	if event != nil {
		broadcastCombatEvent(combat, event)
	}
	if endEvent != nil {
		broadcastCombatEvent(combat, endEvent)
	}

	return result, nil
}

func (s *CombatService) EndTurn(ctx context.Context, req *pb.EndTurnRequest) (*pb.EndTurnResponse, error) {
	playerID := auth.GetUserID(ctx)

	s.mu.RLock()
	combat, exists := s.combats[req.CombatId]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "combat not found")
	}

	combat.mu.Lock()

	if len(combat.TurnOrder) == 0 || combat.TurnOrder[combat.CurrentTurn] != playerID {
		combat.mu.Unlock()
		return nil, status.Error(codes.FailedPrecondition, "not your turn")
	}

	turnEndedEvent := &pb.CombatEvent{
		Event: &pb.CombatEvent_TurnEnded{
			TurnEnded: &pb.TurnEnded{PlayerId: playerID},
		},
	}

	combat.CurrentTurn = (combat.CurrentTurn + 1) % len(combat.TurnOrder)

	nextPlayerID := combat.TurnOrder[combat.CurrentTurn]
	var turnStartedEvent *pb.CombatEvent
	if nextFighter, ok := combat.Fighters[nextPlayerID]; ok {
		nextFighter.AP = 6
		nextFighter.MP = 3
		turnStartedEvent = &pb.CombatEvent{
			Event: &pb.CombatEvent_TurnStarted{
				TurnStarted: &pb.TurnStarted{
					PlayerId:       nextPlayerID,
					ActionPoints:   int32(nextFighter.AP),
					MovementPoints: int32(nextFighter.MP),
				},
			},
		}
	}

	combat.mu.Unlock()

	broadcastCombatEvent(combat, turnEndedEvent)
	if turnStartedEvent != nil {
		broadcastCombatEvent(combat, turnStartedEvent)
	}

	return &pb.EndTurnResponse{
		Success:      true,
		NextPlayerId: nextPlayerID,
	}, nil
}

// handleMoveLocked must be called with combat.mu held.
func (s *CombatService) handleMoveLocked(combat *Combat, fighter *CombatFighter, move *pb.MoveAction) (*pb.ActionResult, *pb.CombatEvent) {
	dx := math.Abs(float64(int(move.TargetX) - fighter.X))
	dy := math.Abs(float64(int(move.TargetY) - fighter.Y))
	distance := int(dx + dy)

	if distance > fighter.MP {
		return &pb.ActionResult{Success: false, Error: "not enough movement points"}, nil
	}

	if !isCellWalkable(combat, int(move.TargetX), int(move.TargetY)) {
		return &pb.ActionResult{Success: false, Error: "cell not walkable"}, nil
	}

	oldX, oldY := fighter.X, fighter.Y
	fighter.X = int(move.TargetX)
	fighter.Y = int(move.TargetY)
	fighter.MP -= distance

	event := &pb.CombatEvent{
		Event: &pb.CombatEvent_PlayerAction{
			PlayerAction: &pb.PlayerAction{
				PlayerId: fighter.PlayerID,
				ActionResult: &pb.PlayerAction_Move{
					Move: &pb.MoveResult{
						FromX: int32(oldX), FromY: int32(oldY),
						ToX: move.TargetX, ToY: move.TargetY,
					},
				},
			},
		},
	}

	return &pb.ActionResult{
		Success:                 true,
		ActionPointsRemaining:   int32(fighter.AP),
		MovementPointsRemaining: int32(fighter.MP),
	}, event
}

// handleSpellLocked must be called with combat.mu held.
func (s *CombatService) handleSpellLocked(combat *Combat, fighter *CombatFighter, spell *pb.SpellAction) (*pb.ActionResult, *pb.CombatEvent, *pb.CombatEvent) {
	if fighter.AP < spellAPCost {
		return &pb.ActionResult{Success: false, Error: "not enough action points"}, nil, nil
	}

	fighter.AP -= spellAPCost

	var damages []*pb.DamageResult
	for _, f := range combat.Fighters {
		if f.PlayerID == fighter.PlayerID {
			continue
		}
		if f.X == int(spell.TargetX) && f.Y == int(spell.TargetY) {
			damage := spellBaseDamage
			f.Health -= damage
			if f.Health < 0 {
				f.Health = 0
			}
			damages = append(damages, &pb.DamageResult{
				TargetId:        f.PlayerID,
				Damage:          int32(damage),
				HealthRemaining: int32(f.Health),
			})
		}
	}

	event := &pb.CombatEvent{
		Event: &pb.CombatEvent_PlayerAction{
			PlayerAction: &pb.PlayerAction{
				PlayerId: fighter.PlayerID,
				ActionResult: &pb.PlayerAction_Spell{
					Spell: &pb.SpellResult{
						SpellId: spell.SpellId,
						CasterX: int32(fighter.X), CasterY: int32(fighter.Y),
						TargetX: spell.TargetX, TargetY: spell.TargetY,
						Damages: damages,
					},
				},
			},
		},
	}

	var endEvent *pb.CombatEvent
	if endEv, ended := s.checkCombatEndLocked(combat); ended {
		endEvent = endEv
	}

	return &pb.ActionResult{
		Success:                 true,
		ActionPointsRemaining:   int32(fighter.AP),
		MovementPointsRemaining: int32(fighter.MP),
	}, event, endEvent
}

// handleSkipLocked must be called with combat.mu held.
func (s *CombatService) handleSkipLocked(fighter *CombatFighter) (*pb.ActionResult, *pb.CombatEvent) {
	event := &pb.CombatEvent{
		Event: &pb.CombatEvent_PlayerAction{
			PlayerAction: &pb.PlayerAction{
				PlayerId: fighter.PlayerID,
				ActionResult: &pb.PlayerAction_Skip{
					Skip: &pb.SkipResult{},
				},
			},
		},
	}

	return &pb.ActionResult{
		Success:                 true,
		ActionPointsRemaining:   int32(fighter.AP),
		MovementPointsRemaining: int32(fighter.MP),
	}, event
}

func (s *CombatService) startTurn(combat *Combat) {
	if len(combat.TurnOrder) == 0 {
		return
	}

	visited := 0
	for visited < len(combat.TurnOrder) {
		playerID := combat.TurnOrder[combat.CurrentTurn]
		fighter, ok := combat.Fighters[playerID]
		if ok {
			fighter.AP = defaultCombatAP
			fighter.MP = defaultCombatMP
			broadcastCombatEvent(combat, &pb.CombatEvent{
				Event: &pb.CombatEvent_TurnStarted{
					TurnStarted: &pb.TurnStarted{
						PlayerId:       playerID,
						ActionPoints:   int32(fighter.AP),
						MovementPoints: int32(fighter.MP),
					},
				},
			})
			return
		}
		combat.CurrentTurn = (combat.CurrentTurn + 1) % len(combat.TurnOrder)
		visited++
	}
}

// checkCombatEndLocked must be called with combat.mu held.
func (s *CombatService) checkCombatEndLocked(combat *Combat) (*pb.CombatEvent, bool) {
	teamsAlive := make(map[int]bool)
	for _, f := range combat.Fighters {
		if f.Health > 0 {
			teamsAlive[f.Team] = true
		}
	}

	if len(teamsAlive) > 1 {
		return nil, false
	}

	var winningTeam int
	for t := range teamsAlive {
		winningTeam = t
	}

	var rewards []*pb.CombatReward
	for _, f := range combat.Fighters {
		if f.Team == winningTeam {
			rewards = append(rewards, &pb.CombatReward{
				PlayerId:   f.PlayerID,
				Experience: winExperience,
				Gold:       winGold,
			})
		}
	}

	event := &pb.CombatEvent{
		Event: &pb.CombatEvent_CombatEnded{
			CombatEnded: &pb.CombatEnded{
				WinningTeam: int32(winningTeam),
				Rewards:     rewards,
			},
		},
	}

	combat.Active = false
	s.mu.Lock()
	delete(s.combats, combat.ID)
	s.mu.Unlock()

	return event, true
}

func buildCellLookup(cells []*pb.CombatCell) map[[2]int]bool {
	lookup := make(map[[2]int]bool, len(cells))
	for _, c := range cells {
		lookup[[2]int{int(c.X), int(c.Y)}] = c.Walkable
	}
	return lookup
}

func generateCombatGrid(width, height int) []*pb.CombatCell {
	cells := make([]*pb.CombatCell, width*height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			walkable := true
			if (x == 4 || x == 9) && y > 3 && y < 14 {
				walkable = false
			}
			cells[y*width+x] = &pb.CombatCell{
				X:        int32(x),
				Y:        int32(y),
				Walkable: walkable,
			}
		}
	}
	return cells
}

func isCellWalkable(combat *Combat, x, y int) bool {
	return combat.CellLookup[[2]int{x, y}]
}

func buildCombatFighters(combat *Combat) []*pb.CombatFighter {
	var fighters []*pb.CombatFighter
	for _, f := range combat.Fighters {
		fighters = append(fighters, &pb.CombatFighter{
			PlayerId:       f.PlayerID,
			Username:       f.Username,
			Team:           int32(f.Team),
			X:              int32(f.X),
			Y:              int32(f.Y),
			Health:         int32(f.Health),
			MaxHealth:      int32(f.MaxHealth),
			ActionPoints:   int32(f.AP),
			MovementPoints: int32(f.MP),
		})
	}
	return fighters
}

// broadcastCombatEvent sends an event to all fighters.
// Caller must NOT hold combat.mu — this method acquires it.
func broadcastCombatEvent(combat *Combat, event *pb.CombatEvent) {
	combat.mu.Lock()
	type sendTarget struct {
		id     string
		stream pb.CombatService_JoinCombatServer
	}
	var targets []sendTarget
	for _, f := range combat.Fighters {
		if f.Stream != nil {
			targets = append(targets, sendTarget{id: f.PlayerID, stream: f.Stream})
		}
	}
	combat.mu.Unlock()

	for _, t := range targets {
		if err := t.stream.Send(event); err != nil {
			log.Printf("Failed to send combat event to %s: %v", t.id, err)
		}
	}
}

func (s *CombatService) GetCombat(id string) (*Combat, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.combats[id]
	return c, ok
}

func (s *CombatService) CreateCombatFromGame(playerData []struct {
	ID       string
	Username string
	X, Y     int
}) string {
	var fighters []*CombatFighter
	for i, p := range playerData {
		fighters = append(fighters, &CombatFighter{
			PlayerID:  p.ID,
			Username:  p.Username,
			Team:      i % 2,
			X:         2 + (i%2)*10,
			Y:         4 + i*2,
			Health:    100,
			MaxHealth: 100,
			AP:        defaultCombatAP,
			MP:        defaultCombatMP,
		})
	}
	return s.CreateCombat(fighters)
}

var _ pb.CombatServiceServer = (*CombatService)(nil)

func FormatCombatID(id string) string {
	return fmt.Sprintf("combat:%s", id)
}
