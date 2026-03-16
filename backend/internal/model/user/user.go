/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 17:25:30
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-10 16:52:53
 */
package user

import (
	"go-fiber-starter/internal/model/base"
)

type User struct {
	base.BaseModel
	Username string `gorm:"uniqueIndex;size:64" json:"username" example:"admin"`
	Password string `json:"-" example:"123456"`
}
