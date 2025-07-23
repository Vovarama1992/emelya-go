package model_deposit

import (
	"time"
)

type Status string

type TarifType string

const (
	StatusPending  Status = "pending"  // ожидает подтверждения
	StatusApproved Status = "approved" // активный
	StatusClosed   Status = "closed"   // закрыт, разблокирован
)

type Deposit struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Amount      float64    `json:"amount"`
	CreatedAt   time.Time  `json:"created_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	BlockUntil  *time.Time `json:"block_until,omitempty"`
	DailyReward *float64   `json:"daily_reward,omitempty"`
	Status      Status     `json:"status"`
}
