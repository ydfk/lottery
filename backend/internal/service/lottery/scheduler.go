package lottery

import (
	"context"

	"go-fiber-starter/pkg/logger"

	"github.com/robfig/cron/v3"
)

func startSyncLoop(ctx context.Context) {
	scheduler := cron.New(cron.WithSeconds())

	for _, definition := range ListDefinitions() {
		code := definition.Code
		if definition.Enabled && definition.Sync.Enabled && definition.Sync.Cron != "" {
			_, err := scheduler.AddFunc(definition.Sync.Cron, func() {
				if _, syncErr := SyncLatestDraw(ctx, code, ""); syncErr != nil {
					logger.Error("定时同步 %s 失败: %v", code, syncErr)
					return
				}
				logger.Info("已按计划同步 %s 当期开奖", code)
			})
			if err != nil {
				logger.Warn("忽略非法同步 cron 配置 %s: %v", code, err)
			}
		}

		if definition.Enabled && definition.Recommendation.Enabled && definition.Recommendation.Cron != "" {
			_, err := scheduler.AddFunc(definition.Recommendation.Cron, func() {
				users, userErr := loadSchedulerUsers()
				if userErr != nil {
					logger.Error("加载推荐用户 %s 失败: %v", code, userErr)
					return
				}
				for _, user := range users {
					if _, recommendationErr := GenerateRecommendation(ctx, code, 0, user.Id.String()); recommendationErr != nil {
						logger.Error("定时生成推荐 %s/%s 失败: %v", code, user.Username, recommendationErr)
					}
				}
				logger.Info("已按计划生成 %s 推荐", code)
			})
			if err != nil {
				logger.Warn("忽略非法推荐 cron 配置 %s: %v", code, err)
			}
		}
	}

	scheduler.Start()
	<-ctx.Done()
	stopContext := scheduler.Stop()
	<-stopContext.Done()
}
