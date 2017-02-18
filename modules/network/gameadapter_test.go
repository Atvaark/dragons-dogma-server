package network

import (
	"testing"

	"github.com/atvaark/dragons-dogma-server/modules/game"
)

func TestNetworkToDragonProperties(t *testing.T) {
	props := []Property{
		{1, 11, 111},
		{2, 22, 222},
	}

	dragonProps := networkToDragonProperties(props)

	if len(dragonProps) != len(props) {
		t.Errorf("dragon property length mismatch: got %d expected %d", len(dragonProps), len(props))
		return
	}

	for i := 0; i < len(dragonProps); i++ {
		dp := dragonProps[i]
		np := props[i]

		if dp.Index != np.Index {
			t.Errorf("dragon property index mismatch at pos %d: got %d expected %d", i, dp.Index, np.Index)
		}

		if dp.Value1 != np.Value1 {
			t.Errorf("dragon property value 1 mismatch at pos %d: got %d expected %d", i, dp.Value1, np.Value1)
		}

		if dp.Value2 != np.Value2 {
			t.Errorf("dragon property value 2 mismatch at pos %d: got %d expected %d", i, dp.Value2, np.Value2)
		}
	}
}

func TestDragonToNetworkProperties(t *testing.T) {
	dragonProps := []game.DragonProperty{
		{1, 11, 111},
		{2, 22, 222},
	}

	props := dragonToNetworkProperties(dragonProps)

	if len(props) != len(dragonProps) {
		t.Errorf("network property length mismatch: got %d expected %d", len(props), len(dragonProps))
		return
	}

	for i := 0; i < len(props); i++ {
		np := props[i]
		dp := dragonProps[i]

		if np.Index != dp.Index {
			t.Errorf("network property index mismatch at pos %d: got %d expected %d", i, np.Index, dp.Index)
		}

		if np.Value1 != dp.Value1 {
			t.Errorf("network property value 1 mismatch at pos %d: got %d expected %d", i, np.Value1, dp.Value1)
		}

		if np.Value2 != dp.Value2 {
			t.Errorf("network property value 2 mismatch at pos %d: got %d expected %d", i, np.Value2, dp.Value2)
		}
	}
}
