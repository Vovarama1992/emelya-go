package user

import "context"

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	SetEmailVerified(ctx context.Context, userID int) error
	SetPhoneVerified(ctx context.Context, userID int) error
	GetUserByID(ctx context.Context, userID int) (*User, error)
	UpdateProfile(ctx context.Context, userID int, cardNumber string) error // ← добавляем
}
