package user

import (
	"context"

	models "github.com/Vovarama1992/emelya-go/internal/user/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindUserByPhone(ctx context.Context, phone string) (*models.User, error)
	FindUserByLogin(ctx context.Context, login string) (*models.User, error)
	SetEmailVerified(ctx context.Context, userID int64) error
	SetPhoneVerified(ctx context.Context, userID int64) error
	FindUserByID(ctx context.Context, userID int64) (*models.User, error)
	UpdateBalance(ctx context.Context, userID int64, balance float64) error
	UpdateCardNumber(ctx context.Context, userID int64, cardNumber string) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
}
