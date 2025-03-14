package handlers

import (
	"time"

	"lottery-backend/internal/config"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register 用户注册
func Register(c *fiber.Ctx) error {
	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的请求数据",
		})
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "用户名已存在",
		})
	}

	// 创建新用户
	user := models.User{
		Username: req.Username,
		Password: req.Password,
	}

	// 加密密码
	if err := user.HashPassword(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "创建用户失败",
		})
	}

	// 保存用户
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "创建用户失败",
		})
	}

	// 生成JWT
	token, err := generateToken(user.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "生成token失败",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

// Login 用户登录
func Login(c *fiber.Ctx) error {
	var req AuthRequest

	// 尝试从请求体解析数据
	if err := c.BodyParser(&req); err != nil {
		// 如果请求体解析失败，尝试从查询参数获取
		username := c.Query("username")
		password := c.Query("password")

		if username == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "无效的请求数据",
			})
		}

		req.Username = username
		req.Password = password
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "用户名或密码错误",
		})
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "用户名或密码错误",
		})
	}

	// 生成JWT
	token, err := generateToken(user.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "生成token失败",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

// generateToken 生成JWT token
func generateToken(userID uint) (string, error) {
	// 创建token
	token := jwt.New(jwt.SigningMethodHS256)

	// 设置声明
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Duration(config.Current.JWT.Expiration) * time.Second).Unix()

	// 签名token
	return token.SignedString([]byte(config.Current.JWT.Secret))
}
