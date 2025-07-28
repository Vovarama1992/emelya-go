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
	"github.com/Vovarama1992/go-utils/ctxutil"
)

var (
	ErrDepositNotFound = errors.New("депозит не найден")
)

type DepositService struct {
	repo      ports.DepositRepository
	rewardSvc ports.RewardService
	tarifSvc  ports.TariffService
	notifier  notifier.NotifierInterface
	db        *db.DB
}

func NewDepositService(
	repo ports.DepositRepository,
	rewardSvc ports.RewardService,
	tarifSvc ports.TariffService,
	db *db.DB,
	notifier *notifier.Notifier,
) *DepositService {
	return &DepositService{
		repo:      repo,
		rewardSvc: rewardSvc,
		tarifSvc:  tarifSvc,
		db:        db,
		notifier:  notifier,
	}
}

// CreateDeposit — создаёт заявку на депозит без транзакции
func (s *DepositService) CreateDeposit(ctx context.Context, userID int64, amount float64) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

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

func (s *DepositService) ApproveDeposit(
	ctx context.Context,
	depositID int64,
	approvedAt time.Time,
	blockUntil *time.Time,
	dailyReward *float64,
	tariffID *int64,
) (err error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

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

	if (blockUntil == nil || dailyReward == nil) && tariffID == nil {
		return errors.New("либо передайте blockUntil/dailyReward, либо tariffID")
	}

	if tariffID != nil {
		tariff, err := s.tarifSvc.FindByID(ctx, *tariffID)
		if err != nil {
			return fmt.Errorf("не удалось найти тариф: %w", err)
		}
		if tariff.BlockUntil == nil || tariff.DailyReward == nil {
			return fmt.Errorf("у тарифа нет необходимых полей blockUntil или dailyReward")
		}
		blockUntil = tariff.BlockUntil
		dailyReward = tariff.DailyReward
	}

	err = txDepositRepo.Approve(ctx, depositID, approvedAt, *blockUntil, *dailyReward)
	if err != nil {
		return err
	}

	reward := &reward_model.Reward{
		UserID:    deposit.UserID,
		DepositID: &deposit.ID,
		Type:      "deposit",
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
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	return s.repo.FindPending(ctx)
}

func (s *DepositService) GetDepositByID(ctx context.Context, id int64) (*model.Deposit, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	return s.repo.FindByID(ctx, id)
}

func (s *DepositService) GetDepositsByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	return s.repo.FindByUserID(ctx, userID)
}

func (s *DepositService) CloseDeposit(ctx context.Context, id int64) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	return s.repo.Close(ctx, id)
}

func (s *DepositService) AccrueDailyRewardsForAllDeposits(ctx context.Context) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

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
	dailyReward *float64,
	tariffID *int64,
	initialRewardAmount *float64, // ← добавили
) (int64, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return 0, err
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

	if (blockUntil == nil || dailyReward == nil) && tariffID == nil {
		return 0, errors.New("либо передайте blockUntil/dailyReward, либо tariffID")
	}

	if tariffID != nil {
		tariff, err := s.tarifSvc.FindByID(ctx, *tariffID)
		if err != nil {
			return 0, fmt.Errorf("не удалось найти тариф: %w", err)
		}
		if tariff.BlockUntil == nil || tariff.DailyReward == nil {
			return 0, fmt.Errorf("у тарифа нет необходимых полей blockUntil или dailyReward")
		}
		blockUntil = tariff.BlockUntil
		dailyReward = tariff.DailyReward
	}

	deposit := &model.Deposit{
		UserID:      userID,
		Amount:      amount,
		CreatedAt:   createdAt,
		ApprovedAt:  approvedAt,
		BlockUntil:  blockUntil,
		DailyReward: dailyReward,
		Status:      model.StatusApproved,
	}

	err = txDepositRepo.CreateApproved(ctx, deposit)
	if err != nil {
		return 0, err
	}

	rewardAmount := 0.0
	if initialRewardAmount != nil {
		rewardAmount = *initialRewardAmount
	}

	reward := &reward_model.Reward{
		UserID:    userID,
		DepositID: &deposit.ID,
		Type:      "deposit",
		Amount:    rewardAmount,
		Withdrawn: 0,
		CreatedAt: time.Now(),
	}

	err = txRewardRepo.Create(ctx, reward)
	if err != nil {
		return 0, err
	}

	return deposit.ID, nil
}

func (s *DepositService) DeleteDepositByAdmin(ctx context.Context, id int64) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	return s.repo.Delete(ctx, id)
}

func (s *DepositService) GetTotalApprovedAmount(ctx context.Context) (float64, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	return s.repo.GetTotalApprovedAmount(ctx)
}

func (s *DepositService) GetAllApprovedDeposits(ctx context.Context) ([]*model.Deposit, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindAllApproved(ctx)
}

func (s *DepositService) GetApprovedDepositsByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindApprovedByUserID(ctx, userID)
}
