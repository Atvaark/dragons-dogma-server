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

func userAreaToPawnRewards(userID uint64, area *UserArea) *game.PawnRewards {
	if area == nil {
		return nil
	}

	var r game.PawnRewards
	r.PawnUserID = userID
	r.Rewards = make([]*game.PawnReward, game.PawnRewardsMax)
	r.Revision = int32(area.Revision)

	for i, slot := range area.Slots {
		var reward *game.PawnReward
		if slot.IsFree == 1 {
			reward = nil
		} else {
			reward = &game.PawnReward{}
			reward.UserID = slot.User
			reward.ItemsRefs = make([]game.ItemRef, 0)
			for j := 0; j < int(slot.ItemsCount) && j < game.PawnRewardItemRefsMax; j++ {
				reward.ItemsRefs = append(reward.ItemsRefs, game.ItemRef(slot.Items[j]))
			}
		}
		r.Rewards[i] = reward
	}

	return &r
}

func pawnRewardsToUserArea(rewards *game.PawnRewards) *UserArea {
	if rewards == nil {
		return nil
	}

	var a UserArea
	a.Revision = uint32(rewards.Revision)

	for i, r := range rewards.Rewards {
		slot := &a.Slots[i]
		slot.IsFree = 1
		for j := 0; j < len(slot.Items); j++ {
			slot.Items[j] = NoUserAreaItem
		}

		if r != nil {
			slot.IsFree = 0
			slot.User = r.UserID

			for j, ref := range r.ItemsRefs {
				slot.Items[j] = UserAreaItem(ref)
			}
			slot.ItemsCount = uint32(len(r.ItemsRefs))
		}
	}

	return &a
}
