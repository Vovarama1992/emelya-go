package money_ports

import (
	"context"

	model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
)

type RewardService interface {
	Create(ctx context.Context, reward *model.Reward) error
	GetByID(ctx context.Context, id int64) (*model.Reward, error)
	UpdateWithdrawn(ctx context.Context, rewardID int64, delta float64) error
	FindByUserID(ctx context.Context, userID int64) ([]*model.Reward, error)
	AccrueDailyRewardForDeposit(ctx context.Context, depositID int64, dailyReward float64) error
}
