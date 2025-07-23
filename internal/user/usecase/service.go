package user

import (
	"context"

	"github.com/Vovarama1992/emelya-go/internal/notifier"
	model "github.com/Vovarama1992/emelya-go/internal/user/model"
	ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
)

type Service struct {
	repo     ports.UserRepository
	notifier *notifier.Notifier
}

func NewService(repo ports.UserRepository, notifier *notifier.Notifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

func (s *Service) GetAllUsers(ctx context.Context) ([]model.User, error) {
	return s.repo.GetAllUsers(ctx)
}

func (s *Service) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func (s *Service) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	return s.repo.FindUserByPhone(ctx, phone)
}

func (s *Service) FindUserByLogin(ctx context.Context, login string) (*model.User, error) {
	return s.repo.FindUserByLogin(ctx, login)
}

func (s *Service) VerifyPhone(ctx context.Context, userID int64) error {
	return s.repo.SetPhoneVerified(ctx, userID)
}

func (s *Service) CreateUser(ctx context.Context, newUser *model.User) error {
	return s.repo.CreateUser(ctx, newUser)
}

func (s *Service) UpdateProfile(ctx context.Context, user *model.User) error {
	return s.repo.UpdateProfile(ctx, user)
}

func (s *Service) SetReferrer(ctx context.Context, userID int64, referrerID int64) error {
	return s.repo.SetReferrer(ctx, userID, referrerID)
}
