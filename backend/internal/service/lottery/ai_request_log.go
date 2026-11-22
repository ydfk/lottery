package lottery

import (
	"context"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

const (
	aiRequestLogDefaultLimit = 1000
	aiRequestLogPayloadLimit = 64 * 1024
	aiRequestPurposeReco     = "recommendation"
	aiRequestPurposeOCR      = "ocr"
)

type aiRequestContextKey struct{}

type aiRequestScope struct {
	userID      uuid.UUID
	purpose     string
	lotteryCode string
	issue       string
	attempt     int
}

func withAIRequestScope(ctx context.Context, scope aiRequestScope) context.Context {
	return context.WithValue(ctx, aiRequestContextKey{}, scope)
}

type AIRequestLogPage struct {
	Items     []model.AIRequestLogSummary `json:"items"`
	Page      int                         `json:"page"`
	PageSize  int                         `json:"pageSize"`
	Total     int64                       `json:"total"`
	KeepLimit int                         `json:"keepLimit"`
}

func ListAIRequestLogs(userID string, page int, purpose string) (AIRequestLogPage, error) {
	result := AIRequestLogPage{
		Items:     make([]model.AIRequestLogSummary, 0),
		Page:      max(1, page),
		PageSize:  20,
		KeepLimit: aiRequestLogKeepLimit(),
	}
	query := db.DB.Model(&model.AIRequestLog{}).Where("user_id = ?", userID)
	if purpose != "" {
		query = query.Where("purpose = ?", purpose)
	}
	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}
	lastPage := max(1, (int(result.Total)+result.PageSize-1)/result.PageSize)
	result.Page = min(result.Page, lastPage)
	err := query.Order("id DESC").Offset((result.Page - 1) * result.PageSize).
		Limit(result.PageSize).Find(&result.Items).Error
	return result, err
}

func GetAIRequestLog(userID string, id uint64) (model.AIRequestLog, error) {
	var entry model.AIRequestLog
	err := db.DB.Where("user_id = ? AND id = ?", userID, id).Take(&entry).Error
	return entry, err
}

func aiRequestLogKeepLimit() int {
	if limit := config.Current.AI.RequestLogLimit; limit > 0 {
		return limit
	}
	return aiRequestLogDefaultLimit
}

var aiRequestLogWriteLock sync.Mutex

func saveAIRequestLog(database *gorm.DB, entry *model.AIRequestLog, limit int) error {
	aiRequestLogWriteLock.Lock()
	defer aiRequestLogWriteLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return database.WithContext(ctx).Session(&gorm.Session{Logger: gormlog.Default.LogMode(gormlog.Silent)}).
		Transaction(func(tx *gorm.DB) error {
			// PostgreSQL 多实例也需要把写入和淘汰串行化。
			if tx.Dialector.Name() == "postgres" {
				if err := tx.Exec("LOCK TABLE ai_request_logs IN SHARE ROW EXCLUSIVE MODE").Error; err != nil {
					return err
				}
			}
			if err := tx.Create(entry).Error; err != nil {
				return err
			}
			cutoff := tx.Model(&model.AIRequestLog{}).Select("id").Order("id DESC").Offset(limit).Limit(1)
			return tx.Where("id <= (?)", cutoff).Delete(&model.AIRequestLog{}).Error
		})
}

type aiRequestTrace struct {
	startedAt   time.Time
	provider    string
	model       string
	method      string
	endpoint    string
	credential  string
	contentType string
	request     any
	response    string
	statusCode  int
}

