package service

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"go-fiber-starter/internal/model/base"
	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"gorm.io/gorm"
)

func TestGenerateJWTUsesConfigExpiration(t *testing.T) {
	setTestJWTConfig(t)

	user := &model.User{
		BaseModel: base.BaseModel{Id: uuid.New()},
		Username:  "test-user",
	}

	start := time.Now()
	tokenString, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT returned error: %v", err)
	}

	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Current.Jwt.Secret), nil
	})
	if err != nil {
		t.Fatalf("Parse token error: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Expected MapClaims, got %T", parsed.Claims)
	}

	expRaw, ok := claims["exp"]
	if !ok {
		t.Fatalf("exp claim missing")
	}

	expFloat, ok := expRaw.(float64)
	if !ok {
		t.Fatalf("exp claim type %T", expRaw)
	}

	expUnix := int64(expFloat)
	expectedMin := start.Unix() + int64(config.Current.Jwt.Expiration)
	end := time.Now()
	expectedMax := end.Unix() + int64(config.Current.Jwt.Expiration)

	if expUnix < expectedMin || expUnix > expectedMax {
		t.Fatalf("exp claim %d not within [%d, %d]", expUnix, expectedMin, expectedMax)
	}
}

func TestCurrentUserUserIDString(t *testing.T) {
	setTestJWTConfig(t)
	user := setupTestUser(t)
	ctx := newCtxWithToken(t, jwt.MapClaims{"user_id": user.Id.String()})

	currentUser, err := CurrentUser(ctx)
	if err != nil {
		t.Fatalf("CurrentUser returned error: %v", err)
	}

	if currentUser.Id != user.Id {
		t.Fatalf("CurrentUser id %s != %s", currentUser.Id, user.Id)
	}
}

func TestParseUserIDClaimUUID(t *testing.T) {
	userID := uuid.New()
	claims := jwt.MapClaims{"user_id": userID}

	parsed, err := parseUserIDClaim(claims)
	if err != nil {
		t.Fatalf("parseUserIDClaim returned error: %v", err)
	}

	if parsed != userID.String() {
		t.Fatalf("parseUserIDClaim id %s != %s", parsed, userID.String())
	}
}

func TestParseUserIDClaimBytes(t *testing.T) {
	userID := uuid.New()
	claims := jwt.MapClaims{"user_id": []byte(userID.String())}

	parsed, err := parseUserIDClaim(claims)
	if err != nil {
		t.Fatalf("parseUserIDClaim returned error: %v", err)
	}

	if parsed != userID.String() {
		t.Fatalf("parseUserIDClaim id %s != %s", parsed, userID.String())
	}
}

func TestCurrentUserUserIDInvalidType(t *testing.T) {
	setTestJWTConfig(t)
	setupTestUser(t)
	ctx := newCtxWithToken(t, jwt.MapClaims{"user_id": 123})

	currentUser, err := CurrentUser(ctx)
	if err == nil {
		t.Fatalf("CurrentUser expected error, got user %v", currentUser)
	}
}

func TestCurrentUserUserIDMissing(t *testing.T) {
	setTestJWTConfig(t)
	setupTestUser(t)
	ctx := newCtxWithToken(t, jwt.MapClaims{})

	currentUser, err := CurrentUser(ctx)
	if err == nil {
		t.Fatalf("CurrentUser expected error, got user %v", currentUser)
	}
}

func newCtxWithToken(t *testing.T, claims jwt.MapClaims) *fiber.Ctx {
	t.Helper()

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	t.Cleanup(func() {
		app.ReleaseCtx(ctx)
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Current.Jwt.Secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Current.Jwt.Secret), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	ctx.Locals("user", parsed)

	return ctx
}

func newCtxWithRawToken(t *testing.T, token *jwt.Token) *fiber.Ctx {
	t.Helper()

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	t.Cleanup(func() {
		app.ReleaseCtx(ctx)
	})

	ctx.Locals("user", token)

	return ctx
}

func setupTestUser(t *testing.T) *model.User {
	t.Helper()

	prevDB := db.DB
	t.Cleanup(func() {
		db.DB = prevDB
	})

	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := gormDB.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	user := &model.User{Username: "test-user"}
	if err := gormDB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	db.DB = gormDB

	return user
}

func setTestJWTConfig(t *testing.T) {
	t.Helper()

	prevSecret := config.Current.Jwt.Secret
	prevExpiration := config.Current.Jwt.Expiration

	config.Current.Jwt.Secret = "test-secret"
	config.Current.Jwt.Expiration = 3600

	t.Cleanup(func() {
		config.Current.Jwt.Secret = prevSecret
		config.Current.Jwt.Expiration = prevExpiration
	})
}

func TestCurrentUserInvalidToken(t *testing.T) {
	setTestJWTConfig(t)
	user := setupTestUser(t)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": user.Id.String()})
	ctx := newCtxWithRawToken(t, token)

	currentUser, err := CurrentUser(ctx)
	if err == nil {
		t.Fatalf("CurrentUser expected error, got user %v", currentUser)
	}
}
