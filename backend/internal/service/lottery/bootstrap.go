package lottery

import (
	"context"

	"go-fiber-starter/pkg/logger"
)

func Bootstrap() error {
	if err := SeedLotteryTypes(); err != nil {
		return err
	}

	if hasEnabledSyncLotteries() {
		go startSyncLoop(context.Background())
		logger.Info("彩票开奖同步任务已启动")
	}

	return nil
}

func hasEnabledSyncLotteries() bool {
	for _, definition := range ListDefinitions() {
		if definition.Enabled && definition.Sync.Enabled && definition.Sync.Cron != "" {
			return true
		}
	}
	return false
}
