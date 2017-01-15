package game

import (
	"testing"
	"time"
)

func TestOnlineUrDragonPropertyRoundtrip(t *testing.T) {
	dragon := NewOnlineUrDragon()
	dragon.Generation = 5
	dragon.FightCount = 1234
	dragon.SpawnTime, _ = time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	dragon.GraceTime, _ = time.Parse("2006-01-02T15:04:05Z", "2006-01-02T16:05:06Z")
	dragon.KillCount = 123
	for i := 0; i < len(dragon.Hearts); i++ {
		dragon.Hearts[i].Health = dragon.Hearts[i].Health - uint32(i)
		dragon.Hearts[i].MaxHealth = dragon.Hearts[i].MaxHealth + uint32(i)
	}

	props := dragon.NetworkProperties()

	parsedDragon := NewOnlineUrDragon()
	parsedDragon.SetNetworkProperties(props)

	if parsedDragon.Generation != dragon.Generation {
		t.Errorf("Generation mismatch %d %d", parsedDragon.Generation, dragon.Generation)
	}

	if parsedDragon.FightCount != dragon.FightCount {
		t.Errorf("FightCount mismatch %d %d", parsedDragon.FightCount, dragon.FightCount)
	}

	if !parsedDragon.SpawnTime.Equal(dragon.SpawnTime) {
		t.Errorf("SpawnTime mismatch %v %v", parsedDragon.SpawnTime, dragon.SpawnTime)
	}

	if !parsedDragon.GraceTime.Equal(dragon.GraceTime) {
		t.Errorf("GraceTime mismatch %v %v", parsedDragon.GraceTime, dragon.GraceTime)
	}

	if parsedDragon.KillCount != dragon.KillCount {
		t.Errorf("KillCount mismatch %d %d", parsedDragon.KillCount, dragon.KillCount)
	}

	if len(parsedDragon.Hearts) != len(dragon.Hearts) {
		t.Errorf("Heart count mismatch %d %d", len(parsedDragon.Hearts), len(dragon.Hearts))
	}

	for i := 0; i < len(parsedDragon.Hearts); i++ {
		if parsedDragon.Hearts[i].Health != dragon.Hearts[i].Health {
			t.Errorf("heart Health mismatch %d %d %d", i, parsedDragon.Hearts[i].Health, dragon.Hearts[i].Health)
		}

		if parsedDragon.Hearts[i].MaxHealth != dragon.Hearts[i].MaxHealth {
			t.Errorf("heart MaxHealth mismatch %d %d %d", i, parsedDragon.Hearts[i].MaxHealth, dragon.Hearts[i].MaxHealth)
		}
	}
}
