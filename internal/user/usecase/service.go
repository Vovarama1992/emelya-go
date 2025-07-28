package user

import (
	"context"

	deposit "github.com/Vovarama1992/emelya-go/internal/money/deposit/model"
	money_ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
	model "github.com/Vovarama1992/emelya-go/internal/user/model"
	ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
	"github.com/Vovarama1992/go-utils/ctxutil"
)

type Service struct {
	repo       ports.UserRepository
	notifier   *notifier.Notifier
	depositSvc money_ports.DepositService
	rewardSvc  money_ports.RewardService
}

func NewService(
	repo ports.UserRepository,
	notifier *notifier.Notifier,
	depositSvc money_ports.DepositService,
	rewardSvc money_ports.RewardService,
) *Service {
	return &Service{
		repo:       repo,
		notifier:   notifier,
		depositSvc: depositSvc,
		rewardSvc:  rewardSvc,
	}
}

func (s *Service) GetAllUsers(ctx context.Context) ([]model.User, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.GetAllUsers(ctx)
}

func (s *Service) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindUserByID(ctx, userID)
}

func (s *Service) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindUserByPhone(ctx, phone)
}

func (s *Service) FindUserByLogin(ctx context.Context, login string) (*model.User, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.FindUserByLogin(ctx, login)
}

func (s *Service) VerifyPhone(ctx context.Context, userID int64) error {
	return s.repo.SetPhoneVerified(ctx, userID)
}

func (s *Service) CreateUser(ctx context.Context, newUser *model.User) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.CreateUser(ctx, newUser)
}

func (s *Service) UpdateProfile(ctx context.Context, user *model.User) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.UpdateProfile(ctx, user)
}

func (s *Service) SetReferrer(ctx context.Context, userID int64, referrerID int64) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.repo.SetReferrer(ctx, userID, referrerID)
}

func (s *Service) GetCurrentBalance(ctx context.Context, userID int64) (float64, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	deposits, err := s.depositSvc.GetApprovedDepositsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	var userDeposits []*deposit.Deposit
	var depositIDs []int64
	for _, d := range deposits {
		if d.UserID == userID {
			userDeposits = append(userDeposits, d)
			depositIDs = append(depositIDs, d.ID)
		}
	}

	if len(userDeposits) == 0 {
		return 0, nil
	}

	rewards, err := s.rewardSvc.FindByDepositIDs(ctx, depositIDs)
	if err != nil {
		return 0, err
	}

	withdrawnMap := make(map[int64]float64)
	for _, r := range rewards {
		if r.DepositID != nil {
			withdrawnMap[*r.DepositID] = r.Withdrawn
		}
	}

	var total float64
	for _, d := range userDeposits {
		withdrawn := withdrawnMap[d.ID]
		total += d.Amount - withdrawn
	}

	return total, nil
}

func (s *Service) GetTotalRewardBalance(ctx context.Context, userID int64) (float64, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()

	rewards, err := s.rewardSvc.FindByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, r := range rewards {
		total += r.Amount - r.Withdrawn
	}

	return total, nil
}
