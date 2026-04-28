package api

import (
	"net/http"
	"strings"

	"backend/model"
	"backend/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Response 统一响应结构
type Response struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func SendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: "0000",
		Data: data,
		Msg:  "success",
	})
}

func SendError(c *gin.Context, code string, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: nil,
		Msg:  msg,
	})
}

// HandleLogin 登录处理
func HandleLogin(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			SendError(c, "400", "请求参数错误: "+err.Error())
			return
		}

		res, err := authService.Login(req.UserName, req.Password)
		if err != nil {
			SendError(c, "1001", err.Error())
			return
		}

		SendSuccess(c, res)
	}
}

// HandleGetUserInfo 获取用户信息
func HandleGetUserInfo(authService *service.AuthService, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			SendError(c, "401", "未登录")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			SendError(c, "401", "无效的 Token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			SendError(c, "401", "无效的 Token Claims")
			return
		}

		userId := uint(claims["userId"].(float64))
		userInfo, err := authService.GetUserInfo(userId)
		if err != nil {
			SendError(c, "1002", err.Error())
			return
		}

		SendSuccess(c, userInfo)
	}
}
