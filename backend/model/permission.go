package model

import "time"

type Role struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:50;not null;unique" json:"code"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"size:100;not null;unique" json:"code"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	GroupName string    `gorm:"size:100;not null" json:"groupName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Permission) TableName() string {
	return "permissions"
}

type RolePermission struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	RoleCode       string    `gorm:"size:50;not null;index"`
	PermissionCode string    `gorm:"size:100;not null;index"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

type RolePermissionRequest struct {
	Permissions []string `json:"permissions"`
}
