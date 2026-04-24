package domain

type MobType string

const (
	MobTypeTofu MobType = "tofu"
)

type Position struct {
	X, Y int32
}

type Stats struct {
	Health    int
	MaxHealth int
	AP        int
	MP        int
}

type Mob struct {
	ID       string
	Type     MobType
	Name     string
	Position Position
	Stats    Stats
	Team     int
}