func (trace aiRequestTrace) finish(ctx context.Context, requestErr error) {
	duration := time.Since(trace.startedAt).Milliseconds()
	redact := newAIRequestLogRedactor(trace.endpoint, trace.credential)
	endpoint := redact(aiRequestLogEndpoint(trace.endpoint), 2048)
	request := redact(mustJSON(trace.request), aiRequestLogPayloadLimit)
	response := redact(trace.response, aiRequestLogPayloadLimit)
	var errorMessage string
	if requestErr != nil {
		errorMessage = redact(requestErr.Error(), externalAPILogLimit)
		logger.Error("第三方接口调用失败: %s", mustJSON(buildThirdPartyLogPayload(
			trace.provider, trace.method, endpoint, request, []byte(response), trace.statusCode, trace.startedAt, errorMessage,
		)))
	} else {
		logThirdPartySuccess(trace.provider, trace.method, endpoint, request, []byte(response), trace.statusCode, trace.startedAt)
	}

	scope, ok := ctx.Value(aiRequestContextKey{}).(aiRequestScope)
	if !ok || scope.userID == uuid.Nil || db.DB == nil {
		return
	}
	logEntry := model.AIRequestLog{
		AIRequestLogSummary: model.AIRequestLogSummary{
			StartedAt: trace.startedAt.UTC(), Purpose: scope.purpose, Provider: trace.provider,
			Model: redact(trace.model, 512), LotteryCode: scope.lotteryCode, Issue: scope.issue,
			Attempt: max(1, scope.attempt), StatusCode: trace.statusCode, DurationMs: duration,
			Success: requestErr == nil,
		},
		UserID:      scope.userID,
		Method:      trace.method,
		Endpoint:    endpoint,
		ContentType: redact(trace.contentType, 256),
		Request:     request,
		Response:    response,
		Error:       errorMessage,
	}
	if err := saveAIRequestLog(db.DB, &logEntry, aiRequestLogKeepLimit()); err != nil {
		logger.Warn("保存 AI 请求日志失败 (%s): %T", trace.provider, err)
	}
}

var (
	aiLogURL        = regexp.MustCompile(`https?://[^\s"'<>\\]+`)
	aiLogCredential = regexp.MustCompile(`(?i)(bearer\s+|token\s+)[a-z0-9._~+/=-]+`)
	aiLogJSONSecret = regexp.MustCompile(`(?i)("(?:authorization|api[_-]?key|access[_-]?token|token|password|secret|cookie|set-cookie)"\s*:\s*")([^"\\]|\\.)*"`)
	aiLogImage      = regexp.MustCompile(`(?i)data:[^;,\s"]+;base64,[a-z0-9+/=_-]+`)
)

func newAIRequestLogRedactor(endpoint string, credential string) func(string, int) string {
	secrets := []string{credential, strings.TrimSpace(credential)}
	if parsed, err := url.Parse(endpoint); err == nil {
		if parsed.User != nil {
			password, _ := parsed.User.Password()
			secrets = append(secrets, password)
		}
		for key, values := range parsed.Query() {
			switch strings.ToLower(key) {
			case "appkey", "token", "api_key", "apikey", "key", "access_token", "password", "secret", "signature":
				secrets = append(secrets, values...)
			}
		}
	}
	var replacements []string
	for _, secret := range secrets {
		if secret != "" {
			replacements = append(replacements, secret, "[已脱敏]")
		}
	}
	mask := strings.NewReplacer(replacements...)
	return func(value string, limit int) string {
		value = aiLogURL.ReplaceAllStringFunc(value, aiRequestLogEndpoint)
		value = aiLogCredential.ReplaceAllString(value, "${1}[已脱敏]")
		value = aiLogJSONSecret.ReplaceAllString(value, "${1}[已脱敏]\"")
		value = aiLogImage.ReplaceAllString(value, "[图片已省略]")
		value = strings.ToValidUTF8(mask.Replace(value), "�")
		if len(value) <= limit {
			return value
		}
		for !utf8.RuneStart(value[limit]) {
			limit--
		}
		return value[:limit] + "\n[内容已截断]"
	}
}

func aiRequestLogEndpoint(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return ""
	}
	parsed.User = nil
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	ip, ipErr := netip.ParseAddr(host)
	if !strings.Contains(host, ".") && ipErr != nil ||
		strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") ||
		ipErr == nil && (ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()) {
		parsed.Host = "private-host"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.ForceQuery = false
	return parsed.String()
}
