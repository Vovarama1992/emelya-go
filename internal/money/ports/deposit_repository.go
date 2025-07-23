package money_ports

import (
	"context"
	"time"

	models "github.com/Vovarama1992/emelya-go/internal/money/deposit/model"
)

type DepositRepository interface {
	Create(ctx context.Context, deposit *models.Deposit) error
	FindByID(ctx context.Context, id int64) (*models.Deposit, error)
	FindByUserID(ctx context.Context, userID int64) ([]*models.Deposit, error)
	Approve(ctx context.Context, id int64, approvedAt, blockUntil time.Time, dailyReward float64) error
	Close(ctx context.Context, id int64) error
	FindPending(ctx context.Context) ([]*models.Deposit, error)
	FindAllApproved(ctx context.Context) ([]*models.Deposit, error)
	CreateApproved(ctx context.Context, d *models.Deposit) error
	Delete(ctx context.Context, id int64) error
	GetTotalApprovedAmount(ctx context.Context) (float64, error)
}
