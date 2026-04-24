package domain

import "github.com/google/uuid"

type MobFactory struct{}

func NewMobFactory() *MobFactory {
	return &MobFactory{}
}

func (f *MobFactory) Create(mobType MobType, pos Position) *Mob {
	id := string(mobType) + "-" + uuid.New().String()[:8]

	stats := f.defaultStats(mobType)
	name := f.defaultName(mobType)

	return &Mob{
		ID:       id,
		Type:     mobType,
		Name:     name,
		Position: pos,
		Stats:    stats,
		Team:     1,
	}
}

func (f *MobFactory) CreateBehavior(mobType MobType) MobBehavior {
	switch mobType {
	case MobTypeTofu:
		return &TofuBehavior{}
	default:
		return &TofuBehavior{}
	}
}

func (f *MobFactory) defaultStats(mobType MobType) Stats {
	switch mobType {
	case MobTypeTofu:
		return Stats{Health: 50, MaxHealth: 50, AP: 6, MP: 3}
	default:
		return Stats{Health: 50, MaxHealth: 50, AP: 6, MP: 3}
	}
}

func (f *MobFactory) defaultName(mobType MobType) string {
	switch mobType {
	case MobTypeTofu:
		return "Tofu"
	default:
		return string(mobType)
	}
}
