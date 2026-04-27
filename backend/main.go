package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "AI English Learning Backend is running",
		})
	})

	// AI 纠错接口占位符
	r.POST("/api/correct", func(c *gin.Context) {
		var json struct {
			Text string `json:"text" binding:"required"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 这里未来将对接 AI 接口
		c.JSON(http.StatusOK, gin.H{
			"original": json.Text,
			"correction": "This is a placeholder correction for: " + json.Text,
			"explanation": "AI integration is coming soon.",
		})
	})

	// 启动服务器
	r.Run(":8080")
}
