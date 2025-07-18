package withdrawal_infra

import (
	"context"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/db"
	model "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Интерфейс для работы и с пулом, и с транзакцией
type PgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type WithdrawalRepository struct {
	querier PgxQuerier
}

func NewWithdrawalRepository(db *db.DB) *WithdrawalRepository {
	return &WithdrawalRepository{querier: db.Pool}
}

func NewWithdrawalRepositoryWithTx(tx pgx.Tx) *WithdrawalRepository {
	return &WithdrawalRepository{querier: tx}
}

func (r *WithdrawalRepository) Create(ctx context.Context, w *model.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (user_id, reward_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.querier.QueryRow(ctx, query,
		w.UserID,
		w.RewardID,
		w.Amount,
		w.Status,
		time.Now(),
	).Scan(&w.ID)
	return err
}

func (r *WithdrawalRepository) UpdateStatus(ctx context.Context, id int64, status string, approvedAt, rejectedAt *time.Time, reason *string) error {
	query := `
		UPDATE withdrawals
		SET status = $1, approved_at = $2, rejected_at = $3, reason = $4
		WHERE id = $5
	`
	_, err := r.querier.Exec(ctx, query, status, approvedAt, rejectedAt, reason, id)
	return err
}

func (r *WithdrawalRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	query := `
		SELECT id, user_id, reward_id, amount, status, created_at, approved_at, rejected_at, reason
		FROM withdrawals
		WHERE user_id = $1
	`
	rows, err := r.querier.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.RewardID,
			&w.Amount,
			&w.Status,
			&w.CreatedAt,
			&w.ApprovedAt,
			&w.RejectedAt,
			&w.Reason,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, nil
}

func (r *WithdrawalRepository) GetByID(ctx context.Context, id int64) (*model.Withdrawal, error) {
	query := `
		SELECT id, user_id, reward_id, amount, status, created_at, approved_at, rejected_at, reason
		FROM withdrawals
		WHERE id = $1
	`
	var w model.Withdrawal
	err := r.querier.QueryRow(ctx, query, id).Scan(
		&w.ID,
		&w.UserID,
		&w.RewardID,
		&w.Amount,
		&w.Status,
		&w.CreatedAt,
		&w.ApprovedAt,
		&w.RejectedAt,
		&w.Reason,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WithdrawalRepository) FindAll(ctx context.Context) ([]*model.Withdrawal, error) {
	query := `
		SELECT id, user_id, reward_id, amount, status, created_at, approved_at, rejected_at, reason
		FROM withdrawals
		ORDER BY created_at DESC
	`
	rows, err := r.querier.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.RewardID,
			&w.Amount,
			&w.Status,
			&w.CreatedAt,
			&w.ApprovedAt,
			&w.RejectedAt,
			&w.Reason,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, nil
}

func (r *WithdrawalRepository) FindAllPendings(ctx context.Context) ([]*model.Withdrawal, error) {
	query := `
		SELECT id, user_id, reward_id, amount, status, created_at, approved_at, rejected_at, reason
		FROM withdrawals
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`
	rows, err := r.querier.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.RewardID,
			&w.Amount,
			&w.Status,
			&w.CreatedAt,
			&w.ApprovedAt,
			&w.RejectedAt,
			&w.Reason,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, nil
}
