package lottery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupAIRequestLogTestDB(t *testing.T) {
	t.Helper()
	setupImportTicketTestDB(t)
	if err := db.DB.AutoMigrate(&model.AIRequestLog{}); err != nil {
		t.Fatal(err)
	}
}

func TestAIRequestLogRetentionAndAccountScope(t *testing.T) {
	setupAIRequestLogTestDB(t)
	user, otherUser := uuid.New(), uuid.New()
	limit := config.Current.AI.RequestLogLimit
	config.Current.AI.RequestLogLimit = 0
	t.Cleanup(func() { config.Current.AI.RequestLogLimit = limit })

	for _, entry := range []*model.AIRequestLog{
		{UserID: user, AIRequestLogSummary: model.AIRequestLogSummary{Purpose: aiRequestPurposeReco}},
		{UserID: otherUser, AIRequestLogSummary: model.AIRequestLogSummary{Purpose: aiRequestPurposeReco}},
		{UserID: user, AIRequestLogSummary: model.AIRequestLogSummary{Purpose: aiRequestPurposeOCR}},
		{UserID: user, AIRequestLogSummary: model.AIRequestLogSummary{Purpose: aiRequestPurposeReco}},
	} {
		entry.Request = "request details"
		if err := saveAIRequestLog(db.DB, entry, 3); err != nil {
			t.Fatal(err)
		}
	}
	logs, err := ListAIRequestLogs(user.String(), 999, "")
	if err != nil {
		t.Fatal(err)
	}
	if logs.Total != 2 || logs.KeepLimit != 1000 || logs.Page != 1 || len(logs.Items) != 2 ||
		logs.Items[0].ID != 4 || logs.Items[1].ID != 3 {
		t.Fatalf("unexpected scoped logs: %+v", logs)
	}
	filtered, err := ListAIRequestLogs(user.String(), 1, aiRequestPurposeOCR)
	if err != nil || filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].ID != 3 {
		t.Fatalf("unexpected filtered logs: %+v (%v)", filtered, err)
	}
	for _, id := range []uint64{1, 2} {
		if _, err := GetAIRequestLog(user.String(), id); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected deleted or inaccessible log %d, got %v", id, err)
		}
	}
	detail, err := GetAIRequestLog(user.String(), 4)
	if err != nil || detail.Request != "request details" {
		t.Fatalf("unexpected detail: %+v (%v)", detail, err)
	}
}

func TestOpenAIRequestsPersistSuccessAndFailureWithRedaction(t *testing.T) {
	setupAIRequestLogTestDB(t)
	const key = "test-api-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer "+key {
			t.Error("upstream request lost its Authorization header")
		}
		if r.URL.Path == "/failure/chat/completions" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"error":"rate limit","token":%q}`, key)
			return
		}
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"recommendation"}}],"usage":{"total_tokens":123}}`)
	}))
	defer server.Close()

	user := uuid.New()
	scope := aiRequestScope{
		userID: user, purpose: aiRequestPurposeReco, lotteryCode: "ssq", issue: "2026136", attempt: 1,
	}
	prompt := strings.Repeat("开奖历史\n", 600) + "末尾也要保留"
	messages := []openAIMessage{
		{Role: "user", Content: prompt},
		{Role: "user", Content: []map[string]any{
			{"type": "image_url", "image_url": map[string]string{"url": "data:image/jpeg;base64,PRIVATEIMG"}},
		}},
	}
	ctx := withAIRequestScope(context.Background(), scope)
	if content, err := callOpenAICompatible(ctx, server.URL, key, "test-model", time.Second, messages); err != nil || content != "recommendation" {
		t.Fatalf("unexpected model result: %q (%v)", content, err)
	}
	scope.attempt = 2
	ctx = withAIRequestScope(context.Background(), scope)
	if _, err := callOpenAICompatible(ctx, server.URL+"/failure", key, "test-model", time.Second, messages); err == nil {
		t.Fatal("expected an HTTP error")
	}

	logs, err := ListAIRequestLogs(user.String(), 1, "")
	if err != nil || len(logs.Items) != 2 {
		t.Fatalf("expected both calls: %+v (%v)", logs, err)
	}
	failed, err := GetAIRequestLog(user.String(), logs.Items[0].ID)
	if err != nil || failed.Success || failed.Attempt != 2 || failed.StatusCode != 429 ||
		!strings.Contains(failed.Error, "HTTP 429") || !strings.Contains(failed.Response, "rate limit") {
		t.Fatalf("unexpected failed log: %+v (%v)", failed.AIRequestLogSummary, err)
	}
	succeeded, err := GetAIRequestLog(user.String(), logs.Items[1].ID)
	if err != nil || !succeeded.Success || succeeded.Attempt != 1 || succeeded.Issue != "2026136" ||
		!strings.Contains(succeeded.Request, "末尾也要保留") ||
		!strings.Contains(succeeded.Response, "total_tokens") || !strings.Contains(succeeded.Request, "base64 omitted") {
		t.Fatalf("unexpected successful log: %+v (%v)", succeeded.AIRequestLogSummary, err)
	}
	for _, entry := range []model.AIRequestLog{failed, succeeded} {
		for _, forbidden := range []string{key, "PRIVATEIMG", "127.0.0.1"} {
			if strings.Contains(entry.Request+entry.Response+entry.Error+entry.Endpoint, forbidden) {
				t.Fatalf("credential, private address or image was written to log %d", entry.ID)
			}
		}
	}
}

func TestAIRequestLogRedactionAndOCRProjection(t *testing.T) {
	endpoint := "https://someone:secret-pass@localhost:8080/ocr?token=secret-query&version=1"
	redact := newAIRequestLogRedactor(endpoint, "test-key")
	text := redact(endpoint+" secret-pass secret-query 11 "+`{"cookie":"session-123","url":"data:image/png;base64,SECRETIMG"}`, 1024)
	if text != `https://private-host/ocr [已脱敏] [已脱敏] 11 {"cookie":"[已脱敏]","url":"[图片已省略]"}` {
		t.Fatalf("unexpected redaction: %q", text)
	}
	payload := paddleOCRResponseLog([]byte(`{"result":{"inputImage":"PRIVATE","layoutParsingResults":[{"markdown":{"text":"双色球", "images":{"a":"PRIVATE"}}}]}}`))
	if !strings.Contains(payload, "双色球") || strings.Contains(payload, "PRIVATE") {
		t.Fatalf("unexpected PaddleOCR response projection: %s", payload)
	}
	if cut := redact("123双色球", 5); !utf8.ValidString(cut) || cut != "123\n[内容已截断]" {
		t.Fatalf("unexpected truncation: %q", cut)
	}
}
