package model

import (
	"time"
)

// User 用户实体类 (GORM Model)
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"userId"`
	Username     string    `gorm:"size:50;not null;unique" json:"userName"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Nickname     string    `gorm:"size:100" json:"nickname"`
	Role         string    `gorm:"size:20;default:'R_USER'" json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// LoginRequest 登录请求
type LoginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录成功的响应数据
type LoginResponseData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// UserInfoResponse 用户信息响应数据
type UserInfoResponseData struct {
	UserId      string   `json:"userId"`
	UserName    string   `json:"userName"`
	Roles       []string `json:"roles"`
	Buttons     []string `json:"buttons"`
	Permissions []string `json:"permissions"`
}

type UserListItem struct {
	UserId    uint      `json:"userId"`
	UserName  string    `json:"userName"`
	Nickname  string    `json:"nickname"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateUserRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Role     string `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Role     string `json:"role" binding:"required"`
}

// UserProfileResponse 用户详细信息响应
type UserProfileResponse struct {
	UserId    uint      `json:"userId"`
	UserName  string    `json:"userName"`
	Nickname  string    `json:"nickname"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UpdateProfileRequest 更新当前用户信息请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
