package network

import (
	"time"

	"fmt"

	"github.com/atvaark/dragons-dogma-server/modules/game"
)

//const MaxProperties = 64
const UsedProperties = 43

func GetDragonProperties(d *game.OnlineUrDragon) []Property {
	var props [UsedProperties]Property
	for i := 0; i < UsedProperties; i++ {
		props[i].Index = uint8(i)
	}

	props[0].Value2 = d.Generation

	const dragonHeatsIndex = 1
	const heartPropCount = game.UrDragonHeartCount / 2
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
	if !d.GraceTime.IsZero() {
		props[32].Value2 = uint32(d.GraceTime.Unix())
	}
	props[33].Value2 = d.KillCount

	const Unknown1 = 17825793
	props[35].Value1 = Unknown1
	props[37].Value1 = Unknown1
	props[39].Value1 = Unknown1

	const Unknown2 = 10800
	props[41].Value2 = Unknown2

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
	const dragonHeartsHealthIndex = 1
	const dragonHeartsHealthIndexEnd = dragonHeartsHealthIndex + heartPropCount
	const dragonHeartsHealthMaxIndex = dragonHeartsHealthIndexEnd
	const dragonHeartsHealthMaxIndexEnd = dragonHeartsHealthMaxIndex + heartPropCount

	for _, prop := range props {
		switch {
		case prop.Index == 0:
			d.Generation = prop.Value2
		case prop.Index >= dragonHeartsHealthIndex && prop.Index < dragonHeartsHealthIndexEnd:
			d.Hearts[(prop.Index-dragonHeartsHealthIndex)*2].Health = prop.Value1
			d.Hearts[(prop.Index-dragonHeartsHealthIndex)*2+1].Health = prop.Value2
		case prop.Index >= dragonHeartsHealthMaxIndex && prop.Index < dragonHeartsHealthMaxIndexEnd:
			d.Hearts[(prop.Index-dragonHeartsHealthMaxIndex)*2].MaxHealth = prop.Value1
			d.Hearts[(prop.Index-dragonHeartsHealthMaxIndex)*2+1].MaxHealth = prop.Value2
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
