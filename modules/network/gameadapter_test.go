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

func TestUserAreaToPawnRewards(t *testing.T) {
	userID := uint64(0xFFFFFFFFFFFF)
	area := NewUserArea()
	area.Revision = 20
	s10 := &area.Slots[10]
	s10.IsFree = 0
	s10.Items[0] = 0
	s10.Items[1] = 1
	s10.ItemsCount = 2
	s10.User = uint64(0xEEEEEEEEEEEE)
	s11 := &area.Slots[11]
	s11.IsFree = 0
	s11.Items[0] = 2
	s11.Items[1] = 3
	s11.Items[2] = 4
	s11.ItemsCount = 3
	s11.User = uint64(0xDDDDDDDDDDDD)

	rewards := userAreaToPawnRewards(userID, area)

	if rewards.PawnUserID != userID {
		t.Errorf("pawn UserID mismatch: got %d expected %d", rewards.PawnUserID, userID)
	}

	if rewards.Revision != int32(area.Revision) {
		t.Errorf("revision mismatch: got %d expected %d", rewards.Revision, area.Revision)
	}

	for i, reward := range rewards.Rewards {
		var expected *game.PawnReward
		switch i {
		case 10:
			expected = &game.PawnReward{
				UserID: uint64(0xEEEEEEEEEEEE),
				ItemsRefs: []game.ItemRef{
					game.ItemRef(0),
					game.ItemRef(1),
				},
			}
		case 11:
			expected = &game.PawnReward{
				UserID: uint64(0xDDDDDDDDDDDD),
				ItemsRefs: []game.ItemRef{
					game.ItemRef(2),
					game.ItemRef(3),
					game.ItemRef(4),
				},
			}
		default:
			expected = nil
		}

		switch {
		case reward == nil && expected == nil:
		// ok
		case reward != nil && expected == nil:
			t.Errorf("reward %d mismatch: expected nil", i)
		case reward == nil && expected != nil:
			t.Errorf("reward %d mismatch: expected not nil", i)
		default:
			if reward.UserID != expected.UserID {
				t.Errorf("reward %d mismatch: got UserID %d expected %d", i, reward.UserID, expected.UserID)
			}

			if len(reward.ItemsRefs) != len(expected.ItemsRefs) {
				t.Errorf("reward %d item ref mismatch: got %d expected %d ", i, len(reward.ItemsRefs), len(expected.ItemsRefs))
			}

			for j := 0; j < len(reward.ItemsRefs); j++ {
				item := reward.ItemsRefs[j]
				itemExpected := expected.ItemsRefs[j]
				if item != itemExpected {
					t.Errorf("reward %d mismatch: got item ref %d expected %d", i, item, itemExpected)
				}
			}
		}
	}
}

func TestPawnRewardsToUserArea(t *testing.T) {
	rewards := game.PawnRewards{}
	rewards.Rewards = make([]*game.PawnReward, game.PawnRewardsMax)
	rewards.PawnUserID = uint64(0x123456789876)
	rewards.Revision = 10
	r15 := &game.PawnReward{}
	r15.UserID = uint64(0xEEEEEEEEEEEE)
	r15.ItemsRefs = make([]game.ItemRef, 1)
	r15.ItemsRefs[0] = game.ItemRef(5)
	rewards.Rewards[15] = r15
	r16 := &game.PawnReward{}
	r16.UserID = uint64(0xDDDDDDDDDDDD)
	r16.ItemsRefs = make([]game.ItemRef, 3)
	r16.ItemsRefs[0] = game.ItemRef(6)
	r16.ItemsRefs[1] = game.ItemRef(7)
	r16.ItemsRefs[2] = game.ItemRef(8)
	rewards.Rewards[16] = r16

	area := pawnRewardsToUserArea(&rewards)

	if area.Revision != uint32(rewards.Revision) {
		t.Errorf("revision mismatch: got %d expected %d", area.Revision, rewards.Revision)
	}

	for i, slot := range area.Slots {
		expected := UserAreaSlot{
			IsFree: 1,
		}
		for j := 0; j < len(expected.Items); j++ {
			expected.Items[j] = NoUserAreaItem
		}
		switch i {
		case 15:
			expected.IsFree = 0
			expected.User = uint64(0xEEEEEEEEEEEE)
			expected.Items[0] = UserAreaItem(5)
		case 16:
			expected.IsFree = 0
			expected.User = uint64(0xDDDDDDDDDDDD)
			expected.Items[0] = UserAreaItem(6)
			expected.Items[1] = UserAreaItem(7)
			expected.Items[2] = UserAreaItem(8)
		}

		if slot.User != expected.User {
			t.Errorf("reward %d mismatch: got User %d expected %d", i, slot.User, expected.User)
		}

		for j := 0; j < len(slot.Items); j++ {
			item := slot.Items[j]
			itemExpected := expected.Items[j]
			if item != itemExpected {
				t.Errorf("reward %d mismatch: got item ref %d expected %d", i, item, itemExpected)
			}
		}
	}
}
