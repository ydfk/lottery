package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"gorm.io/gorm"

	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
)

type responseEnvelope struct {
	Flag bool            `json:"flag"`
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type userResponse struct {
	Username string `json:"username"`
}

func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()

	prevConfig := config.Current
	config.Current.Jwt.Secret = "test-secret"
	config.Current.Jwt.Expiration = 3600
	config.Current.App.Env = "test"
	config.Current.App.Port = "0"
	config.Current.Database.Path = ""
	config.IsProduction = false

	prevDB := db.DB
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gormDB.AutoMigrate(&model.User{}, &model.PasskeyCredential{}, &model.PasskeyCeremony{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	closeSQLDB(t, gormDB)
	db.DB = gormDB

	t.Cleanup(func() {
		config.Current = prevConfig
		db.DB = prevDB
	})

	app := fiber.New()
	RegisterUnProtectedRoutes(app)

	api := app.Group("/api")
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(config.Current.Jwt.Secret),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": "认证失败，请先登录",
			})
		},
	}))
	RegisterRoutes(api)

	return app
}

func closeSQLDB(t *testing.T, gormDB *gorm.DB) {
	t.Helper()
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
}

func doJSONRequest(t *testing.T, app *fiber.App, method, path string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	var reader *bytes.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	return resp
}

func decodeEnvelope(t *testing.T, resp *http.Response) responseEnvelope {
	t.Helper()
	defer resp.Body.Close()
	var envelope responseEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return envelope
}

func TestAuthFlowRegisterLoginProfile(t *testing.T) {
	app := setupTestApp(t)

	registerResp := doJSONRequest(t, app, http.MethodPost, "/api/auth/register", fiber.Map{
		"username": "alice",
		"password": "pass123",
	}, nil)
	if registerResp.StatusCode != http.StatusOK {
		t.Fatalf("register status: %d", registerResp.StatusCode)
	}
	registerEnvelope := decodeEnvelope(t, registerResp)
	if !registerEnvelope.Flag {
		t.Fatalf("register failed: %s", registerEnvelope.Msg)
	}
	var registeredUser userResponse
	if err := json.Unmarshal(registerEnvelope.Data, &registeredUser); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	if registeredUser.Username != "alice" {
		t.Fatalf("unexpected username: %s", registeredUser.Username)
	}

	loginResp := doJSONRequest(t, app, http.MethodPost, "/api/auth/login", fiber.Map{
		"username": "alice",
		"password": "pass123",
	}, nil)
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status: %d", loginResp.StatusCode)
	}
	loginEnvelope := decodeEnvelope(t, loginResp)
	if !loginEnvelope.Flag {
		t.Fatalf("login failed: %s", loginEnvelope.Msg)
	}
	var token tokenResponse
	if err := json.Unmarshal(loginEnvelope.Data, &token); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if token.Token == "" {
		t.Fatalf("empty token")
	}

	profileResp := doJSONRequest(t, app, http.MethodGet, "/api/auth/profile", nil, map[string]string{
		"Authorization": "Bearer " + token.Token,
	})
	if profileResp.StatusCode != http.StatusOK {
		t.Fatalf("profile status: %d", profileResp.StatusCode)
	}
	profileEnvelope := decodeEnvelope(t, profileResp)
	if !profileEnvelope.Flag {
		t.Fatalf("profile failed: %s", profileEnvelope.Msg)
	}
	var profileUser userResponse
	if err := json.Unmarshal(profileEnvelope.Data, &profileUser); err != nil {
		t.Fatalf("decode profile user: %v", err)
	}
	if profileUser.Username != "alice" {
		t.Fatalf("unexpected profile username: %s", profileUser.Username)
	}
}

func TestProfileRequiresAuth(t *testing.T) {
	app := setupTestApp(t)

	resp := doJSONRequest(t, app, http.MethodGet, "/api/auth/profile", nil, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", resp.StatusCode)
	}
}

