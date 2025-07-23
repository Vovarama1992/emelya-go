package tariffhttp

type CreateTariffRequest struct {
	Name        string   `json:"name" validate:"required"`
	BlockUntil  *string  `json:"block_until,omitempty"` // ISO8601 строка или nil
	DailyReward *float64 `json:"daily_reward,omitempty"`
}

type UpdateTariffRequest struct {
	ID          int64    `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	BlockUntil  *string  `json:"block_until,omitempty"`
	DailyReward *float64 `json:"daily_reward,omitempty"`
}
