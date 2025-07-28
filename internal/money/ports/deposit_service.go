package money_ports

import (
	"context"
	"time"

	model "github.com/Vovarama1992/emelya-go/internal/money/deposit/model"
)

type DepositService interface {
	CreateDeposit(ctx context.Context, userID int64, amount float64) error

	ApproveDeposit(
		ctx context.Context,
		depositID int64,
		approvedAt time.Time,
		blockDays *int,
		dailyReward *float64,
		tariffID *int64,
	) error

	GetDepositByID(ctx context.Context, id int64) (*model.Deposit, error)
	GetDepositsByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error)
	AccrueDailyRewardsForAllDeposits(ctx context.Context) error
	CloseDeposit(ctx context.Context, id int64) error
	ListPendingDeposits(ctx context.Context) ([]*model.Deposit, error)

	CreateDepositByAdmin(
		ctx context.Context,
		userID int64,
		amount float64,
		createdAt time.Time,
		approvedAt *time.Time,
		blockDays *int,
		dailyReward *float64,
		tariffID *int64,
		initialRewardAmount *float64,
	) (int64, error)

	DeleteDepositByAdmin(ctx context.Context, id int64) error
	GetTotalApprovedAmount(ctx context.Context) (float64, error)
	GetAllApprovedDeposits(ctx context.Context) ([]*model.Deposit, error)
	GetApprovedDepositsByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error)
}
