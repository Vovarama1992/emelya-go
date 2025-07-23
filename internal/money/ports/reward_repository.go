package money_ports

import (
	"context"
	"time"

	model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
)

type RewardRepository interface {
	Create(ctx context.Context, reward *model.Reward) error
	UpdateWithdrawn(ctx context.Context, rewardID int64, delta float64) error
	FindByUserID(ctx context.Context, userID int64) ([]*model.Reward, error)
	GetByID(ctx context.Context, id int64) (*model.Reward, error)
	FindByDepositID(ctx context.Context, depositID int64) (*model.Reward, error)
	UpdateAmountAndLastAccruedAt(ctx context.Context, rewardID int64, delta float64, accruedAt time.Time) error
	GetTotalAvailableAmount(ctx context.Context) (float64, error)
}
