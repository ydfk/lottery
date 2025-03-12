package scheduler

import (
	"context"
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/ai"
	"lottery-backend/internal/pkg/config"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/draw"
	"lottery-backend/internal/pkg/logger"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	mu       sync.Mutex
	cron     *cron.Cron
	entries  map[string]cron.EntryID // 使用taskType+ID作为key
	aiClient *ai.Client
}

const (
	TaskTypeGenerate = "generate" // 推荐号码生成任务
	TaskTypeFetch    = "fetch"    // 开奖结果爬取任务
	TaskTypeAnalyze  = "analyze"  // 中奖分析任务
)

func NewScheduler(aiClient *ai.Client) *Scheduler {
	return &Scheduler{
		entries:  make(map[string]cron.EntryID),
		aiClient: aiClient,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	logger.Info("开始启动调度器...")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		logger.Info("停止现有调度器...")
		s.cron.Stop()
	}

	s.cron = cron.New()
	s.entries = make(map[string]cron.EntryID)

	// 加载所有活跃的彩票类型
	logger.Info("加载活跃彩票类型...")
	var types []models.LotteryType
	if err := database.DB.Where("is_active = ?", true).Find(&types).Error; err != nil {
		logger.Error("加载彩票类型失败: %v", err)
		return fmt.Errorf("加载彩票类型失败: %v", err)
	}
	logger.Info("找到%d个活跃彩票类型", len(types))

	// 为每个彩票类型创建任务
	for _, t := range types {
		logger.Info("正在为彩票类型[%s]添加任务...", t.Name)

		// 添加号码推荐任务
		if err := s.addLotteryGenerationTask(t); err != nil {
			logger.Error("添加号码推荐任务失败[%s]: %v", t.Name, err)
			return fmt.Errorf("添加号码推荐任务失败[%s]: %v", t.Name, err)
		}
		logger.Info("成功添加彩票类型[%s]的号码推荐任务", t.Name)
	}

	// 添加每日开奖结果爬取任务（所有彩票类型共用一个任务）
	logger.Info("添加开奖结果爬取任务...")
	if err := s.addDrawResultFetchTask(); err != nil {
		logger.Error("添加开奖结果爬取任务失败: %v", err)
		return fmt.Errorf("添加开奖结果爬取任务失败: %v", err)
	}
	logger.Info("成功添加开奖结果爬取任务")

	s.cron.Start()
	logger.Info("调度器启动完成")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	logger.Info("正在停止调度器...")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		s.cron.Stop()
		logger.Info("调度器已停止")
	}
}

// addLotteryGenerationTask 添加彩票号码生成任务
func (s *Scheduler) addLotteryGenerationTask(lt models.LotteryType) error {
	logger.Info("正在添加彩票类型[%s]的号码生成任务，Cron表达式：%s", lt.Code, lt.ScheduleCron)
	taskKey := fmt.Sprintf("%s_%d", TaskTypeGenerate, lt.ID)

	entryID, err := s.cron.AddFunc(lt.ScheduleCron, func() {
		ctx := context.Background()
		logger.Info("开始执行彩票类型[%s]的号码生成任务...", lt.Code)

		// 获取开奖信息（日期和期号）
		drawInfo, err := draw.GetLotteryDrawInfo(lt.Code, lt.ScheduleCron)
		if err != nil {
			logger.Error("获取%s开奖信息失败: %v", lt.Code, err)
			return
		}
		logger.Info("获取到%s下一期信息: 日期=%v, 期号=%s", lt.Code, drawInfo.NextDrawDate, drawInfo.NextDrawNum)

		// 使用AI客户端生成号码
		numbers, err := s.aiClient.GenerateLotteryNumbers(ctx, lt.Code, lt.ModelName)
		if err != nil {
			logger.Error("生成%s号码失败: %v", lt.Code, err)
			return
		}
		logger.Info("成功生成%s号码: %s", lt.Code, numbers)

		// 保存推荐记录
		recommendation := models.Recommendation{
			LotteryTypeID:    lt.ID,
			Numbers:          numbers,
			ModelName:        lt.ModelName,
			ExpectedDrawTime: drawInfo.NextDrawDate, // 使用预计开奖时间
			DrawNumber:       drawInfo.NextDrawNum,
		}

		if err := database.DB.Create(&recommendation).Error; err != nil {
			logger.Error("保存%s推荐号码失败: %v", lt.Code, err)
		} else {
			logger.Info("成功保存%s推荐号码，ID：%d, 期号：%s", lt.Code, recommendation.ID, recommendation.DrawNumber)
		}
	})

	if err != nil {
		logger.Error("添加彩票类型[%s]的号码生成任务失败: %v", lt.Code, err)
		return err
	}

	s.entries[taskKey] = entryID
	logger.Info("成功添加彩票类型[%s]的号码生成任务", lt.Code)
	return nil
}

