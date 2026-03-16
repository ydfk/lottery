package lottery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go-fiber-starter/pkg/config"
)

type openAIMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callRecommendationModel(ctx context.Context, model string, prompt string) (string, error) {
	return callOpenAICompatible(
		ctx,
		config.Current.AI.BaseURL,
		config.Current.AI.APIKey,
		model,
		time.Duration(max(10, config.Current.AI.TimeoutSeconds))*time.Second,
		[]openAIMessage{
			{
				Role:    "system",
				Content: "你是彩票推荐助手。你必须返回严格 JSON，不要输出解释文字。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	)
}

func callVisionModel(ctx context.Context, model string, prompt string, imagePath string) (string, error) {
	rawImage, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	return callOpenAICompatible(
		ctx,
		config.Current.Vision.BaseURL,
		config.Current.Vision.APIKey,
		model,
		time.Duration(max(10, config.Current.Vision.TimeoutSeconds))*time.Second,
		[]openAIMessage{
			{
				Role:    "system",
				Content: "你是彩票 OCR 助手。你必须返回严格 JSON，不要输出解释文字。",
			},
			{
				Role: "user",
				Content: []map[string]any{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(rawImage),
						},
					},
				},
			},
		},
	)
}

func callOpenAICompatible(ctx context.Context, baseURL string, apiKey string, model string, timeout time.Duration, messages []openAIMessage) (string, error) {
	if baseURL == "" || apiKey == "" || model == "" {
		return "", fmt.Errorf("未配置 OpenAI 兼容模型")
	}

	endpoint := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}

	requestBody := openAIRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
	}
	rawRequest, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawRequest))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("模型请求失败: %s", string(body))
	}

	parsed := openAIResponse{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("模型未返回内容")
	}

	return stripJSONFence(parsed.Choices[0].Message.Content), nil
}
