package lottery

import (
	"context"

	"go-fiber-starter/pkg/logger"
)

func Bootstrap() error {
	if err := SeedLotteryTypes(); err != nil {
		return err
	}
	if err := repairLotteryState(); err != nil {
		return err
	}

	if hasScheduledLotteries() {
		go startSyncLoop(context.Background())
		logger.Info("彩票定时任务已启动")
	}

	return nil
}

func repairLotteryState() error {
	for _, definition := range ListDefinitions() {
		if !definition.Enabled {
			continue
		}
		if err := normalizeLocalScheduleDates(definition); err != nil {
			return err
		}
		if err := EvaluatePendingTickets(definition.Code); err != nil {
			return err
		}
	}
	return nil
}

func hasScheduledLotteries() bool {
	for _, definition := range ListDefinitions() {
		if definition.Enabled && definition.Sync.Enabled && definition.Sync.Cron != "" {
			return true
		}
		if definition.Enabled && definition.Recommendation.Enabled && definition.Recommendation.Cron != "" {
			return true
		}
	}
	return false
}
