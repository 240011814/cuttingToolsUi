package service

import (
	"errors"
	"fmt"
	"time"

	"backend/config"
	"backend/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Login(username, password string) (*model.LoginResponseData, error) {
	var user model.User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := s.generateToken(user, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponseData{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateToken(user model.User, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"userId":   user.ID,
		"userName": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(duration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Auth.JWTSecret))
}

func (s *AuthService) GetUserInfo(userId uint) (*model.UserInfoResponseData, error) {
	var user model.User
	if err := DB.First(&user, userId).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	permissions, err := s.getPermissionsByRole(user.Role)
	if err != nil {
		return nil, err
	}

	return &model.UserInfoResponseData{
		UserId:      fmt.Sprintf("%d", user.ID),
		UserName:    user.Username,
		Roles:       []string{user.Role},
		Buttons:     permissions,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) getPermissionsByRole(role string) ([]string, error) {
	if role == "R_SUPER" {
		var permissions []model.Permission
		if err := DB.Find(&permissions).Error; err != nil {
			return nil, err
		}
		codes := make([]string, 0, len(permissions))
		for _, permission := range permissions {
			codes = append(codes, permission.Code)
		}
		return codes, nil
	}

	var rows []model.RolePermission
	if err := DB.Where("role_code = ?", role).Find(&rows).Error; err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.PermissionCode)
	}
	return codes, nil
}
