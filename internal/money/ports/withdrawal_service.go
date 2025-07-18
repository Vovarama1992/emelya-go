package money_ports

import (
	"context"

	model "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/model"
)

type WithdrawalService interface {
	CreateWithdrawal(ctx context.Context, userID, rewardID int64, amount float64) error
	ApproveWithdrawal(ctx context.Context, withdrawalID int64) error
	RejectWithdrawal(ctx context.Context, withdrawalID int64, reason string) error
	ListWithdrawalsByUser(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
	ListAllWithdrawals(ctx context.Context) ([]*model.Withdrawal, error)
	ListPendingWithdrawals(ctx context.Context) ([]*model.Withdrawal, error)
}
