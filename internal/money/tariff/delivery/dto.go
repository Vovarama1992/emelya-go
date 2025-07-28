package tariffhttp

type CreateTariffRequest struct {
	Name        string   `json:"name" validate:"required"`
	BlockDays   *int     `json:"block_days,omitempty"`
	DailyReward *float64 `json:"daily_reward,omitempty"`
}

type UpdateTariffRequest struct {
	ID          int64    `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	BlockDays   *int     `json:"block_days,omitempty"`
	DailyReward *float64 `json:"daily_reward,omitempty"`
}
