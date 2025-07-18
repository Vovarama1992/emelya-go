package withdrawalhttp

type CreateWithdrawalRequest struct {
	RewardID int64   `json:"reward_id" validate:"required"`
	Amount   float64 `json:"amount" validate:"required"`
}

type AdminRejectWithdrawalRequest struct {
	WithdrawalID int64  `json:"withdrawal_id" validate:"required"`
	Reason       string `json:"reason" validate:"required"`
}
