package deposithttp

type DepositCreateRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type AdminCreateDepositRequest struct {
	Amount              float64  `json:"amount" validate:"required"`
	CreatedAt           string   `json:"created_at" validate:"required"`
	ApprovedAt          string   `json:"approved_at,omitempty"`
	BlockDays           *int     `json:"block_days,omitempty"`
	DailyReward         *float64 `json:"daily_reward,omitempty"`
	TariffID            *int64   `json:"tariff_id,omitempty"`
	InitialRewardAmount *float64 `json:"initial_reward_amount,omitempty"`
}
