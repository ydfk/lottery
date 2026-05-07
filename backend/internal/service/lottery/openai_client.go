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
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
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

	startedAt := time.Now()
	endpoint := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}

	requestBody := openAIRequest{
		Model:    model,
		Messages: messages,
	}
	rawRequest, err := json.Marshal(requestBody)
	if err != nil {
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), nil, 0, startedAt, err)
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawRequest))
	if err != nil {
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), nil, 0, startedAt, err)
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), nil, 0, startedAt, err)
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), nil, response.StatusCode, startedAt, err)
		return "", err
	}
	if response.StatusCode >= http.StatusBadRequest {
		requestErr := fmt.Errorf("模型请求失败: %s", string(body))
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), body, response.StatusCode, startedAt, requestErr)
		return "", requestErr
	}

	parsed := openAIResponse{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), body, response.StatusCode, startedAt, err)
		return "", err
	}
	if len(parsed.Choices) == 0 {
		requestErr := fmt.Errorf("模型未返回内容")
		logThirdPartyFailure("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), body, response.StatusCode, startedAt, requestErr)
		return "", requestErr
	}

	logThirdPartySuccess("openai-compatible", http.MethodPost, endpoint, buildOpenAIRequestLog(requestBody), body, response.StatusCode, startedAt)
	return stripJSONFence(parsed.Choices[0].Message.Content), nil
}

func buildOpenAIRequestLog(request openAIRequest) map[string]any {
	return map[string]any{
		"model":    request.Model,
		"messages": sanitizeOpenAIMessages(request.Messages),
	}
}

func sanitizeOpenAIMessages(messages []openAIMessage) []map[string]any {
	result := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		item := map[string]any{
			"role": message.Role,
		}

		switch content := message.Content.(type) {
		case string:
			item["content"] = truncateLogValue(content)
		case []map[string]any:
			item["content"] = sanitizeOpenAIContentItems(content)
		default:
			item["content"] = truncateLogValue(mustJSON(content))
		}
		result = append(result, item)
	}
	return result
}

func sanitizeOpenAIContentItems(items []map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		cloned := make(map[string]any, len(item))
		for key, value := range item {
			cloned[key] = value
		}

		itemType, _ := cloned["type"].(string)
		if itemType == "image_url" {
			cloned["image_url"] = map[string]string{"url": "<base64 omitted>"}
		}
		if itemType == "text" {
			if text, ok := cloned["text"].(string); ok {
				cloned["text"] = truncateLogValue(text)
			}
		}
		result = append(result, cloned)
	}
	return result
}
