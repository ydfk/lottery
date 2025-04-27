package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"lottery-backend/internal/config"
	"lottery-backend/internal/pkg/logger"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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

	// 根据配置决定是否使用代理
	if cfg.UseProxy && cfg.ProxyAddress != "" {
		proxyUrl, err := url.Parse(cfg.ProxyAddress)
		if err != nil {
			logger.Error("解析代理地址失败：%v", err)
		} else {
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			clientConfig.HTTPClient = &http.Client{
				Transport: transport,
			}
			logger.Info("已启用代理: %s", cfg.ProxyAddress)
		}
	}

	client := openai.NewClientWithConfig(clientConfig)
	return &Client{
		client: client,
		config: cfg,
	}
}

// logRequest 记录请求参数
func (c *Client) logRequest(req interface{}) {
	if reqBytes, err := json.MarshalIndent(req, "", "  "); err == nil {
		logger.Debug("API请求参数: %s", string(reqBytes))
	} else {
		logger.Error("序列化请求参数失败: %v", err)
	}
}

// logResponse 记录响应结果
func (c *Client) logResponse(resp interface{}) {
	if respBytes, err := json.MarshalIndent(resp, "", "  "); err == nil {
		logger.Debug("API响应结果: %s", string(respBytes))
	} else {
		logger.Error("序列化响应结果失败: %v", err)
	}
}

