package tariff

import "time"

type Tariff struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	BlockUntil  *time.Time `json:"block_until,omitempty"`
	DailyReward *float64   `json:"daily_reward,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
