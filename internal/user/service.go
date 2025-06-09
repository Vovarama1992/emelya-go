package user

import (
	"context"

	"github.com/Vovarama1992/emelya-go/internal/notifier"
)

type Service struct {
	repo     Repository
	notifier *notifier.Notifier
}

func NewService(repo Repository, notifier *notifier.Notifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

func (s *Service) UpdateCardNumber(ctx context.Context, userID int, cardNumber string) error {
	return s.repo.UpdateCardNumber(ctx, userID, cardNumber)
}

func (s *Service) UpdateTarif(ctx context.Context, userID int, tarif TarifType) error {
	return s.repo.UpdateTarif(ctx, userID, tarif)
}

func (s *Service) UpdateBalance(ctx context.Context, userID int, balance float64) error {
	return s.repo.UpdateBalance(ctx, userID, balance)
}

func (s *Service) GetAllUsers(ctx context.Context) ([]User, error) {
	return s.repo.GetAllUsers(ctx)
}
