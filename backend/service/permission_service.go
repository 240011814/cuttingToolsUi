package service

import "backend/model"

// CheckUserPermission 检查用户是否拥有所需权限中的任意一个
func CheckUserPermission(userID uint, permissions []string) (bool, error) {
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		return false, err
	}

	// 查询用户角色拥有的权限
	var count int64
	err := DB.Table("role_permissions").
		Where("role_code = ? AND permission_code IN ?", user.Role, permissions).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
