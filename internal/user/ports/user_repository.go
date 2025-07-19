package user

import (
	"context"

	model "github.com/Vovarama1992/emelya-go/internal/user/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindUserByPhone(ctx context.Context, phone string) (*model.User, error)
	FindUserByLogin(ctx context.Context, login string) (*model.User, error)
	SetEmailVerified(ctx context.Context, userID int64) error
	SetPhoneVerified(ctx context.Context, userID int64) error
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	UpdateBalance(ctx context.Context, userID int64, balance float64) error
	UpdateCardNumber(ctx context.Context, userID int64, cardNumber string) error
	GetAllUsers(ctx context.Context) ([]model.User, error)
}
