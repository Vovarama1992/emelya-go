package user

type TarifType string

const (
	TarifLegkiyStart TarifType = "Легкий старт"
	TarifTriumf      TarifType = "Триумф"
	TarifMaksimum    TarifType = "Максимум"
)

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
	ReferrerID      *int
	CardNumber      *string
	Balance         float64   `json:"balance"`
	Tarif           TarifType `json:"tarif"`
}
