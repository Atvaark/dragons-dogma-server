package game

import (
	"fmt"
	"math"
	"time"
)

const UrDragonHeartCount = 30
const UrDragonHeartHealth = 10000000
const UserIdCount = 3
const ArmorMax = 100000

//const GraceTime = 40 * time.Minute

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

func (d *OnlineUrDragon) NextGeneration() *OnlineUrDragon {
	var next OnlineUrDragon

	next.Generation = d.Generation + 1

	spawnTime := time.Now().UTC()
	next.SpawnTime = &spawnTime

	defense := d.Defense
	if defense < ArmorMax {
		// reach max defense in 100 generations
		defense = 900*next.Generation + uint32(math.Pow(float64(next.Generation), 2))
	}
	next.Defense = defense

	for i := 0; i < len(d.Hearts); i++ {
		next.Hearts[i].Health = UrDragonHeartHealth
		next.Hearts[i].MaxHealth = UrDragonHeartHealth
	}

	for i := 0; i < len(d.PawnUserIDs); i++ {
		next.PawnUserIDs[i] = d.PawnUserIDs[i]
	}

	return &next
}

type Database interface {
	GetOnlineUrDragon() (*OnlineUrDragon, error)
	PutOnlineUrDragon(*OnlineUrDragon) error
}

type DragonProperty struct {
	Index  uint8
	Value1 uint32
	Value2 uint32
}

const MaxDragonProperties = 64
const UsedDragonProperties = 43

func AllDragonPropertyIndices() []byte {
	var indices [UsedDragonProperties]byte
	for i := 0; i < len(indices); i++ {
		indices[i] = byte(i)
	}
	return indices[:]
}

func (d *OnlineUrDragon) Properties() []DragonProperty {
	var props [UsedDragonProperties]DragonProperty
	for i := 0; i < UsedDragonProperties; i++ {
		props[i].Index = uint8(i)
	}

	props[0].Value2 = d.Generation

	const dragonHeartsHealthIndexStart = 1
	const heartPropCount = UrDragonHeartCount / 2
	heartHealthIndex := dragonHeartsHealthIndexStart
	heartHealthMaxIndex := dragonHeartsHealthIndexStart + heartPropCount
	for i := 0; i < len(d.Hearts); i++ {
		heart := &d.Hearts[i]

		healthProp := &props[heartHealthIndex]
		healthMaxProp := &props[heartHealthMaxIndex]

		if i%2 == 0 {
			healthProp.Value1 = heart.Health
			healthMaxProp.Value1 = heart.MaxHealth
		} else {
			healthProp.Value2 = heart.Health
			healthMaxProp.Value2 = heart.MaxHealth

			heartHealthIndex++
			heartHealthMaxIndex++
		}
	}

	props[31].Value2 = d.FightCount
	if d.KillTime != nil {
		props[32].Value2 = uint32(d.KillTime.Unix())
	}
	props[33].Value2 = d.KillCount
	// props[34] is not used

	const userIdsStartIndex = 35
	for i := 0; i < len(d.PawnUserIDs); i++ {
		userId := d.PawnUserIDs[i]
		userIdIdx := userIdsStartIndex + i*2
		props[userIdIdx+0].Value1 = uint32(userId >> 32)
		props[userIdIdx+0].Value2 = uint32(userId)
		// props[userIdIdx+1] is not used
	}

	props[41].Value2 = d.Defense

	if d.SpawnTime != nil {
		props[42].Value2 = uint32(d.SpawnTime.Unix())
	}

	return props[:]
}

func (d *OnlineUrDragon) PropertiesFiltered(indexFilter []byte) ([]DragonProperty, error) {
	props := d.Properties()
	if len(props) == 0 && len(indexFilter) == 0 {
		return props, nil
	}

	maxPropIdx := byte(len(props) - 1)
	filteredProps := make([]DragonProperty, len(indexFilter))

	for filterIdx, propIdx := range indexFilter {
		if propIdx > maxPropIdx {
			return nil, fmt.Errorf("invalid property index %d", propIdx)
		}

		filteredProps[filterIdx] = props[propIdx]
	}

	return filteredProps, nil
}

func (d *OnlineUrDragon) SetProperties(props []DragonProperty) error {
	const heartPropCount = UrDragonHeartCount / 2
	const dragonHeartsHealthIndexStart = 1
	const dragonHeartsHealthIndexEnd = dragonHeartsHealthIndexStart + heartPropCount
	const dragonHeartsHealthMaxIndexStart = dragonHeartsHealthIndexEnd
	const dragonHeartsHealthMaxIndexEnd = dragonHeartsHealthMaxIndexStart + heartPropCount
	const userIdsIndexStart = 35
	const userIdPropCount = UserIdCount * 2
	const userIdsIndexEnd = userIdsIndexStart + userIdPropCount

	for _, prop := range props {
		switch {
		case prop.Index == 0:
			d.Generation = prop.Value2
		case prop.Index >= dragonHeartsHealthIndexStart && prop.Index < dragonHeartsHealthIndexEnd:
			d.Hearts[(prop.Index-dragonHeartsHealthIndexStart)*2].Health = prop.Value1
			d.Hearts[(prop.Index-dragonHeartsHealthIndexStart)*2+1].Health = prop.Value2
		case prop.Index >= dragonHeartsHealthMaxIndexStart && prop.Index < dragonHeartsHealthMaxIndexEnd:
			d.Hearts[(prop.Index-dragonHeartsHealthMaxIndexStart)*2].MaxHealth = prop.Value1
			d.Hearts[(prop.Index-dragonHeartsHealthMaxIndexStart)*2+1].MaxHealth = prop.Value2
		case prop.Index == 31:
			d.FightCount = prop.Value2
		case prop.Index == 32:
			d.KillTime = nillableUnixTime(prop.Value2)
		case prop.Index == 33:
			d.KillCount = prop.Value2
		case prop.Index >= userIdsIndexStart && prop.Index < userIdsIndexEnd && (prop.Index-userIdsIndexStart)%2 == 0:
			d.PawnUserIDs[(prop.Index-userIdsIndexStart)/2] = uint64(prop.Value1)<<32 | uint64(prop.Value2)
		case prop.Index == 41:
			d.Defense = prop.Value2
		case prop.Index == 42:
			d.SpawnTime = nillableUnixTime(prop.Value2)
		case prop.Index > MaxDragonProperties:
			return fmt.Errorf("invalid property index %d", prop.Index)
		}
	}

	// TODO: Test if the client sends the kill date or the kill date is set once all hearts have been killed
	if d.KillTime == nil {
		var isAlive bool
		for _, h := range d.Hearts {
			if h.Health > 0 {
				isAlive = true
				break
			}
		}

		if !isAlive {
			killTime := time.Now().UTC()
			d.KillTime = &killTime
		}
	}

	return nil
}

func (d *OnlineUrDragon) AddProperties(props []DragonProperty) ([]DragonProperty, error) {
	indices := make([]byte, len(props))

	for i, prop := range props {
		switch {
		case prop.Index == 31:
			d.FightCount += prop.Value2
		case prop.Index == 33:
			d.KillCount += prop.Value2
		default:
			// TODO: Test if the client increments values besides the counters.
		}

		indices[i] = prop.Index
	}

	return d.PropertiesFiltered(indices)
}

func nillableUnixTime(i uint32) *time.Time {
	if i == 0 {
		return nil
	}

	t := time.Unix(int64(i), 0)
	return &t
}
