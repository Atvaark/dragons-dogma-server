package game

const PawnRewardsMax = 100

type PawnRewards struct {
	PawnUserID uint64
	Revision   int32
	Rewards    []*PawnReward
}

const PawnRewardItemRefsMax = 10

type PawnReward struct {
	ItemsRefs []ItemRef
	UserID    uint64
}

type ItemRef int32

type PawnRewardsDatabase interface {
	GetPawnRewards(userID uint64) (*PawnRewards, error)
	PutPawnRewards(*PawnRewards) error
}
