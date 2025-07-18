package reward

import "time"

type RewardType string

const (
	RewardTypeDeposit  RewardType = "deposit"
	RewardTypeReferral RewardType = "referral"
)

type Reward struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	DepositID     *int64     `json:"deposit_id,omitempty"`
	Type          RewardType `json:"type"`
	Amount        float64    `json:"amount"`
	Withdrawn     float64    `json:"withdrawn"`
	LastAccruedAt *time.Time `json:"last_accrued_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
