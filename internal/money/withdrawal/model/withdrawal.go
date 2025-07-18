package withdrawal_model

import "time"

type WithdrawalStatus string

const (
	WithdrawalStatusPending  WithdrawalStatus = "pending"
	WithdrawalStatusApproved WithdrawalStatus = "approved"
	WithdrawalStatusRejected WithdrawalStatus = "rejected"
)

type Withdrawal struct {
	ID         int64            `json:"id"`
	UserID     int64            `json:"user_id"`
	RewardID   int64            `json:"reward_id"`
	Amount     float64          `json:"amount"`
	Status     WithdrawalStatus `json:"status"`
	CreatedAt  time.Time        `json:"created_at"`
	ApprovedAt *time.Time       `json:"approved_at,omitempty"`
	RejectedAt *time.Time       `json:"rejected_at,omitempty"`
	Reason     *string          `json:"reason,omitempty"`
}
