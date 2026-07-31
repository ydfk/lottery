package auth

import (
	"encoding/json"
	"strings"

	"go-fiber-starter/internal/api/response"
	userModel "go-fiber-starter/internal/model/user"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/gofiber/fiber/v2"
)

type passkeyLoginRequest struct {
	Username string `json:"username"`
}

type passkeyFinishRequest struct {
	CeremonyID string          `json:"ceremonyId"`
	Credential json.RawMessage `json:"credential"`
}

func BeginPasskeyRegistration(c *fiber.Ctx) error {
	user, err := currentUserWithPasskeys(c)
	if err != nil {
		return response.Error(c, "用户未找到", fiber.StatusUnauthorized)
	}
	options, ceremonyID, err := service.BeginPasskeyRegistration(*user)
	if err != nil {
		return passkeyError(c, err)
	}
	return response.Success(c, fiber.Map{
		"ceremonyId": ceremonyID,
		"publicKey":  options.Response,
	})
}

func FinishPasskeyRegistration(c *fiber.Ctx) error {
	user, err := currentUserWithPasskeys(c)
	if err != nil {
		return response.Error(c, "用户未找到", fiber.StatusUnauthorized)
	}
	request, err := parsePasskeyFinishRequest(c)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest)
	}
	if err := service.FinishPasskeyRegistration(*user, request.CeremonyID, request.Credential); err != nil {
		return passkeyError(c, err)
	}
	return response.Success(c, fiber.Map{"registered": true})
}

func BeginPasskeyLogin(c *fiber.Ctx) error {
	request := passkeyLoginRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}
	user, err := userWithPasskeysByUsername(request.Username)
	if err != nil {
		return response.Error(c, "用户名或 Passkey 不正确", fiber.StatusUnauthorized)
	}
	options, ceremonyID, err := service.BeginPasskeyLogin(*user)
	if err != nil {
		return response.Error(c, "用户名或 Passkey 不正确", fiber.StatusUnauthorized)
	}
	return response.Success(c, fiber.Map{
		"ceremonyId": ceremonyID,
		"publicKey":  options.Response,
	})
}

func FinishPasskeyLogin(c *fiber.Ctx) error {
	request := passkeyFinishRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}
	if strings.TrimSpace(request.CeremonyID) == "" || len(request.Credential) == 0 {
		return response.Error(c, "Passkey 响应不完整", fiber.StatusBadRequest)
	}
	ceremony := userModel.PasskeyCeremony{}
	if err := db.DB.First(&ceremony, "id = ? AND kind = ?", request.CeremonyID, "login").Error; err != nil {
		return response.Error(c, "Passkey 挑战已失效，请重试", fiber.StatusUnauthorized)
	}
	user := userModel.User{}
	if err := db.DB.Preload("Passkeys").First(&user, "id = ?", ceremony.UserID).Error; err != nil {
		return response.Error(c, "用户名或 Passkey 不正确", fiber.StatusUnauthorized)
	}
	if err := service.FinishPasskeyLogin(user, request.CeremonyID, request.Credential); err != nil {
		return response.Error(c, "Passkey 验证失败，请重试", fiber.StatusUnauthorized)
	}
	token, err := service.GenerateJWT(&user)
	if err != nil {
		return response.Error(c, "token生成失败")
	}
	return response.Success(c, fiber.Map{"token": token})
}

func AppleAppSiteAssociation(c *fiber.Ctx) error {
	appID := strings.TrimSpace(config.Current.Passkey.AppID)
	if appID == "" {
		return fiber.ErrNotFound
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(fiber.Map{
		"webcredentials": fiber.Map{"apps": []string{appID}},
	})
}

func currentUserWithPasskeys(c *fiber.Ctx) (*userModel.User, error) {
	current, err := service.CurrentUser(c)
	if err != nil {
		return nil, err
	}
	user := userModel.User{}
	if err := db.DB.Preload("Passkeys").First(&user, "id = ?", current.Id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func userWithPasskeysByUsername(username string) (*userModel.User, error) {
	user := userModel.User{}
	if err := db.DB.Preload("Passkeys").First(&user, "username = ?", strings.TrimSpace(username)).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func parsePasskeyFinishRequest(c *fiber.Ctx) (passkeyFinishRequest, error) {
	request := passkeyFinishRequest{}
	if err := c.BodyParser(&request); err != nil {
		return request, fiber.NewError(fiber.StatusBadRequest, "参数不正确")
	}
	if strings.TrimSpace(request.CeremonyID) == "" || len(request.Credential) == 0 {
		return request, fiber.NewError(fiber.StatusBadRequest, "Passkey 响应不完整")
	}
	return request, nil
}

func passkeyError(c *fiber.Ctx, err error) error {
	return response.Error(c, err.Error(), fiber.StatusBadRequest)
}
