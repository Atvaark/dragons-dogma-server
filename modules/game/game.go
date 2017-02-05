package game

import (
	"time"
)

const UrDragonHeartCount = 30
const UrDragonHeartHealth = 10000000
const UserIdCount = 3

type OnlineUrDragon struct {
	Generation uint32
	FightCount uint32
	SpawnTime  time.Time
	GraceTime  time.Time
	KillCount  uint32
	Hearts     [UrDragonHeartCount]UrDragonHeart
	UserIds    [UserIdCount]uint64
}

type UrDragonHeart struct {
	Health    uint32
	MaxHealth uint32
}

func NewOnlineUrDragon() *OnlineUrDragon {
	var dragon OnlineUrDragon
	dragon.SpawnTime = time.Now().UTC()
	for i := 0; i < len(dragon.Hearts); i++ {
		dragon.Hearts[i].Health = UrDragonHeartHealth
		dragon.Hearts[i].MaxHealth = UrDragonHeartHealth
	}

	return &dragon
}
