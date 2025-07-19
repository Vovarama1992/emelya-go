package scheduler

import (
	"context"
	"fmt"
	"log"

	usecase "github.com/Vovarama1992/emelya-go/internal/money/usecase"
	"github.com/robfig/cron/v3"
)

func StartDepositRewardCron(depositService *usecase.DepositService) *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc("@hourly", func() {
		fmt.Println("[CRON] Начисляю rewards через cron...")
		if err := depositService.AccrueDailyRewardsForAllDeposits(context.Background()); err != nil {
			fmt.Printf("[CRON] Ошибка начисления: %v\n", err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Start()
	return c
}
