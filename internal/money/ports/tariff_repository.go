package money_ports

import (
	"context"

	model "github.com/Vovarama1992/emelya-go/internal/money/tariff/model"
)

type TariffRepository interface {
	GetAll(ctx context.Context) ([]model.Tariff, error)
	Create(ctx context.Context, tariff *model.Tariff) error
	Update(ctx context.Context, tariff *model.Tariff) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*model.Tariff, error)
}
