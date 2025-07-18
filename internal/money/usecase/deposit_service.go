package money_usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/db"
	deposit_infra "github.com/Vovarama1992/emelya-go/internal/money/deposit/infra"
	model "github.com/Vovarama1992/emelya-go/internal/money/deposit/model"
	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	reward_infra "github.com/Vovarama1992/emelya-go/internal/money/reward/infra"
	reward_model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
)

var (
	ErrDepositNotFound = errors.New("депозит не найден")
)

type DepositService struct {
	repo      ports.DepositRepository
	rewardSvc ports.RewardService
	notifier  notifier.NotifierInterface
	db        *db.DB
}

func NewDepositService(
	repo ports.DepositRepository,
	rewardSvc ports.RewardService,
	db *db.DB,
	notifier *notifier.Notifier,
) *DepositService {
	return &DepositService{
		repo:      repo,
		rewardSvc: rewardSvc,
		db:        db,
		notifier:  notifier,
	}
}

// CreateDeposit — создаёт заявку на депозит без транзакции
func (s *DepositService) CreateDeposit(ctx context.Context, userID int64, amount float64) error {
	deposit := &model.Deposit{
		UserID: userID,
		Amount: amount,
		Status: model.StatusPending,
	}

	if err := s.repo.Create(ctx, deposit); err != nil {
		return err
	}

	// Отправляем уведомление операторам
	subject := "Новая заявка на депозит"
	body := fmt.Sprintf(
		"Пользователь ID: %d подал заявку на депозит на сумму %.2f руб.",
		userID, amount,
	)

	if err := s.notifier.SendEmailToOperator(subject, body); err != nil {
		fmt.Printf("[DEPOSIT] Не удалось отправить уведомление оператору: %v\n", err)
	}

	return nil
}

// ApproveDeposit — одобряет депозит и создаёт ревард в рамках транзакции
func (s *DepositService) ApproveDeposit(ctx context.Context, depositID int64, approvedAt, blockUntil time.Time, dailyReward float64) (err error) {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	txDepositRepo := deposit_infra.NewDepositRepositoryWithTx(tx)
	txRewardRepo := reward_infra.NewRewardRepositoryWithTx(tx)

	deposit, err := txDepositRepo.FindByID(ctx, depositID)
	if err != nil {
		return ErrDepositNotFound
	}

	err = txDepositRepo.Approve(ctx, depositID, approvedAt, blockUntil, dailyReward)
	if err != nil {
		return err
	}

	reward := &reward_model.Reward{
		UserID:    deposit.UserID,
		DepositID: &deposit.ID,
		Type:      "deposit", // или лучше сделать константу
		Amount:    deposit.Amount,
		Withdrawn: 0,
		CreatedAt: time.Now(),
	}

	err = txRewardRepo.Create(ctx, reward)
	if err != nil {
		return err
	}

	return nil
}

func (s *DepositService) ListPendingDeposits(ctx context.Context) ([]*model.Deposit, error) {
	return s.repo.FindPending(ctx)
}

func (s *DepositService) GetDepositByID(ctx context.Context, id int64) (*model.Deposit, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DepositService) GetDepositsByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *DepositService) CloseDeposit(ctx context.Context, id int64) error {
	return s.repo.Close(ctx, id)
}

func (s *DepositService) AccrueDailyRewardsForAllDeposits(ctx context.Context) error {
	deposits, err := s.repo.FindAllApproved(ctx)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		// safety guard: без daily reward нам нечего делать
		if deposit.DailyReward == nil {
			continue
		}

		err := s.rewardSvc.AccrueDailyRewardForDeposit(ctx, deposit.ID, *deposit.DailyReward)
		if err != nil {
			return fmt.Errorf("failed to accrue reward for deposit ID %d: %w", deposit.ID, err)
		}
	}

	return nil
}

func (s *DepositService) CreateDepositByAdmin(
	ctx context.Context,
	userID int64,
	amount float64,
	createdAt time.Time,
	approvedAt *time.Time,
	blockUntil *time.Time,
	tarif model.TarifType,
	dailyReward *float64,
) (int64, error) {
	deposit := &model.Deposit{
		UserID:      userID,
		Amount:      amount,
		CreatedAt:   createdAt,
		ApprovedAt:  approvedAt,
		BlockUntil:  blockUntil,
		Tarif:       tarif,
		DailyReward: dailyReward,
		Status:      model.StatusApproved,
	}

	err := s.repo.CreateApproved(ctx, deposit)
	if err != nil {
		return 0, err
	}

	return deposit.ID, nil
}

func (s *DepositService) DeleteDepositByAdmin(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
