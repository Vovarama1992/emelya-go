package tariff

import "time"

type Tariff struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	BlockDays   *int      `json:"block_days,omitempty"`
	DailyReward *float64  `json:"daily_reward,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
