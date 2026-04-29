package service

import (
	"errors"

	"backend/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct{}

func NewAdminService() *AdminService {
	return &AdminService{}
}

func (s *AdminService) ListUsers(keyword, role string) ([]model.UserListItem, error) {
	var users []model.User
	query := DB.Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if err := query.Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}

	list := make([]model.UserListItem, 0, len(users))
	for _, user := range users {
		list = append(list, model.UserListItem{
			UserId:    user.ID,
			UserName:  user.Username,
			Nickname:  user.Nickname,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	return list, nil
}

func (s *AdminService) CreateUser(req model.CreateUserRequest) error {
	if err := s.ensureRoleExists(req.Role); err != nil {
		return err
	}

	var count int64
	if err := DB.Model(&model.User{}).Where("username = ?", req.UserName).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		Username:     req.UserName,
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Role:         req.Role,
	}
	return DB.Create(&user).Error
}

func (s *AdminService) UpdateUser(id uint, req model.UpdateUserRequest) error {
	if err := s.ensureRoleExists(req.Role); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"nickname": req.Nickname,
		"role":     req.Role,
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password_hash"] = string(hash)
	}

	result := DB.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

func (s *AdminService) DeleteUser(id, currentUserID uint) error {
	if id == currentUserID {
		return errors.New("不能删除当前登录用户")
	}
	result := DB.Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

func (s *AdminService) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := DB.Order("id ASC").Find(&roles).Error
	return roles, err
}

func (s *AdminService) ListPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	err := DB.Order("group_name ASC, id ASC").Find(&permissions).Error
	return permissions, err
}

func (s *AdminService) GetRolePermissions(roleCode string) ([]string, error) {
	if err := s.ensureRoleExists(roleCode); err != nil {
		return nil, err
	}

	var rows []model.RolePermission
	if err := DB.Where("role_code = ?", roleCode).Find(&rows).Error; err != nil {
		return nil, err
	}

	permissions := make([]string, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, row.PermissionCode)
	}
	return permissions, nil
}

func (s *AdminService) UpdateRolePermissions(roleCode string, permissions []string) error {
	if roleCode == "R_SUPER" {
		return errors.New("超级管理员默认拥有全部权限，无需配置")
	}
	if err := s.ensureRoleExists(roleCode); err != nil {
		return err
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_code = ?", roleCode).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}

		for _, permission := range permissions {
			row := model.RolePermission{RoleCode: roleCode, PermissionCode: permission}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AdminService) ensureRoleExists(roleCode string) error {
	var role model.Role
	if err := DB.Where("code = ?", roleCode).First(&role).Error; err != nil {
		return errors.New("角色不存在")
	}
	return nil
}
