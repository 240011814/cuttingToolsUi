package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			SendError(c, "401", "未登录")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			SendError(c, "401", "无效的 Token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			SendError(c, "401", "无效的 Token Claims")
			c.Abort()
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			SendError(c, "401", "无效的用户信息")
			c.Abort()
			return
		}

		c.Set("userId", uint(userID))
		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		}
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleValue, _ := role.(string)
		if _, ok := allowed[roleValue]; !ok {
			SendError(c, "403", "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	val, exists := c.Get("userId")
	if !exists {
		return 0
	}
	return val.(uint)
}
