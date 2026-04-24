package domain

type ActionType int

const (
	ActionAttack ActionType = iota
	ActionMove
	ActionEndTurn
)

type MobAction struct {
	Type    ActionType
	TargetX int
	TargetY int
}

type FighterView struct {
	ID       string
	Team     int
	Position Position
	Health   int
}

type CombatView struct {
	FighterPos Position
	FighterAP  int
	FighterMP  int
	IsWalkable func(x, y int) bool
	Fighters   []FighterView
}

type MobBehavior interface {
	Decide(view CombatView) MobAction
}

type TofuBehavior struct{}

const tofuSpellRange = 4

func (b *TofuBehavior) Decide(view CombatView) MobAction {
	target := b.findNearestEnemy(view)
	if target == nil {
		return MobAction{Type: ActionEndTurn}
	}

	dist := manhattan(view.FighterPos, target.Position)

	if dist <= tofuSpellRange && view.FighterAP >= 3 {
		return MobAction{Type: ActionAttack, TargetX: int(target.Position.X), TargetY: int(target.Position.Y)}
	}

	if view.FighterMP > 0 {
		nx, ny := stepToward(view.FighterPos, target.Position, view.IsWalkable)
		if nx != int(view.FighterPos.X) || ny != int(view.FighterPos.Y) {
			return MobAction{Type: ActionMove, TargetX: nx, TargetY: ny}
		}
	}

	return MobAction{Type: ActionEndTurn}
}

func (b *TofuBehavior) findNearestEnemy(view CombatView) *FighterView {
	var best *FighterView
	bestDist := -1
	for i := range view.Fighters {
		f := &view.Fighters[i]
		if f.Team == 1 || f.Health <= 0 {
			continue
		}
		d := manhattan(view.FighterPos, f.Position)
		if bestDist < 0 || d < bestDist {
			best = f
			bestDist = d
		}
	}
	return best
}

func stepToward(from Position, to Position, isWalkable func(x, y int) bool) (int, int) {
	cardinals := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	bestX, bestY := int(from.X), int(from.Y)
	bestDist := manhattan(from, to)

	for _, d := range cardinals {
		nx, ny := int(from.X)+d[0], int(from.Y)+d[1]
		if isWalkable == nil || !isWalkable(nx, ny) {
			continue
		}
		dist := manhattan(Position{X: int32(nx), Y: int32(ny)}, to)
		if dist < bestDist {
			bestDist = dist
			bestX, bestY = nx, ny
		}
	}
	return bestX, bestY
}

func manhattan(a, b Position) int {
	dx := int(a.X - b.X)
	dy := int(a.Y - b.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
