package deposit_infra

import (
	"context"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/db"
	model "github.com/Vovarama1992/emelya-go/internal/money/deposit/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Интерфейс для пула и транзакции
type PgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type DepositRepository struct {
	querier PgxQuerier
}

func NewDepositRepository(db *db.DB) *DepositRepository {
	return &DepositRepository{querier: db.Pool}
}

func NewDepositRepositoryWithTx(tx pgx.Tx) *DepositRepository {
	return &DepositRepository{querier: tx}
}

func (r *DepositRepository) Create(ctx context.Context, d *model.Deposit) error {
	query := `
		INSERT INTO deposits (user_id, amount, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.querier.QueryRow(ctx, query, d.UserID, d.Amount, d.Status).Scan(&d.ID, &d.CreatedAt)
}

func (r *DepositRepository) Approve(ctx context.Context, id int64, approvedAt time.Time, blockDays int, dailyReward float64) error {
	query := `
		UPDATE deposits
		SET approved_at = $1, block_until = $2, daily_reward = $3, status = 'approved'
		WHERE id = $4
	`
	_, err := r.querier.Exec(ctx, query, approvedAt, blockDays, dailyReward, id)
	return err
}

func (r *DepositRepository) FindByID(ctx context.Context, id int64) (*model.Deposit, error) {
	query := `
		SELECT id, user_id, amount, created_at, approved_at, block_days, daily_reward, status
		FROM deposits
		WHERE id = $1
	`
	var d model.Deposit
	err := r.querier.QueryRow(ctx, query, id).Scan(
		&d.ID,
		&d.UserID,
		&d.Amount,
		&d.CreatedAt,
		&d.ApprovedAt,
		&d.BlockDays,
		&d.DailyReward,
		&d.Status,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DepositRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error) {
	query := `
		SELECT id, user_id, amount, created_at, approved_at, block_days, daily_reward, status
		FROM deposits
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.querier.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*model.Deposit
	for rows.Next() {
		var d model.Deposit
		if err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.Amount,
			&d.CreatedAt,
			&d.ApprovedAt,
			&d.BlockDays,
			&d.DailyReward,
			&d.Status,
		); err != nil {
			return nil, err
		}
		deposits = append(deposits, &d)
	}
	return deposits, nil
}

func (r *DepositRepository) Close(ctx context.Context, id int64) error {
	query := `
		UPDATE deposits
		SET status = 'closed'
		WHERE id = $1
	`
	_, err := r.querier.Exec(ctx, query, id)
	return err
}

func (r *DepositRepository) FindPending(ctx context.Context) ([]*model.Deposit, error) {
	query := `
		SELECT id, user_id, amount, created_at, approved_at, block_days, daily_reward, status
		FROM deposits
		WHERE status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.querier.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*model.Deposit
	for rows.Next() {
		var d model.Deposit
		if err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.Amount,
			&d.CreatedAt,
			&d.ApprovedAt,
			&d.BlockDays,
			&d.DailyReward,
			&d.Status,
		); err != nil {
			return nil, err
		}
		deposits = append(deposits, &d)
	}
	return deposits, nil
}

func (r *DepositRepository) CreateApproved(ctx context.Context, d *model.Deposit) error {
	query := `
		INSERT INTO deposits (user_id, amount, created_at, approved_at, block_days, daily_reward, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.querier.QueryRow(ctx, query,
		d.UserID,
		d.Amount,
		d.CreatedAt,
		d.ApprovedAt,
		d.BlockDays,
		d.DailyReward,
		d.Status, // передаём как $7
	).Scan(&d.ID)
}

func (r *DepositRepository) FindAllApproved(ctx context.Context) ([]*model.Deposit, error) {
	query := `
		SELECT id, user_id, amount, created_at, approved_at, block_days, daily_reward, status
		FROM deposits
		WHERE status = 'approved'
		ORDER BY created_at DESC
	`
	rows, err := r.querier.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*model.Deposit
	for rows.Next() {
		var d model.Deposit
		if err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.Amount,
			&d.CreatedAt,
			&d.ApprovedAt,
			&d.BlockDays,
			&d.DailyReward,
			&d.Status,
		); err != nil {
			return nil, err
		}
		deposits = append(deposits, &d)
	}
	return deposits, nil
}

func (r *DepositRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM deposits WHERE id = $1`
	_, err := r.querier.Exec(ctx, query, id)
	return err
}

func (r *DepositRepository) GetTotalApprovedAmount(ctx context.Context) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(amount), 0) FROM deposits WHERE status = 'approved'`
	err := r.querier.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *DepositRepository) FindApprovedByUserID(ctx context.Context, userID int64) ([]*model.Deposit, error) {
	query := `
		SELECT id, user_id, amount, created_at, approved_at, block_days, daily_reward, status
		FROM deposits
		WHERE user_id = $1 AND status = 'approved'
		ORDER BY created_at DESC
	`
	rows, err := r.querier.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*model.Deposit
	for rows.Next() {
		var d model.Deposit
		if err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.Amount,
			&d.CreatedAt,
			&d.ApprovedAt,
			&d.BlockDays,
			&d.DailyReward,
			&d.Status,
		); err != nil {
			return nil, err
		}
		deposits = append(deposits, &d)
	}
	return deposits, nil
}
