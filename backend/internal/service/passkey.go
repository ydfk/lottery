package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	userModel "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

const (
	passkeyRegistrationCeremony = "registration"
	passkeyLoginCeremony        = "login"
	passkeyCeremonyLifetime     = 5 * time.Minute
)

func BeginPasskeyRegistration(user userModel.User) (*protocol.CredentialCreation, string, error) {
	server, err := passkeyServer()
	if err != nil {
		return nil, "", err
	}
	selection := protocol.AuthenticatorSelection{
		ResidentKey:      protocol.ResidentKeyRequirementPreferred,
		UserVerification: protocol.VerificationRequired,
	}
	creation, session, err := server.BeginRegistration(user, webauthn.WithAuthenticatorSelection(selection))
	if err != nil {
		return nil, "", err
	}
	ceremonyID, err := savePasskeyCeremony(user.Id, passkeyRegistrationCeremony, *session)
	return creation, ceremonyID, err
}

func FinishPasskeyRegistration(user userModel.User, ceremonyID string, credentialResponse []byte) error {
	server, err := passkeyServer()
	if err != nil {
		return err
	}
	session, err := takePasskeyCeremony(user.Id, ceremonyID, passkeyRegistrationCeremony)
	if err != nil {
		return err
	}
	credential, err := server.FinishRegistration(user, session, passkeyHTTPRequest(credentialResponse))
	if err != nil {
		return err
	}
	encodedCredential, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("保存 Passkey 凭据失败: %w", err)
	}
	passkey := userModel.PasskeyCredential{
		UserID:         user.Id,
		CredentialID:   credential.ID,
		CredentialJSON: encodedCredential,
	}
	if err := db.DB.Create(&passkey).Error; err != nil {
		return fmt.Errorf("保存 Passkey 凭据失败: %w", err)
	}
	return nil
}

func BeginPasskeyLogin(user userModel.User) (*protocol.CredentialAssertion, string, error) {
	if len(user.Passkeys) == 0 {
		return nil, "", fmt.Errorf("该账户尚未绑定 Passkey")
	}
	server, err := passkeyServer()
	if err != nil {
		return nil, "", err
	}
	assertion, session, err := server.BeginLogin(user, webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, "", err
	}
	ceremonyID, err := savePasskeyCeremony(user.Id, passkeyLoginCeremony, *session)
	return assertion, ceremonyID, err
}

func FinishPasskeyLogin(user userModel.User, ceremonyID string, credentialResponse []byte) error {
	server, err := passkeyServer()
	if err != nil {
		return err
	}
	session, err := takePasskeyCeremony(user.Id, ceremonyID, passkeyLoginCeremony)
	if err != nil {
		return err
	}
	credential, err := server.FinishLogin(user, session, passkeyHTTPRequest(credentialResponse))
	if err != nil {
		return err
	}
	encodedCredential, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("更新 Passkey 凭据失败: %w", err)
	}
	if err := db.DB.Model(&userModel.PasskeyCredential{}).
		Where("user_id = ? AND credential_id = ?", user.Id, credential.ID).
		Update("credential_json", encodedCredential).Error; err != nil {
		return fmt.Errorf("更新 Passkey 凭据失败: %w", err)
	}
	return nil
}

func passkeyServer() (*webauthn.WebAuthn, error) {
	rpID := strings.TrimSpace(config.Current.Passkey.RPID)
	origins := config.Current.Passkey.Origins
	if rpID == "" || len(origins) == 0 {
		return nil, fmt.Errorf("Passkey 服务尚未配置")
	}
	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: resolvePasskeyValue(config.Current.Passkey.RPName, "彩迹"),
		RPOrigins:     origins,
	})
}

func savePasskeyCeremony(userID uuid.UUID, kind string, session webauthn.SessionData) (string, error) {
	encodedSession, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("保存 Passkey 挑战失败: %w", err)
	}
	if err := db.DB.Where("expires_at <= ?", time.Now()).Delete(&userModel.PasskeyCeremony{}).Error; err != nil {
		return "", fmt.Errorf("清理过期 Passkey 挑战失败: %w", err)
	}
	ceremony := userModel.PasskeyCeremony{
		UserID:      userID,
		Kind:        kind,
		SessionData: encodedSession,
		ExpiresAt:   time.Now().Add(passkeyCeremonyLifetime),
	}
	if err := db.DB.Create(&ceremony).Error; err != nil {
		return "", fmt.Errorf("保存 Passkey 挑战失败: %w", err)
	}
	return ceremony.Id.String(), nil
}

func takePasskeyCeremony(userID uuid.UUID, ceremonyID string, kind string) (webauthn.SessionData, error) {
	parsedID, err := uuid.Parse(ceremonyID)
	if err != nil {
		return webauthn.SessionData{}, fmt.Errorf("Passkey 挑战无效")
	}
	ceremony := userModel.PasskeyCeremony{}
	if err := db.DB.Where("id = ? AND user_id = ? AND kind = ? AND expires_at > ?", parsedID, userID, kind, time.Now()).First(&ceremony).Error; err != nil {
		return webauthn.SessionData{}, fmt.Errorf("Passkey 挑战已失效，请重试")
	}
	if err := db.DB.Delete(&ceremony).Error; err != nil {
		return webauthn.SessionData{}, fmt.Errorf("删除 Passkey 挑战失败: %w", err)
	}
	session := webauthn.SessionData{}
	if err := json.Unmarshal(ceremony.SessionData, &session); err != nil {
		return webauthn.SessionData{}, fmt.Errorf("Passkey 挑战无效")
	}
	return session, nil
}

func passkeyHTTPRequest(body []byte) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, "https://"+config.Current.Passkey.RPID, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func resolvePasskeyValue(value string, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}
