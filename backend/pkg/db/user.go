/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-10 13:58:58
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-10 16:43:35
 */
package db

import model "go-fiber-starter/internal/model/user"

func GetUserById(id string) (model.User, error) {
	var user model.User
	result := DB.First(&user, "id = ?", id)
	if result.Error != nil {
		return user, result.Error
	}

	return user, nil
}