func TestPasskeyListAndDelete(t *testing.T) {
	app := setupTestApp(t)

	doJSONRequest(t, app, http.MethodPost, "/api/auth/register", fiber.Map{
		"username": "passkey-owner",
		"password": "pass123",
	}, nil).Body.Close()
	loginResp := doJSONRequest(t, app, http.MethodPost, "/api/auth/login", fiber.Map{
		"username": "passkey-owner",
		"password": "pass123",
	}, nil)
	loginEnvelope := decodeEnvelope(t, loginResp)
	var token tokenResponse
	if err := json.Unmarshal(loginEnvelope.Data, &token); err != nil {
		t.Fatalf("decode token: %v", err)
	}

	owner := model.User{}
	if err := db.DB.First(&owner, "username = ?", "passkey-owner").Error; err != nil {
		t.Fatalf("find owner: %v", err)
	}
	passkey := model.PasskeyCredential{
		UserID:         owner.Id,
		CredentialID:   []byte("credential-id"),
		CredentialJSON: []byte(`{"id":"credential-id"}`),
	}
	if err := db.DB.Create(&passkey).Error; err != nil {
		t.Fatalf("create passkey: %v", err)
	}
	headers := map[string]string{"Authorization": "Bearer " + token.Token}

	listResp := doJSONRequest(t, app, http.MethodGet, "/api/auth/passkeys", nil, headers)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list passkeys status: %d", listResp.StatusCode)
	}
	listEnvelope := decodeEnvelope(t, listResp)
	var items []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(listEnvelope.Data, &items); err != nil {
		t.Fatalf("decode passkeys: %v", err)
	}
	if len(items) != 1 || items[0].ID != passkey.Id.String() {
		t.Fatalf("unexpected passkeys: %#v", items)
	}

	deleteResp := doJSONRequest(t, app, http.MethodDelete, "/api/auth/passkeys/"+passkey.Id.String(), nil, headers)
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete passkey status: %d", deleteResp.StatusCode)
	}
	decodeEnvelope(t, deleteResp)
	var count int64
	if err := db.DB.Model(&model.PasskeyCredential{}).Where("id = ?", passkey.Id).Count(&count).Error; err != nil {
		t.Fatalf("count passkeys: %v", err)
	}
	if count != 0 {
		t.Fatalf("passkey was not deleted")
	}
}

func TestPasskeyLoginOptionsDoNotRequireUsername(t *testing.T) {
	app := setupTestApp(t)
	config.Current.Passkey.RPID = "localhost"
	config.Current.Passkey.RPName = "彩迹测试"
	config.Current.Passkey.Origins = []string{"https://localhost"}

	resp := doJSONRequest(t, app, http.MethodPost, "/api/auth/passkeys/login/options", nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("passkey options status: %d", resp.StatusCode)
	}
	envelope := decodeEnvelope(t, resp)
	if !envelope.Flag {
		t.Fatalf("passkey options failed: %s", envelope.Msg)
	}
	var options struct {
		CeremonyID string `json:"ceremonyId"`
		PublicKey  struct {
			Challenge string `json:"challenge"`
		} `json:"publicKey"`
	}
	if err := json.Unmarshal(envelope.Data, &options); err != nil {
		t.Fatalf("decode passkey options: %v", err)
	}
	if options.CeremonyID == "" || options.PublicKey.Challenge == "" {
		t.Fatalf("incomplete passkey options: %#v", options)
	}
}

func TestAppleAppSiteAssociation(t *testing.T) {
	previousConfig := config.Current
	config.Current.Passkey.AppID = "B6AA23TY3H.com.ydfk.lottery"
	t.Cleanup(func() { config.Current = previousConfig })

	app := fiber.New()
	RegisterAssociatedDomainRoutes(app)
	resp := doJSONRequest(t, app, http.MethodGet, "/.well-known/apple-app-site-association", nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("association status: %d", resp.StatusCode)
	}
	var document struct {
		WebCredentials struct {
			Apps []string `json:"apps"`
		} `json:"webcredentials"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&document); err != nil {
		t.Fatalf("decode association document: %v", err)
	}
	if len(document.WebCredentials.Apps) != 1 || document.WebCredentials.Apps[0] != config.Current.Passkey.AppID {
		t.Fatalf("unexpected app IDs: %#v", document.WebCredentials.Apps)
	}
}
