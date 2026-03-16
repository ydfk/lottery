package lottery

import (
	"context"

	"go-fiber-starter/pkg/logger"

	"github.com/robfig/cron/v3"
)

func startSyncLoop(ctx context.Context) {
	scheduler := cron.New(cron.WithSeconds())

	for _, definition := range ListDefinitions() {
		if !definition.Enabled || !definition.Sync.Enabled || definition.Sync.Cron == "" {
			continue
		}

		code := definition.Code
		_, err := scheduler.AddFunc(definition.Sync.Cron, func() {
			if _, syncErr := SyncLatestDraw(ctx, code, ""); syncErr != nil {
				logger.Error("定时同步 %s 失败: %v", code, syncErr)
				return
			}
			logger.Info("已按计划同步 %s 当期开奖", code)
		})
		if err != nil {
			logger.Warn("忽略非法 cron 配置 %s: %v", code, err)
		}
	}

	scheduler.Start()
	<-ctx.Done()
	stopContext := scheduler.Stop()
	<-stopContext.Done()
}
