package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegister_HashError(t *testing.T) {
	original := generateFromPassword
	t.Cleanup(func() {
		generateFromPassword = original
	})
	generateFromPassword = func(_ []byte, _ int) ([]byte, error) {
		return nil, errors.New("boom")
	}

	app := fiber.New()
	app.Post("/api/auth/register", Register)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("{\"username\":\"alice\",\"password\":\"secret\"}"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	var payload struct {
		Flag bool   `json:"flag"`
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Flag {
		t.Fatalf("expected error response")
	}
	if payload.Code != http.StatusInternalServerError {
		t.Fatalf("expected code %d, got %d", http.StatusInternalServerError, payload.Code)
	}
	if payload.Msg != "密码加密失败" {
		t.Fatalf("expected message %q, got %q", "密码加密失败", payload.Msg)
	}
}
