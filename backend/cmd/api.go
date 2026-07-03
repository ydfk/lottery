package main

import (
	"os"
	urlpath "path"
	"path/filepath"
	"strings"

	"go-fiber-starter/internal/api/auth"
	lotteryApi "go-fiber-starter/internal/api/lottery"
	"go-fiber-starter/internal/middleware"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
)

func api() {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Static("/uploads", config.Current.Storage.UploadDir)

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format: "${ip} ${status} ${latency} ${method} ${path}\n",
		Output: logger.GetFiberLogWriter(),
	}))

	auth.RegisterUnProtectedRoutes(app)

	api := app.Group("/api")
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(config.Current.Jwt.Secret),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error("JWT验证失败: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": "认证失败，请先登录",
			})
		},
	}))

	auth.RegisterRoutes(api)
	lotteryApi.RegisterRoutes(api)
	registerFrontend(app)

	if err := app.Listen(":" + config.Current.App.Port); err != nil {
		logger.Fatal("启动服务器失败: %v", err)
	}
	logger.Info("服务器启动成功: http://127.0.0.1:%v ", config.Current.App.Port)
}

func registerFrontend(app *fiber.App) {
	webRoot := filepath.Join(".", "web")
	indexPath := filepath.Join(webRoot, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	app.Use(applyFrontendCacheHeaders)
	app.Static("/", webRoot, fiber.Static{
		Compress:  true,
		ByteRange: true,
	})

	app.Use(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") || strings.HasPrefix(c.Path(), "/swagger") || strings.HasPrefix(c.Path(), "/uploads") {
			return fiber.ErrNotFound
		}
		setNoStoreHeaders(c)
		return c.SendFile(indexPath)
	})
}

func applyFrontendCacheHeaders(c *fiber.Ctx) error {
	requestPath := c.Path()
	if isNoStoreFrontendPath(requestPath) {
		setNoStoreHeaders(c)
		return c.Next()
	}
	if strings.HasPrefix(requestPath, "/assets/") {
		c.Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	return c.Next()
}

func isNoStoreFrontendPath(requestPath string) bool {
	if requestPath == "/" || requestPath == "/index.html" || requestPath == "/app-version.json" {
		return true
	}
	if strings.HasPrefix(requestPath, "/api") ||
		strings.HasPrefix(requestPath, "/swagger") ||
		strings.HasPrefix(requestPath, "/uploads") {
		return false
	}
	return urlpath.Ext(requestPath) == ""
}

func setNoStoreHeaders(c *fiber.Ctx) {
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
}
