package user

import (
	"context"

	user "github.com/Vovarama1992/emelya-go/internal/user/models"
)

type UserServiceInterface interface {
	CreateUser(ctx context.Context, newUser *user.User) error
	FindUserByID(ctx context.Context, userID int64) (*user.User, error)
	FindUserByPhone(ctx context.Context, phone string) (*user.User, error)
	FindUserByLogin(ctx context.Context, login string) (*user.User, error)
	VerifyPhone(ctx context.Context, userID int64) error
	UpdateCardNumber(ctx context.Context, userID int64, cardNumber string) error
	UpdateBalance(ctx context.Context, userID int64, balance float64) error
	GetAllUsers(ctx context.Context) ([]user.User, error)
}
