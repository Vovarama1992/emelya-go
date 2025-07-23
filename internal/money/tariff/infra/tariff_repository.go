package tariff_infra

import (
	"context"

	"github.com/Vovarama1992/emelya-go/internal/db"
	model "github.com/Vovarama1992/emelya-go/internal/money/tariff/model"
)

type TariffRepository struct {
	DB *db.DB
}

func NewTariffRepository(db *db.DB) *TariffRepository {
	return &TariffRepository{DB: db}
}

func (r *TariffRepository) GetAll(ctx context.Context) ([]model.Tariff, error) {
	query := `SELECT id, name, block_until, daily_reward, created_at FROM tariffs`
	rows, err := r.DB.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tariffs []model.Tariff
	for rows.Next() {
		var t model.Tariff
		err := rows.Scan(&t.ID, &t.Name, &t.BlockUntil, &t.DailyReward, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		tariffs = append(tariffs, t)
	}

	return tariffs, nil
}

func (r *TariffRepository) Create(ctx context.Context, tariff *model.Tariff) error {
	query := `
		INSERT INTO tariffs (name, block_until, daily_reward)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.DB.Pool.QueryRow(ctx, query,
		tariff.Name,
		tariff.BlockUntil,
		tariff.DailyReward,
	).Scan(&tariff.ID, &tariff.CreatedAt)
}

func (r *TariffRepository) Update(ctx context.Context, tariff *model.Tariff) error {
	query := `
		UPDATE tariffs
		SET name = $1, block_until = $2, daily_reward = $3
		WHERE id = $4
	`
	_, err := r.DB.Pool.Exec(ctx, query,
		tariff.Name,
		tariff.BlockUntil,
		tariff.DailyReward,
		tariff.ID,
	)
	return err
}

func (r *TariffRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tariffs WHERE id = $1`
	_, err := r.DB.Pool.Exec(ctx, query, id)
	return err
}

func (r *TariffRepository) FindByID(ctx context.Context, id int64) (*model.Tariff, error) {
	query := `SELECT id, name, block_until, daily_reward, created_at FROM tariffs WHERE id = $1`
	row := r.DB.Pool.QueryRow(ctx, query, id)

	var t model.Tariff
	err := row.Scan(&t.ID, &t.Name, &t.BlockUntil, &t.DailyReward, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
