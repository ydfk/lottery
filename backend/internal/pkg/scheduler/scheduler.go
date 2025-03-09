package scheduler

import (
	"context"
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/ai"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	mu       sync.Mutex
	cron     *cron.Cron
	entries  map[uint]cron.EntryID
	aiClient *ai.Client
}

func NewScheduler(aiClient *ai.Client) *Scheduler {
	return &Scheduler{
		entries:  make(map[uint]cron.EntryID),
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
	s.entries = make(map[uint]cron.EntryID)

	// 加载所有活跃的彩票类型
	logger.Info("加载活跃彩票类型...")
	var types []models.LotteryType
	if err := database.DB.Where("is_active = ?", true).Find(&types).Error; err != nil {
		logger.Error("加载彩票类型失败: %v", err)
		return fmt.Errorf("加载彩票类型失败: %v", err)
	}
	logger.Info("找到%d个活跃彩票类型", len(types))

	// 为每个彩票类型创建生成任务
	for _, t := range types {
		logger.Info("正在为彩票类型[%s]添加任务...", t.Name)
		if err := s.addLotteryTask(t); err != nil {
			logger.Error("添加任务失败[%s]: %v", t.Name, err)
			return fmt.Errorf("添加任务失败[%s]: %v", t.Name, err)
		}
		logger.Info("成功添加彩票类型[%s]的任务", t.Name)
	}

	// 添加开奖结果爬取任务
	logger.Info("添加开奖结果爬取任务...")
	if err := s.addDrawResultTask(); err != nil {
		logger.Error("添加开奖结果任务失败: %v", err)
		return fmt.Errorf("添加开奖结果任务失败: %v", err)
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

// addLotteryTask 添加彩票号码生成任务
func (s *Scheduler) addLotteryTask(lt models.LotteryType) error {
	logger.Info("正在添加彩票类型[%s]的号码生成任务，Cron表达式：%s", lt.Name, lt.ScheduleCron)
	entryID, err := s.cron.AddFunc(lt.ScheduleCron, func() {
		ctx := context.Background()
		logger.Info("开始执行彩票类型[%s]的号码生成任务...", lt.Name)

		// 生成号码
		numbers, err := s.aiClient.GenerateLotteryNumbers(ctx, lt.Name, lt.ModelName)
		if err != nil {
			logger.Error("生成%s号码失败: %v", lt.Name, err)
			return
		}
		logger.Info("成功生成%s号码: %s", lt.Name, numbers)

		// 计算开奖时间（这里需要根据具体彩种规则计算）
		drawTime := s.calculateNextDrawTime(lt.Name)
		logger.Info("计算得到下次开奖时间: %v", drawTime)

		// 保存推荐记录
		recommendation := models.Recommendation{
			LotteryTypeID: lt.ID,
			Numbers:       numbers,
			ModelName:     lt.ModelName,
			DrawTime:      drawTime,
		}

		if err := database.DB.Create(&recommendation).Error; err != nil {
			logger.Error("保存%s推荐号码失败: %v", lt.Name, err)
		} else {
			logger.Info("成功保存%s推荐号码，ID：%d", lt.Name, recommendation.ID)
		}
	})

	if err != nil {
		logger.Error("添加彩票类型[%s]的任务失败: %v", lt.Name, err)
		return err
	}

	s.entries[lt.ID] = entryID
	logger.Info("成功添加彩票类型[%s]的任务", lt.Name)
	return nil
}

// addDrawResultTask 添加开奖结果爬取任务
func (s *Scheduler) addDrawResultTask() error {
	logger.Info("正在添加开奖结果爬取任务...")
	// 每天22:00执行开奖结果爬取
	_, err := s.cron.AddFunc("0 22 * * *", func() {
		logger.Info("开始执行开奖结果爬取任务...")
		// TODO: 实现开奖结果爬取逻辑
	})
	if err != nil {
		logger.Error("添加开奖结果爬取任务失败: %v", err)
		return err
	}
	logger.Info("成功添加开奖结果爬取任务")
	return nil
}

// ReloadTask 重新加载指定彩票类型的任务
func (s *Scheduler) ReloadTask(lt models.LotteryType) error {
	logger.Info("开始重新加载彩票类型[%s]的任务...", lt.Name)
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果存在旧任务，先移除
	if oldEntryID, exists := s.entries[lt.ID]; exists {
		logger.Info("移除彩票类型[%s]的旧任务...", lt.Name)
		s.cron.Remove(oldEntryID)
		delete(s.entries, lt.ID)
	}

	// 如果彩票类型处于活跃状态，添加新任务
	if lt.IsActive {
		logger.Info("彩票类型[%s]处于活跃状态，添加新任务...", lt.Name)
		return s.addLotteryTask(lt)
	}

	logger.Info("彩票类型[%s]处于非活跃状态，跳过添加任务", lt.Name)
	return nil
}

// calculateNextDrawTime 计算下次开奖时间
func (s *Scheduler) calculateNextDrawTime(lotteryType string) time.Time {
	// TODO: 根据不同彩种的开奖规则计算下次开奖时间
	// 这里使用简单的示例实现
	return time.Now().Add(24 * time.Hour)
}

// ValidateCron 验证cron表达式
func ValidateCron(cronStr string) error {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronStr)
	return err
}
