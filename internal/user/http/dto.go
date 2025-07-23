package user

type UpdateProfileRequest struct {
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty"`
	Patronymic *string `json:"patronymic,omitempty" validate:"omitempty"`
	Phone      *string `json:"phone,omitempty" validate:"omitempty,e164"`
	CardNumber *string `json:"card_number,omitempty" validate:"omitempty"`
}

type AdminUpdateProfileRequest struct {
	UserID     int64   `json:"user_id" validate:"required"`
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty"`
	Patronymic *string `json:"patronymic,omitempty" validate:"omitempty"`
	Phone      *string `json:"phone,omitempty" validate:"omitempty,e164"`
	CardNumber *string `json:"card_number,omitempty" validate:"omitempty"`
}

type AddReferralRequest struct {
	UserID     int64 `json:"user_id" validate:"required"`
	ReferrerID int64 `json:"referrer_id" validate:"required"`
}

type RequestWithdrawRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0" example:"1500"`
}