// validateNumberRange 验证号码数值范围
func validateNumberRange(numbers string, lotteryType string) bool {
	// 分离前后区号码
	parts := strings.Split(numbers, "+")
	if len(parts) != 2 {
		return false
	}

	if lotteryType == "fc_ssq" {
		// 验证红球范围(01-33)
		redBalls := strings.Split(parts[0], ",")
		if len(redBalls) != 6 {
			return false
		}
		for _, ball := range redBalls {
			num, err := strconv.Atoi(ball)
			if err != nil || num < 1 || num > 33 {
				return false
			}
		}
		// 验证蓝球范围(01-16)
		blueBall, err := strconv.Atoi(parts[1])
		if err != nil || blueBall < 1 || blueBall > 16 {
			return false
		}
	} else if lotteryType == "tc_dlt" {
		// 验证前区范围(01-35)
		frontArea := strings.Split(parts[0], ",")
		if len(frontArea) != 5 {
			return false
		}
		for _, ball := range frontArea {
			num, err := strconv.Atoi(ball)
			if err != nil || num < 1 || num > 35 {
				return false
			}
		}
		// 验证后区范围(01-12)
		backArea := strings.Split(parts[1], ",")
		if len(backArea) != 2 {
			return false
		}
		for _, ball := range backArea {
			num, err := strconv.Atoi(ball)
			if err != nil || num < 1 || num > 12 {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

// validateNoDuplicates 验证号码是否有重复
func validateNoDuplicates(numbers string) bool {
	parts := strings.Split(numbers, "+")
	if len(parts) != 2 {
		return false
	}

	// 检查前区号码是否重复
	frontNumbers := strings.Split(parts[0], ",")
	frontSet := make(map[string]bool)
	for _, num := range frontNumbers {
		if frontSet[num] {
			return false
		}
		frontSet[num] = true
	}

	// 检查后区号码是否重复（仅大乐透需要）
	if strings.Contains(parts[1], ",") {
		backNumbers := strings.Split(parts[1], ",")
		backSet := make(map[string]bool)
		for _, num := range backNumbers {
			if backSet[num] {
				return false
			}
			backSet[num] = true
		}
	}

	return true
}

// GenerateMultipleLotteryNumbers 生成多组彩票号码
func (c *Client) GenerateMultipleLotteryNumbers(ctx context.Context, lotteryType string, model string, count int) ([]string, error) {
	if count <= 0 {
		count = 1 // 确保至少生成一组号码
	}

	logger.Info("开始为彩票类型[%s]生成%d组号码，使用模型：%s", lotteryType, count, model)

	// 用于存储生成的所有号码
	numbers := make([]string, 0, count)

	// 如果只需要生成一组，直接调用单个生成方法
	if count == 1 {
		number, err := c.GenerateLotteryNumbers(ctx, lotteryType, model)
		if err != nil {
			return nil, err
		}
		return []string{number}, nil
	}

	// 使用批量生成模式，一次性请求多组号码
	systemPrompt := fmt.Sprintf(`你是一个专业的彩票号码生成器。请严格按照以下要求生成%d组彩票号码：

1. 必须使用这个格式输出每一组号码：<NUMBER>号码</NUMBER>
2. 严格按照下面的格式规范生成号码：
   - fc_ssq：6个红球(01-33)+1个蓝球(01-16)，格式如 01,05,13,22,29,33+07
   - tc_dlt：5个前区(01-35)+2个后区(01-12)，格式如 03,05,18,27,34+08,11
3. 所有数字必须按从小到大排序
4. 所有数字必须补零，保持两位数格式
5. 不要输出任何额外的解释文字，只需要输出用<NUMBER>标签包裹的号码
6. 确保生成的每一组号码都是有效且符合规则的随机组合
7. 一共必须生成%d组不同的号码`, count, count)

	userPrompt := fmt.Sprintf("请生成%d组不同的%s彩票号码，每组号码必须用<NUMBER></NUMBER>标签包裹", count, lotteryType)

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.9, // 增加创造性，确保生成的号码组彼此不同
		TopP:        1,
		Stream:      false,
	}

	// 记录请求参数
	c.logRequest(req)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout*2) // 给批量生成多一点时间
	defer cancel()

	// 尝试获取批量号码
	start := time.Now()
	resp, err := c.client.CreateChatCompletion(ctx, req)
	duration := time.Since(start)

	if err == nil && len(resp.Choices) > 0 {
		// 记录响应
		c.logResponse(resp)

		result := resp.Choices[0].Message.Content
		logger.Info("AI批量生成原始返回 (耗时:%v): %s", duration, result)

		// 提取所有<NUMBER>标签中的内容
		numberPattern := `<NUMBER>(.*?)</NUMBER>`
		re := regexp.MustCompile(numberPattern)
		matches := re.FindAllStringSubmatch(result, -1)

		// 验证并添加有效的号码
		for _, match := range matches {
			if len(match) >= 2 {
				extractedNumber := strings.TrimSpace(match[1])

				// 验证格式
				var formatValid bool
				if lotteryType == "fc_ssq" {
					ssbPattern := `^(\d{2},){5}\d{2}\+\d{2}$`
					ssbRe := regexp.MustCompile(ssbPattern)
					formatValid = ssbRe.MatchString(extractedNumber)
				} else if lotteryType == "tc_dlt" {
					dltPattern := `^(\d{2},){4}\d{2}\+\d{2},\d{2}$`
					dltRe := regexp.MustCompile(dltPattern)
					formatValid = dltRe.MatchString(extractedNumber)
				}

				// 验证数值范围和重复性
				if formatValid && validateNoDuplicates(extractedNumber) && validateNumberRange(extractedNumber, lotteryType) {
					numbers = append(numbers, extractedNumber)
					logger.Info("成功提取并验证号码组 %d/%d: %s", len(numbers), count, extractedNumber)
				} else {
					logger.Error("提取到的号码无效: %s", extractedNumber)
				}
			}
		}
	} else {
		logger.Error("批量生成号码失败: %v", err)
	}

	// 如果批量生成失败或数量不足，使用单个生成方法补足
	if len(numbers) < count {
		logger.Info("批量生成不足%d组，已获取%d组，使用单个生成方法补足", count, len(numbers))

		// 创建一个号码集合，避免重复
		numbersSet := make(map[string]bool)
		for _, num := range numbers {
			numbersSet[num] = true
		}

		// 一个个补充生成
		for len(numbers) < count {
			number, err := c.GenerateLotteryNumbers(ctx, lotteryType, model)
			if err != nil {
				logger.Error("单个生成号码失败: %v", err)
				// 如果已经有一些号码，就返回已有的，否则返回错误
				if len(numbers) > 0 {
					logger.Info("虽然未能生成足够数量，但返回已生成的%d组号码", len(numbers))
					return numbers, nil
				}
				return nil, err
			}

			// 确保没有重复
			if !numbersSet[number] {
				numbers = append(numbers, number)
				numbersSet[number] = true
				logger.Info("成功生成补充号码 %d/%d: %s", len(numbers), count, number)
			}
		}
	}

	logger.Info("成功为%s彩票生成%d组号码", lotteryType, len(numbers))
	return numbers, nil
}

// GenerateLotteryNumbers 生成彩票号码
func (c *Client) GenerateLotteryNumbers(ctx context.Context, lotteryType string, model string) (string, error) {
	logger.Info("开始为彩票类型[%s]生成号码，使用模型：%s", lotteryType, model)

	// 构建带有特殊标记格式的提示词
	systemPrompt := `你是一个专业的彩票号码生成器。请严格按照以下要求生成号码：

1. 必须使用这个格式输出：<NUMBER>你生成的号码</NUMBER>
2. 严格按照下面的格式规范生成号码：
   - fc_ssq：6个红球(01-33)+1个蓝球(01-16)，格式如 01,05,13,22,29,33+07
   - tc_dlt：5个前区(01-35)+2个后区(01-12)，格式如 03,05,18,27,34+08,11
3. 所有数字必须按从小到大排序
4. 所有数字必须补零，保持两位数格式
5. 除了<NUMBER>标签内的内容外，不要输出任何其他文字
6. 确保生成的号码是有效且符合规则的随机组合`

	userPrompt := fmt.Sprintf("请生成一注%s彩票号码，务必包含在<NUMBER></NUMBER>标签内", lotteryType)

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.7, // 平衡创造性和一致性
		TopP:        1,   // 不限制词汇选择范围
		Stream:      false,
	}

	// 记录请求参数
	c.logRequest(req)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// 尝试多次调用
	var lastErr error
	for i := 0; i < c.config.MaxRetries; i++ {
		logger.Info("尝试调用API（第%d次）...", i+1)
		start := time.Now()

		resp, err := c.client.CreateChatCompletion(ctx, req)
		duration := time.Since(start)

		if err == nil && len(resp.Choices) > 0 {
			// 记录完整的响应结果
			c.logResponse(resp)

			result := resp.Choices[0].Message.Content
			logger.Info("AI原始返回: %s", result)

			// 使用正则表达式提取<NUMBER>标签中的内容
			numberPattern := `<NUMBER>(.*?)</NUMBER>`
			re := regexp.MustCompile(numberPattern)
			matches := re.FindStringSubmatch(result)

			if len(matches) >= 2 {
				extractedNumber := strings.TrimSpace(matches[1])
				logger.Info("成功提取号码: %s", extractedNumber)

				// 先验证格式
				var formatValid bool
				if lotteryType == "fc_ssq" {
					ssbPattern := `^(\d{2},){5}\d{2}\+\d{2}$`
					ssbRe := regexp.MustCompile(ssbPattern)
					formatValid = ssbRe.MatchString(extractedNumber)
				} else if lotteryType == "tc_dlt" {
					dltPattern := `^(\d{2},){4}\d{2}\+\d{2},\d{2}$`
					dltRe := regexp.MustCompile(dltPattern)
					formatValid = dltRe.MatchString(extractedNumber)
				}

				// 如果格式正确，继续验证数值范围
				if formatValid {
					if !validateNoDuplicates(extractedNumber) {
						logger.Error("%s号码存在重复数字: %s", lotteryType, extractedNumber)
						continue
					}

					if validateNumberRange(extractedNumber, lotteryType) {
						logger.Info("%s号码格式、重复性和范围验证都通过: %s, 耗时：%v", lotteryType, extractedNumber, duration)
						return extractedNumber, nil
					} else {
						logger.Error("%s号码数值范围验证失败: %s", lotteryType, extractedNumber)
					}
				} else {
					logger.Error("%s号码格式验证失败: %s", lotteryType, extractedNumber)
				}
			} else {
				logger.Error("未能从AI回复中提取出号码标记: %s", result)
			}
		} else {
			lastErr = err
			logger.Error("API调用失败（第%d次）：%v，耗时：%v", i+1, err, duration)
			if err != nil {
				// 记录错误详情
				logger.Debug("API错误详情: %+v", err)
			}
		}

		retryDelay := time.Second * time.Duration(i+1)
		logger.Info("等待%v后重试...", retryDelay)
		time.Sleep(retryDelay)
	}

	logger.Error("生成号码最终失败，已重试%d次，最后错误：%v", c.config.MaxRetries, lastErr)
	return "", fmt.Errorf("生成号码失败（重试%d次）: %v", c.config.MaxRetries, lastErr)
}
