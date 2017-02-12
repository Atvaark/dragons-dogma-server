package network

import (
	"testing"
	"time"

	"github.com/atvaark/dragons-dogma-server/modules/game"
)

func parseNillableTime(t *testing.T, layout, value string) *time.Time {
	v, err := time.Parse(layout, value)
	if err != nil {
		t.Error(err)
	}

	return &v
}

func nillableTimeEquals(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}

	if t1 == nil || t2 == nil {
		return false
	}

	return t1.Equal(*t2)
}

func TestOnlineUrDragonPropertyRoundtrip(t *testing.T) {
	dragon := game.NewOnlineUrDragon()
	dragon.Generation = 5
	dragon.SpawnTime = parseNillableTime(t, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	dragon.Defense = 100000
	dragon.FightCount = 1234
	dragon.KillTime = parseNillableTime(t, "2006-01-02T15:04:05Z", "2006-01-02T16:05:06Z")
	dragon.KillCount = 123
	for i := 0; i < len(dragon.Hearts); i++ {
		dragon.Hearts[i].Health = dragon.Hearts[i].Health - uint32(i)
		dragon.Hearts[i].MaxHealth = dragon.Hearts[i].MaxHealth + uint32(i)
	}
	for i := 0; i < len(dragon.PawnUserIDs); i++ {
		dragon.PawnUserIDs[i] = uint64(i+1)<<32 | uint64(i+1)
	}

	props := GetDragonProperties(dragon)

	parsedDragon := game.NewOnlineUrDragon()
	SetDragonProperties(parsedDragon, props)

	if parsedDragon.Generation != dragon.Generation {
		t.Errorf("Generation mismatch %d %d", parsedDragon.Generation, dragon.Generation)
	}

	if parsedDragon.Defense != dragon.Defense {
		t.Errorf("Defense mismatch %d %d", parsedDragon.Defense, dragon.Defense)
	}

	if !nillableTimeEquals(parsedDragon.SpawnTime, dragon.SpawnTime) {
		t.Errorf("SpawnTime mismatch %v %v", parsedDragon.SpawnTime, dragon.SpawnTime)
	}

	if !nillableTimeEquals(parsedDragon.KillTime, dragon.KillTime) {
		t.Errorf("KillTime mismatch %v %v", parsedDragon.KillTime, dragon.KillTime)
	}

	if parsedDragon.FightCount != dragon.FightCount {
		t.Errorf("FightCount mismatch %d %d", parsedDragon.FightCount, dragon.FightCount)
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

	if len(parsedDragon.PawnUserIDs) != len(dragon.PawnUserIDs) {
		t.Errorf("UserId count mismatch %d %d", len(parsedDragon.PawnUserIDs), len(dragon.PawnUserIDs))
	}

	for i := 0; i < len(parsedDragon.PawnUserIDs); i++ {
		if parsedDragon.PawnUserIDs[i] != dragon.PawnUserIDs[i] {
			t.Errorf("UserId mismatch %d %d %d", i, parsedDragon.PawnUserIDs[i], dragon.PawnUserIDs[i])
		}
	}
}
