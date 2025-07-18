package operation

import (
	deposit_model "github.com/Vovarama1992/emelya-go/internal/money/deposit/model"
	reward_model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
	withdrawal_model "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/model"
)

type Operations struct {
	Deposits    []*deposit_model.Deposit
	Withdrawals []*withdrawal_model.Withdrawal
	Rewards     []*reward_model.Reward
}
