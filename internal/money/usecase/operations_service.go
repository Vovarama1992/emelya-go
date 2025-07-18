package money_usecase

import (
	"context"

	operation "github.com/Vovarama1992/emelya-go/internal/money/operation_model"
)

type OperationsService struct {
	depositService    *DepositService
	rewardService     *RewardService
	withdrawalService *WithdrawalService
}

func NewOperationsService(
	depositService *DepositService,
	rewardService *RewardService,
	withdrawalService *WithdrawalService,
) *OperationsService {
	return &OperationsService{
		depositService:    depositService,
		rewardService:     rewardService,
		withdrawalService: withdrawalService,
	}
}

func (s *OperationsService) ListUserOperations(ctx context.Context, userID int64) (*operation.Operations, error) {
	deposits, err := s.depositService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	withdrawals, err := s.withdrawalService.ListWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	rewards, err := s.rewardService.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &operation.Operations{
		Deposits:    deposits,
		Withdrawals: withdrawals,
		Rewards:     rewards,
	}, nil
}
