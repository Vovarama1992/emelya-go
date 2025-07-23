package money_usecase

import (
	"context"
	"errors"
	"time"

	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
)

var (
	ErrRewardNotFound = errors.New("награда не найдена")
)

type RewardService struct {
	repo ports.RewardRepository
}

func NewRewardService(repo ports.RewardRepository) *RewardService {
	return &RewardService{repo: repo}
}

func (s *RewardService) Create(ctx context.Context, reward *model.Reward) error {
	return s.repo.Create(ctx, reward)
}

func (s *RewardService) GetByID(ctx context.Context, id int64) (*model.Reward, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RewardService) UpdateWithdrawn(ctx context.Context, rewardID int64, delta float64) error {
	return s.repo.UpdateWithdrawn(ctx, rewardID, delta)
}

func (s *RewardService) FindByUserID(ctx context.Context, userID int64) ([]*model.Reward, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *RewardService) AccrueDailyRewardForDeposit(ctx context.Context, depositID int64, dailyReward float64) error {
	reward, err := s.repo.FindByDepositID(ctx, depositID)
	if err != nil {
		return err
	}

	now := time.Now()
	last := reward.LastAccruedAt
	if last == nil {
		// Первый раз, ничего не начисляем задним числом, просто фиксируем дату
		return s.repo.UpdateAmountAndLastAccruedAt(ctx, reward.ID, 0, now)
	}

	daysMissed := int(now.Sub(*last).Hours() / 24)
	if daysMissed <= 0 {
		return nil
	}

	delta := float64(daysMissed) * dailyReward
	newAccruedAt := last.Add(time.Duration(daysMissed) * 24 * time.Hour)

	return s.repo.UpdateAmountAndLastAccruedAt(ctx, reward.ID, delta, newAccruedAt)
}

func (s *RewardService) GetTotalAvailableAmount(ctx context.Context) (float64, error) {
	return s.repo.GetTotalAvailableAmount(ctx)
}
