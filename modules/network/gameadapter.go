package network

import (
	"github.com/atvaark/dragons-dogma-server/modules/game"
)

func networkToDragonProperties(networkProps []Property) []game.DragonProperty {
	props := make([]game.DragonProperty, len(networkProps))
	for i := 0; i < len(networkProps); i++ {
		props[i] = game.DragonProperty(networkProps[i])
	}
	return props
}

func dragonToNetworkProperties(dragonProps []game.DragonProperty) []Property {
	props := make([]Property, len(dragonProps))
	for i := 0; i < len(dragonProps); i++ {
		props[i] = Property(dragonProps[i])
	}
	return props
}
