package game

import (
	"time"

	"github.com/atvaark/dragons-dogma-server/modules/network"
)

const UrDragonHeartCount = 30
const UrDragonHeartHealth = 10000000

//const MaxProperties = 64
const UsedProperties = 43

type OnlineUrDragon struct {
	Generation uint32
	FightCount uint32
	SpawnTime  time.Time
	GraceTime  time.Time
	KillCount  uint32
	Hearts     [UrDragonHeartCount]UrDragonHeart
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

func (d *OnlineUrDragon) NetworkProperties() []network.Property {
	var props [UsedProperties]network.Property
	for i := 0; i < UsedProperties; i++ {
		props[i].Index = uint8(i)
	}

	props[0].Value2 = d.Generation

	const dragonHeatsIndex = 1
	const heartPropCount = UrDragonHeartCount / 2
	heartHealthIdx := dragonHeatsIndex
	heartHealthMaxIdx := dragonHeatsIndex + heartPropCount
	for i := 0; i < len(d.Hearts); i++ {
		heart := &d.Hearts[i]

		healthProp := &props[heartHealthIdx]
		healthMaxProp := &props[heartHealthMaxIdx]

		if i%2 == 0 {
			healthProp.Value1 = heart.Health
			healthMaxProp.Value1 = heart.MaxHealth
		} else {
			healthProp.Value2 = heart.Health
			healthMaxProp.Value2 = heart.MaxHealth

			heartHealthIdx++
			heartHealthMaxIdx++
		}
	}

	props[31].Value2 = d.FightCount
	props[32].Value2 = uint32(d.GraceTime.Unix())
	props[33].Value2 = d.KillCount

	const Unknown1 = 17825793
	props[35].Value1 = Unknown1
	props[37].Value1 = Unknown1
	props[39].Value1 = Unknown1

	const Unknown2 = 10800
	props[41].Value2 = Unknown2

	props[42].Value2 = uint32(d.SpawnTime.Unix())

	return props[:]
}

func (d *OnlineUrDragon) SetNetworkProperties(props []network.Property) {
	const heartPropCount = UrDragonHeartCount / 2
	const dragonHeatsHealthIndex = 1
	const dragonHeatsHealthIndexEnd = dragonHeatsHealthIndex + heartPropCount
	const dragonHeatsHealthMaxIndex = dragonHeatsHealthIndexEnd
	const dragonHeatsHealthMaxIndexEnd = dragonHeatsHealthMaxIndex + heartPropCount

	for _, prop := range props {
		switch {
		case prop.Index == 0:
			d.Generation = prop.Value2
		case prop.Index >= dragonHeatsHealthIndex && prop.Index < dragonHeatsHealthIndexEnd:
			d.Hearts[(prop.Index-dragonHeatsHealthIndex)*2].Health = prop.Value1
			d.Hearts[(prop.Index-dragonHeatsHealthIndex)*2+1].Health = prop.Value2
		case prop.Index >= dragonHeatsHealthMaxIndex && prop.Index < dragonHeatsHealthMaxIndexEnd:
			d.Hearts[(prop.Index-dragonHeatsHealthMaxIndex)*2].MaxHealth = prop.Value1
			d.Hearts[(prop.Index-dragonHeatsHealthMaxIndex)*2+1].MaxHealth = prop.Value2
		case prop.Index == 31:
			d.FightCount = prop.Value2
		case prop.Index == 32:
			d.GraceTime = time.Unix(int64(prop.Value2), 0)
		case prop.Index == 33:
			d.KillCount = prop.Value2
		case prop.Index == 42:
			d.SpawnTime = time.Unix(int64(prop.Value2), 0)
		}
	}
}
