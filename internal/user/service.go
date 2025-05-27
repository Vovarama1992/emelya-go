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

func (s *Service) UpdateProfile(ctx context.Context, userID int, cardNumber string) error {
	return s.repo.UpdateProfile(ctx, userID, cardNumber)
}
