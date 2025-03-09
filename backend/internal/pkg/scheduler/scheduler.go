package scheduler

import (
	"context"
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/ai"
	"lottery-backend/internal/pkg/database"
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		s.cron.Stop()
	}

	s.cron = cron.New()
	s.entries = make(map[uint]cron.EntryID)

	// 加载所有活跃的彩票类型
	var types []models.LotteryType
	if err := database.DB.Where("is_active = ?", true).Find(&types).Error; err != nil {
		return fmt.Errorf("加载彩票类型失败: %v", err)
	}

	// 为每个彩票类型创建生成任务
	for _, t := range types {
		if err := s.addLotteryTask(t); err != nil {
			return fmt.Errorf("添加任务失败[%s]: %v", t.Name, err)
		}
	}

	// 添加开奖结果爬取任务
	if err := s.addDrawResultTask(); err != nil {
		return fmt.Errorf("添加开奖结果任务失败: %v", err)
	}

	s.cron.Start()
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		s.cron.Stop()
	}
}

// addLotteryTask 添加彩票号码生成任务
func (s *Scheduler) addLotteryTask(lt models.LotteryType) error {
	entryID, err := s.cron.AddFunc(lt.ScheduleCron, func() {
		ctx := context.Background()

		// 生成号码
		numbers, err := s.aiClient.GenerateLotteryNumbers(ctx, lt.Name)
		if err != nil {
			fmt.Printf("生成%s号码失败: %v\n", lt.Name, err)
			return
		}

		// 计算开奖时间（这里需要根据具体彩种规则计算）
		drawTime := s.calculateNextDrawTime(lt.Name)

		// 保存推荐记录
		recommendation := models.Recommendation{
			LotteryTypeID: lt.ID,
			Numbers:       numbers,
			ModelName:     lt.ModelName,
			DrawTime:      drawTime,
		}

		if err := database.DB.Create(&recommendation).Error; err != nil {
			fmt.Printf("保存%s推荐号码失败: %v\n", lt.Name, err)
		}
	})

	if err != nil {
		return err
	}

	s.entries[lt.ID] = entryID
	return nil
}

// addDrawResultTask 添加开奖结果爬取任务
func (s *Scheduler) addDrawResultTask() error {
	// 每天22:00执行开奖结果爬取
	_, err := s.cron.AddFunc("0 22 * * *", func() {
		// TODO: 实现开奖结果爬取逻辑
		fmt.Println("开始爬取开奖结果...")
	})
	return err
}

// ReloadTask 重新加载指定彩票类型的任务
func (s *Scheduler) ReloadTask(lt models.LotteryType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果存在旧任务，先移除
	if oldEntryID, exists := s.entries[lt.ID]; exists {
		s.cron.Remove(oldEntryID)
		delete(s.entries, lt.ID)
	}

	// 如果彩票类型处于活跃状态，添加新任务
	if lt.IsActive {
		return s.addLotteryTask(lt)
	}

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
