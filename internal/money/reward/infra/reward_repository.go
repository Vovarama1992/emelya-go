package reward_infra

import (
	"context"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/db"
	model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Интерфейс для обобщения Pool и Tx
type PgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type RewardRepository struct {
	querier PgxQuerier
}

func NewRewardRepository(db *db.DB) *RewardRepository {
	return &RewardRepository{querier: db.Pool}
}

// Позволяет создать репу с конкретным pgx.Tx для транзакций
func NewRewardRepositoryWithTx(tx pgx.Tx) *RewardRepository {
	return &RewardRepository{querier: tx}
}

func (r *RewardRepository) Create(ctx context.Context, reward *model.Reward) error {
	query := `
		INSERT INTO rewards (user_id, deposit_id, type, amount, withdrawn, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := r.querier.QueryRow(ctx, query,
		reward.UserID,
		reward.DepositID,
		reward.Type,
		reward.Amount,
		reward.Withdrawn,
		time.Now(),
	).Scan(&reward.ID)
	return err
}

func (r *RewardRepository) UpdateWithdrawn(ctx context.Context, rewardID int64, delta float64) error {
	query := `
		UPDATE rewards
		SET withdrawn = withdrawn + $1
		WHERE id = $2
	`
	_, err := r.querier.Exec(ctx, query, delta, rewardID)
	return err
}

func (r *RewardRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Reward, error) {
	query := `
		SELECT id, user_id, deposit_id, type, amount, withdrawn, created_at
		FROM rewards
		WHERE user_id = $1
	`
	rows, err := r.querier.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rewards []*model.Reward
	for rows.Next() {
		var rw model.Reward
		if err := rows.Scan(
			&rw.ID,
			&rw.UserID,
			&rw.DepositID,
			&rw.Type,
			&rw.Amount,
			&rw.Withdrawn,
			&rw.CreatedAt,
		); err != nil {
			return nil, err
		}
		rewards = append(rewards, &rw)
	}
	return rewards, nil
}

func (r *RewardRepository) GetByID(ctx context.Context, id int64) (*model.Reward, error) {
	query := `
		SELECT id, user_id, deposit_id, type, amount, withdrawn, created_at
		FROM rewards
		WHERE id = $1
	`
	var rw model.Reward
	err := r.querier.QueryRow(ctx, query, id).Scan(
		&rw.ID,
		&rw.UserID,
		&rw.DepositID,
		&rw.Type,
		&rw.Amount,
		&rw.Withdrawn,
		&rw.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *RewardRepository) FindByDepositID(ctx context.Context, depositID int64) (*model.Reward, error) {
	query := `
		SELECT id, user_id, deposit_id, type, amount, withdrawn, last_accrued_at, created_at
		FROM rewards
		WHERE deposit_id = $1
	`
	var rw model.Reward
	err := r.querier.QueryRow(ctx, query, depositID).Scan(
		&rw.ID,
		&rw.UserID,
		&rw.DepositID,
		&rw.Type,
		&rw.Amount,
		&rw.Withdrawn,
		&rw.LastAccruedAt,
		&rw.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *RewardRepository) UpdateAmountAndLastAccruedAt(ctx context.Context, rewardID int64, delta float64, accruedAt time.Time) error {
	query := `
		UPDATE rewards
		SET amount = amount + $1,
		    last_accrued_at = $2
		WHERE id = $3
	`
	_, err := r.querier.Exec(ctx, query, delta, accruedAt, rewardID)
	return err
}