// addDrawResultFetchTask 添加开奖结果爬取任务
func (s *Scheduler) addDrawResultFetchTask() error {
	logger.Info("正在添加开奖结果爬取任务...")
	taskKey := "fetch_all_results"

	// 从配置文件中读取开奖结果爬取时间的cron表达式
	resultFetchCron := config.Current.Scheduler.ResultFetchCron
	logger.Info("从配置文件读取到开奖结果爬取时间配置: %s", resultFetchCron)

	// 按照配置的cron表达式执行开奖结果爬取
	entryID, err := s.cron.AddFunc(resultFetchCron, func() {
		logger.Info("开始执行定时开奖结果爬取任务")

		// 爬取所有活跃彩票类型的最新开奖结果
		if err := draw.FetchAllActiveLotteryDrawResults(); err != nil {
			logger.Error("爬取开奖结果失败: %v", err)
			return
		}

		logger.Info("定时开奖结果爬取任务执行完成")
	})

	if err != nil {
		logger.Error("添加开奖结果爬取任务失败: %v", err)
		return err
	}

	s.entries[taskKey] = entryID
	logger.Info("成功添加开奖结果爬取任务，执行时间: %s", resultFetchCron)
	return nil
}

// ReloadTask 重新加载指定彩票类型的任务
func (s *Scheduler) ReloadTask(lt models.LotteryType) error {
	logger.Info("开始重新加载彩票类型[%s]的任务...", lt.Name)
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成任务键
	genTaskKey := fmt.Sprintf("%s_%d", TaskTypeGenerate, lt.ID)

	// 如果存在旧的号码生成任务，先移除
	if oldEntryID, exists := s.entries[genTaskKey]; exists {
		logger.Info("移除彩票类型[%s]的旧号码生成任务...", lt.Name)
		s.cron.Remove(oldEntryID)
		delete(s.entries, genTaskKey)
	}

	// 如果彩票类型处于活跃状态，添加新任务
	if lt.IsActive {
		logger.Info("彩票类型[%s]处于活跃状态，添加新任务...", lt.Name)
		return s.addLotteryGenerationTask(lt)
	}

	logger.Info("彩票类型[%s]处于非活跃状态，跳过添加任务", lt.Name)
	return nil
}

// ManualFetchLotteryResult 手动触发特定彩票类型的开奖结果爬取
func (s *Scheduler) ManualFetchLotteryResult(lotteryTypeID uint) error {
	logger.Info("手动触发彩票类型[ID:%d]的开奖结果爬取...", lotteryTypeID)

	// 查询彩票类型
	var lotteryType models.LotteryType
	if err := database.DB.First(&lotteryType, lotteryTypeID).Error; err != nil {
		return fmt.Errorf("彩票类型不存在: %v", err)
	}

	// 检查是否配置了彩票ID
	if lotteryType.CaipiaoID <= 0 {
		return fmt.Errorf("彩票类型[%s]未配置彩票ID", lotteryType.Name)
	}

	// 获取最新开奖结果
	drawResult, err := draw.FetchLatestDrawResult(&lotteryType)
	if err != nil {
		return fmt.Errorf("获取彩票[%s]开奖结果失败: %v", lotteryType.Name, err)
	}

	// 处理开奖结果(保存结果并分析中奖情况)
	if err := draw.ProcessDrawResult(drawResult); err != nil {
		return fmt.Errorf("处理彩票[%s]开奖结果失败: %v", lotteryType.Name, err)
	}

	logger.Info("彩票类型[%s]手动开奖结果爬取完成", lotteryType.Name)
	return nil
}

// ManualGenerateLotteryNumbers 手动触发彩票号码生成
func (s *Scheduler) ManualGenerateLotteryNumbers(lotteryTypeID uint) (string, error) {
	logger.Info("手动触发彩票类型[ID:%d]的号码生成...", lotteryTypeID)

	// 查询彩票类型
	var lotteryType models.LotteryType
	if err := database.DB.First(&lotteryType, lotteryTypeID).Error; err != nil {
		return "", fmt.Errorf("彩票类型不存在: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取开奖信息（日期和期号）
	drawInfo, err := draw.GetLotteryDrawInfo(lotteryType.Code, lotteryType.ScheduleCron)
	if err != nil {
		return "", fmt.Errorf("获取开奖信息失败: %v", err)
	}

	// 使用AI生成号码
	numbers, err := s.aiClient.GenerateLotteryNumbers(ctx, lotteryType.Code, lotteryType.ModelName)
	if err != nil {
		return "", fmt.Errorf("生成号码失败: %v", err)
	}

	// 保存推荐记录
	recommendation := models.Recommendation{
		LotteryTypeID:    lotteryType.ID,
		Numbers:          numbers,
		ModelName:        lotteryType.ModelName,
		ExpectedDrawTime: drawInfo.NextDrawDate,
		DrawNumber:       drawInfo.NextDrawNum,
	}

	if err := database.DB.Create(&recommendation).Error; err != nil {
		return "", fmt.Errorf("保存推荐记录失败: %v", err)
	}

	logger.Info("手动生成彩票[%s]号码成功: %s, 期号: %s", lotteryType.Name, numbers, drawInfo.NextDrawNum)
	return numbers, nil
}

// ValidateCron 验证cron表达式
func ValidateCron(cronStr string) error {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronStr)
	return err
}
