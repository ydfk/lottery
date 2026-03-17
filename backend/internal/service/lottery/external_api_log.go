package lottery

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"go-fiber-starter/pkg/logger"
)

const externalAPILogLimit = 4000

func logThirdPartySuccess(provider string, method string, endpoint string, request any, responseBody []byte, statusCode int, startedAt time.Time) {
	payload := buildThirdPartyLogPayload(provider, method, endpoint, request, responseBody, statusCode, startedAt, "")
	logger.Info("第三方接口调用: %s", mustJSON(payload))
}

func logThirdPartyFailure(provider string, method string, endpoint string, request any, responseBody []byte, statusCode int, startedAt time.Time, err error) {
	message := ""
	if err != nil {
		message = err.Error()
	}
	payload := buildThirdPartyLogPayload(provider, method, endpoint, request, responseBody, statusCode, startedAt, message)
	logger.Error("第三方接口调用失败: %s", mustJSON(payload))
}

func buildThirdPartyLogPayload(provider string, method string, endpoint string, request any, responseBody []byte, statusCode int, startedAt time.Time, errorMessage string) map[string]any {
	payload := map[string]any{
		"provider":   provider,
		"method":     method,
		"endpoint":   maskURL(endpoint),
		"request":    sanitizeRequestPayload(request),
		"statusCode": statusCode,
		"durationMs": time.Since(startedAt).Milliseconds(),
	}
	if len(responseBody) > 0 {
		payload["response"] = truncateLogValue(string(responseBody))
	}
	if errorMessage != "" {
		payload["error"] = errorMessage
	}
	return payload
}

func sanitizeRequestPayload(request any) any {
	switch actual := request.(type) {
	case nil:
		return nil
	case string:
		return truncateLogValue(actual)
	default:
		raw, err := json.Marshal(actual)
		if err != nil {
			return truncateLogValue(mustJSON(actual))
		}
		return truncateLogValue(string(raw))
	}
}

func truncateLogValue(value string) string {
	if len(value) <= externalAPILogLimit {
		return value
	}
	return value[:externalAPILogLimit] + "...(truncated)"
}

func maskURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	for _, key := range []string{"appkey", "token", "api_key", "apikey"} {
		if query.Has(key) {
			query.Set(key, maskSecret(query.Get(key)))
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}
