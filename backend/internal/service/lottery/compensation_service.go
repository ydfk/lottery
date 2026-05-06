package lottery

import (
	"context"
	"fmt"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"
)

const compensationTaskDrawPrize = "drawPrize"

type compensationTask func(context.Context, config.CompensationJobConfig) error

var compensationTaskRegistry = map[string]compensationTask{
	compensationTaskDrawPrize: compensatePreviousDrawPrizes,
}

func RunCompensationJob(ctx context.Context, job config.CompensationJobConfig) error {
	task, ok := compensationTaskRegistry[job.Type]
	if !ok {
		return fmt.Errorf("未知补偿任务类型: %s", job.Type)
	}
	return task(ctx, job)
}

func compensatePreviousDrawPrizes(ctx context.Context, job config.CompensationJobConfig) error {
	offsetDays := job.TargetDateOffsetDays
	if offsetDays <= 0 {
		offsetDays = 1
	}

	targetDate := time.Now().AddDate(0, 0, -offsetDays)
	for _, definition := range ListDefinitions() {
		if !definition.Enabled {
			continue
		}
		if err := compensateDefinitionDrawPrize(ctx, definition, targetDate); err != nil {
			logger.Warn("补全 %s 开奖奖级失败: %v", definition.Code, err)
		}
	}
	return nil
}

func compensateDefinitionDrawPrize(ctx context.Context, definition Definition, targetDate time.Time) error {
	issue, shouldCheck, err := resolveCompensationIssue(definition, targetDate)
	if err != nil || !shouldCheck {
		return err
	}
	if hasDrawPrizeDetails(definition.Code, issue) {
		return nil
	}

	result, err := SyncDrawIssue(ctx, definition.Code, issue)
	if err != nil {
		return err
	}
	if result != nil && result.SyncedCount > 0 {
		logger.Info("已补全 %s 第 %s 期开奖数据", definition.Code, issue)
		return nil
	}
	logger.Info("已检查 %s 第 %s 期开奖数据，第三方暂未返回可补全内容", definition.Code, issue)
	return nil
}

func resolveCompensationIssue(definition Definition, targetDate time.Time) (string, bool, error) {
	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return "", false, err
	}

	drawAt := schedule.atDate(targetDate)
	if !schedule.matches(drawAt) {
		return "", false, nil
	}

	anchor, ok, err := parseConfiguredAnchor(definition, schedule)
	if err != nil || !ok {
		return "", false, err
	}

	issue, err := resolveIssueByDrawAt(anchor, drawAt, schedule)
	if err != nil {
		return "", false, err
	}
	return issue, true, nil
}

func hasDrawPrizeDetails(code string, issue string) bool {
	var count int64
	err := db.DB.Model(&model.DrawPrize{}).
		Joins("JOIN draw_results ON draw_results.id = draw_prizes.draw_result_id").
		Where("draw_results.lottery_code = ? AND draw_results.issue IN ?", code, issueAliases(code, issue)).
		Count(&count).Error
	return err == nil && count > 0
}
