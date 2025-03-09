package ai

import (
	"context"
	"fmt"
	"lottery-backend/internal/pkg/config"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	config config.AIConfig
}

func NewClient() *Client {
	cfg := config.Current.AI
	clientConfig := openai.DefaultConfig(cfg.APIKey)
	clientConfig.BaseURL = cfg.BaseURL

	client := openai.NewClientWithConfig(clientConfig)
	return &Client{
		client: client,
		config: cfg,
	}
}

// GenerateLotteryNumbers 生成彩票号码
func (c *Client) GenerateLotteryNumbers(ctx context.Context, lotteryType string) (string, error) {
	// 构建提示词
	prompt := fmt.Sprintf(`作为一个资深的彩票数字生成专家，请为%s生成一组号码。
要求：
1. 生成的数字要符合该彩票类型的规则
2. 考虑历史数据和概率分布
3. 只返回号码本身，不要加任何额外说明

示例返回格式：
双色球：01,15,17,19,27,32+06
大乐透：05,11,17,21,34+02,09`, lotteryType)

	req := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo", // 可以根据配置切换不同模型
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "你是一个专业的彩票号码生成专家，请只返回生成的号码，不要有任何多余的解释。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   50,
		Temperature: 0.7,
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// 尝试多次调用
	var lastErr error
	for i := 0; i < c.config.MaxRetries; i++ {
		resp, err := c.client.CreateChatCompletion(ctx, req)
		if err == nil && len(resp.Choices) > 0 {
			return resp.Choices[0].Message.Content, nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
	}

	return "", fmt.Errorf("生成号码失败（重试%d次）: %v", c.config.MaxRetries, lastErr)
}
