package auth

import (
	"go-fiber-starter/internal/api/response"
	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var generateFromPassword = bcrypt.GenerateFromPassword

type AuthRequest struct {
	Username string `json:"username" example:"alice"`
	Password string `json:"password" example:"pass123"`
}

// @Summary 用户注册
// @Description 创建账号并返回用户信息
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AuthRequest true "注册信息"
// @Success 200 {object} UserResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func Register(c *fiber.Ctx) error {
	var req AuthRequest

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, "参数不正确")
	}

	hash, err := generateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, "密码加密失败")
	}
	user := model.User{Username: req.Username, Password: string(hash)}
	if err := db.DB.Create(&user).Error; err != nil {
		return response.Error(c, "用户名已存在")
	}

	return response.Success(c, user)
}

// @Summary 用户登录
// @Description 用户登录接口
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AuthRequest true "登录信息"
// @Success 200 {object} LoginResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func Login(c *fiber.Ctx) error {
	var req AuthRequest

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, "参数不正确")
	}

	var user model.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return response.Error(c, "用户名不存在")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return response.Error(c, "密码不正确")
	}

	token, err := service.GenerateJWT(&user)
	if err != nil {
		return response.Error(c, "token生成失败")
	}
	return response.Success(c, fiber.Map{"token": token})
}

// @Summary 获取当前用户
// @Description 返回当前登录用户信息
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/profile [get]
func Profile(c *fiber.Ctx) error {
	user, err := service.CurrentUser(c)
	if err != nil {
		return response.Error(c, "用户未找到")
	}
	return response.Success(c, user)
}
