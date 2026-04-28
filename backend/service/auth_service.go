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

// Login 处理登录逻辑
func (s *AuthService) Login(username, password string) (*model.LoginResponseData, error) {
	var user model.User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成 Access Token
	token, err := s.generateToken(user, 2*time.Hour)
	if err != nil {
		return nil, err
	}

	// 生成 Refresh Token
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

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userId uint) (*model.UserInfoResponseData, error) {
	var user model.User
	if err := DB.First(&user, userId).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	return &model.UserInfoResponseData{
		UserId:   fmt.Sprintf("%d", user.ID),
		UserName: user.Username,
		Roles:    []string{user.Role},
		Buttons:  []string{}, // 默认空按钮权限
	}, nil
}
