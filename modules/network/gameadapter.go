package network

import (
	"fmt"
	"time"

	"github.com/atvaark/dragons-dogma-server/modules/game"
)

//const MaxProperties = 64
const UsedProperties = 43

func AllPropertyIndices() []byte {
	var indices [UsedProperties]byte
	for i := 0; i < len(indices); i++ {
		indices[i] = byte(i)
	}
	return indices[:]
}

func GetDragonProperties(d *game.OnlineUrDragon) []Property {
	var props [UsedProperties]Property
	for i := 0; i < UsedProperties; i++ {
		props[i].Index = uint8(i)
	}

	props[0].Value2 = d.Generation

	const dragonHeartsHealthIndexStart = 1
	const heartPropCount = game.UrDragonHeartCount / 2
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
	if !d.KillTime.IsZero() {
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

	if !d.SpawnTime.IsZero() {
		props[42].Value2 = uint32(d.SpawnTime.Unix())
	}

	return props[:]
}

func GetDragonPropertiesFilter(d *game.OnlineUrDragon, indexFilter []byte) ([]Property, error) {
	props := GetDragonProperties(d)
	if len(props) == 0 && len(indexFilter) == 0 {
		return props, nil
	}

	maxPropIdx := byte(len(props) - 1)
	filteredProps := make([]Property, len(indexFilter))

	for filterIdx, propIdx := range indexFilter {
		if propIdx > maxPropIdx {
			return nil, fmt.Errorf("invalid property index %d", propIdx)
		}

		filteredProps[filterIdx] = props[propIdx]
	}

	return filteredProps, nil
}

func SetDragonProperties(d *game.OnlineUrDragon, props []Property) {
	const heartPropCount = game.UrDragonHeartCount / 2
	const dragonHeartsHealthIndexStart = 1
	const dragonHeartsHealthIndexEnd = dragonHeartsHealthIndexStart + heartPropCount
	const dragonHeartsHealthMaxIndexStart = dragonHeartsHealthIndexEnd
	const dragonHeartsHealthMaxIndexEnd = dragonHeartsHealthMaxIndexStart + heartPropCount
	const userIdsIndexStart = 35
	const userIdPropCount = game.UserIdCount * 2
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
		}
	}
}

func nillableUnixTime(i uint32) *time.Time {
	if i == 0 {
		return nil
	}

	t := time.Unix(int64(i), 0)
	return &t
}
