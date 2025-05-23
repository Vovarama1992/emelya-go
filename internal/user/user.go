package user

type User struct {
	ID              int
	FirstName       string
	LastName        string
	Patronymic      string
	Email           string
	Phone           string
	IsEmailVerified bool
	IsPhoneVerified bool
	Login           string
	PasswordHash    string
	ReferrerID      *int // ссылка на другого пользователя, может быть nil
}
