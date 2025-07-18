package authadapter

type RegisterRequest struct {
	FirstName  string `json:"first_name" validate:"required,max=50"`
	LastName   string `json:"last_name" validate:"required,max=50"`
	Patronymic string `json:"patronymic" validate:"omitempty,max=50"`
	Email      string `json:"email" validate:"required,email"`
	Phone      string `json:"phone" validate:"required"`
	ReferrerID *int   `json:"referrerId"`
}

type ConfirmRequest struct {
	Phone string `json:"phone" validate:"required"`
	Code  string `json:"code" validate:"required,len=4"`
}

type PhoneRequest struct {
	Phone string `json:"phone" validate:"required"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}
