package money_ports

import (
	"context"
	"time"

	model "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/model"
)

type WithdrawalRepository interface {
	Create(ctx context.Context, w *model.Withdrawal) error
	UpdateStatus(ctx context.Context, id int64, status string, approvedAt, rejectedAt *time.Time, reason *string) error
	FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
	GetByID(ctx context.Context, id int64) (*model.Withdrawal, error)
	FindAll(ctx context.Context) ([]*model.Withdrawal, error)
	FindAllPendings(ctx context.Context) ([]*model.Withdrawal, error)
}
