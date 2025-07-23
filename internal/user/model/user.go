package user

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID              int64
	FirstName       string
	LastName        string
	Patronymic      string
	Email           string
	Phone           string
	IsEmailVerified bool
	IsPhoneVerified bool
	Login           string
	PasswordHash    string
	ReferrerID      *int64
	CardNumber      *string
	Role            UserRole `json:"role"`
}
