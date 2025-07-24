package money_usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/db"
	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	reward_infra "github.com/Vovarama1992/emelya-go/internal/money/reward/infra"
	withdrawal_infra "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/infra"
	model "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/model"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
	"github.com/Vovarama1992/go-utils/ctxutil"
)

var (
	ErrInsufficientFunds = errors.New("недостаточно средств для вывода")
	ErrNotFound          = errors.New("заявка не найдена")
	ErrAlreadyProcessed  = errors.New("заявка уже обработана")
)

type WithdrawalService struct {
	repo      ports.WithdrawalRepository
	rewardSvc ports.RewardService
	notifier  notifier.NotifierInterface
	db        *db.DB
}

func NewWithdrawalService(
	repo ports.WithdrawalRepository,
	rewardSvc ports.RewardService,
	db *db.DB,
	notifier *notifier.Notifier,
) *WithdrawalService {
	return &WithdrawalService{
		repo:      repo,
		rewardSvc: rewardSvc,
		db:        db,
		notifier:  notifier,
	}
}

// Создание заявки на вывод без транзакции
func (s *WithdrawalService) CreateWithdrawal(ctx context.Context, userID, rewardID int64, amount float64) error {
	reward, err := s.rewardSvc.GetByID(ctx, rewardID)
	if err != nil {
		return err
	}

	available := reward.Amount - reward.Withdrawn
	if available < amount {
		return ErrInsufficientFunds
	}

	withdrawal := &model.Withdrawal{
		UserID:    userID,
		RewardID:  rewardID,
		Amount:    amount,
		Status:    model.WithdrawalStatusPending,
		CreatedAt: time.Now(),
	}

	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	if err := s.repo.Create(ctx, withdrawal); err != nil {
		return err
	}

	// Отправляем уведомление
	subject := "Новая заявка на вывод средств"
	body := fmt.Sprintf(
		"Пользователь ID: %d подал заявку на вывод %.2f руб. с reward ID: %d",
		userID, amount, rewardID,
	)

	if err := s.notifier.SendEmailToOperator(subject, body); err != nil {
		// Не падаем, если уведомление не сработало, просто пишем в лог
		fmt.Printf("[WITHDRAWAL] Не удалось отправить уведомление оператору: %v\n", err)
	}

	return nil
}

// Подтверждение заявки с обновлением награды в транзакции
func (s *WithdrawalService) ApproveWithdrawal(ctx context.Context, withdrawalID int64) (err error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 4)
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

	txWithdrawalRepo := withdrawal_infra.NewWithdrawalRepositoryWithTx(tx)
	txRewardRepo := reward_infra.NewRewardRepositoryWithTx(tx)

	withdrawal, err := txWithdrawalRepo.GetByID(ctx, withdrawalID)
	if err != nil {
		return ErrNotFound
	}

	if withdrawal.Status != model.WithdrawalStatusPending {
		return ErrAlreadyProcessed
	}

	reward, err := txRewardRepo.GetByID(ctx, withdrawal.RewardID)
	if err != nil {
		return err
	}

	available := reward.Amount - reward.Withdrawn
	if available < withdrawal.Amount {
		return ErrInsufficientFunds
	}

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusApproved
	withdrawal.ApprovedAt = &now

	err = txWithdrawalRepo.UpdateStatus(ctx, withdrawal.ID, string(withdrawal.Status), withdrawal.ApprovedAt, nil, nil)
	if err != nil {
		return err
	}

	err = txRewardRepo.UpdateWithdrawn(ctx, reward.ID, withdrawal.Amount)
	if err != nil {
		return err
	}

	return nil
}

// Отклонение заявки на вывод
func (s *WithdrawalService) RejectWithdrawal(ctx context.Context, withdrawalID int64, reason string) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 3)
	defer cancel()
	withdrawal, err := s.repo.GetByID(ctx, withdrawalID)
	if err != nil {
		return ErrNotFound
	}

	if withdrawal.Status != model.WithdrawalStatusPending {
		return ErrAlreadyProcessed
	}

	withdrawal.Status = model.WithdrawalStatusRejected
	now := time.Now()
	withdrawal.RejectedAt = &now
	withdrawal.Reason = &reason

	return s.repo.UpdateStatus(ctx, withdrawal.ID, string(withdrawal.Status), nil, withdrawal.RejectedAt, &reason)
}

// Список заявок конкретного пользователя
func (s *WithdrawalService) ListWithdrawalsByUser(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindByUserID(ctx, userID)
}

// Список всех заявок (для админки)
func (s *WithdrawalService) ListAllWithdrawals(ctx context.Context) ([]*model.Withdrawal, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindAll(ctx)
}

// Список всех заявок в статусе pending
func (s *WithdrawalService) ListPendingWithdrawals(ctx context.Context) ([]*model.Withdrawal, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindAllPendings(ctx)
}
