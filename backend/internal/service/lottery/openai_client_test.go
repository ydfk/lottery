package lottery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCallOpenAICompatibleReportsNonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html>not found</html>"))
	}))
	defer server.Close()

	_, err := callOpenAICompatible(
		context.Background(),
		server.URL,
		"test-key",
		"test-model",
		time.Second,
		[]openAIMessage{{Role: "user", Content: "hello"}},
	)
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	for _, expected := range []string{"模型响应不是 JSON", "Content-Type=text/html", "<html>not found</html>", "ai.baseURL"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("error %q does not contain %q", message, expected)
		}
	}
}

func TestCallOpenAICompatibleReportsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<html>404</html>"))
	}))
	defer server.Close()

	_, err := callOpenAICompatible(
		context.Background(),
		server.URL,
		"test-key",
		"test-model",
		time.Second,
		[]openAIMessage{{Role: "user", Content: "hello"}},
	)
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	for _, expected := range []string{"模型请求失败", "HTTP 404", "Content-Type=text/html", "<html>404</html>"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("error %q does not contain %q", message, expected)
		}
	}
}
