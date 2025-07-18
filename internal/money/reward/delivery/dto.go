package rewardhttp

type AdminCreateReferralRewardRequest struct {
	UserID int64   `json:"user_id" validate:"required"`
	Amount float64 `json:"amount" validate:"required"`
}
