package game

import (
	"math"
	"time"
)

const UrDragonHeartCount = 30
const UrDragonHeartHealth = 10000000
const UserIdCount = 3
const ArmorMax = 100000

type OnlineUrDragon struct {
	Generation  uint32
	SpawnTime   *time.Time
	Defense     uint32
	FightCount  uint32
	KillTime    *time.Time
	KillCount   uint32
	Hearts      [UrDragonHeartCount]UrDragonHeart
	PawnUserIDs [UserIdCount]uint64
}

type UrDragonHeart struct {
	Health    uint32
	MaxHealth uint32
}

func NewOnlineUrDragon() *OnlineUrDragon {
	var dragon OnlineUrDragon
	return dragon.NextGeneration()
}

func (dragon *OnlineUrDragon) NextGeneration() *OnlineUrDragon {
	var next OnlineUrDragon

	next.Generation = dragon.Generation + 1

	spawnTime := time.Now().UTC()
	next.SpawnTime = &spawnTime

	defense := dragon.Defense
	if defense < ArmorMax {
		// reach max defense in 100 generations
		defense = 900*next.Generation + uint32(math.Pow(float64(next.Generation), 2))
	}
	next.Defense = defense

	for i := 0; i < len(dragon.Hearts); i++ {
		next.Hearts[i].Health = UrDragonHeartHealth
		next.Hearts[i].MaxHealth = UrDragonHeartHealth
	}

	for i := 0; i < len(dragon.PawnUserIDs); i++ {
		next.PawnUserIDs[i] = dragon.PawnUserIDs[i]
	}

	return &next
}

type Database interface {
	GetOnlineUrDragon() (*OnlineUrDragon, error)
	PutOnlineUrDragon(*OnlineUrDragon) error
}
