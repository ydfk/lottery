/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 17:25:30
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-10 16:52:53
 */
package user

import (
	"encoding/json"

	"go-fiber-starter/internal/model/base"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type User struct {
	base.BaseModel
	Username string              `gorm:"uniqueIndex;size:64" json:"username" example:"admin"`
	Password string              `json:"-" example:"123456"`
	Passkeys []PasskeyCredential `json:"-"`
}

type PasskeyCredential struct {
	base.BaseModel
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	CredentialID   []byte    `gorm:"not null;uniqueIndex" json:"-"`
	CredentialJSON []byte    `gorm:"type:blob;not null" json:"-"`
}

func (user User) WebAuthnID() []byte {
	return []byte(user.Id.String())
}

func (user User) WebAuthnName() string {
	return user.Username
}

func (user User) WebAuthnDisplayName() string {
	return user.Username
}

func (user User) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, 0, len(user.Passkeys))
	for _, passkey := range user.Passkeys {
		credential := webauthn.Credential{}
		if json.Unmarshal(passkey.CredentialJSON, &credential) == nil {
			credentials = append(credentials, credential)
		}
	}
	return credentials
}
