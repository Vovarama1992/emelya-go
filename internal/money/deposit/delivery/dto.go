package deposithttp

type DepositCreateRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type AdminCreateDepositRequest struct {
	UserID      int64   `json:"user_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	Tarif       string  `json:"tarif" validate:"required"`
	CreatedAt   string  `json:"created_at" validate:"required"`
	ApprovedAt  string  `json:"approved_at" validate:"required"`
	BlockUntil  string  `json:"block_until" validate:"required"`
	DailyReward float64 `json:"daily_reward" validate:"required"`
}
