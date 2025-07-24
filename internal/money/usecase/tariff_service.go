package money_usecase

import (
	"context"

	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	model "github.com/Vovarama1992/emelya-go/internal/money/tariff/model"
	"github.com/Vovarama1992/go-utils/ctxutil"
)

type TariffService struct {
	repo ports.TariffRepository
}

func NewTariffService(repo ports.TariffRepository) *TariffService {
	return &TariffService{repo: repo}
}

func (s *TariffService) GetAll(ctx context.Context) ([]model.Tariff, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.GetAll(ctx)
}

func (s *TariffService) Create(ctx context.Context, tariff *model.Tariff) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.Create(ctx, tariff)
}

func (s *TariffService) Update(ctx context.Context, tariff *model.Tariff) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.Update(ctx, tariff)
}

func (s *TariffService) Delete(ctx context.Context, id int64) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.Delete(ctx, id)
}

func (s *TariffService) FindByID(ctx context.Context, id int64) (*model.Tariff, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindByID(ctx, id)
}
