package main

import (
	"log"
	"net/http"

	"backend/api"
	"backend/config"
	"backend/service"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 允许处理 CORS 如果前后端分离且没代理 (前端目前使用 vite proxy 或是 nginx proxy, 所以按理可以不配)
	// 但如果以后需要跨域，可以在这里添加 CORS 中间件

	// 加载配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config.yaml: %v", err)
		cfg = &config.Config{}
	}

	// 初始化数据库
	_, err = service.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Database: %v", err)
	}

	// 初始化 AI Service
	aiService, err := service.NewAIService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize AI Service: %v", err)
	}

	// 初始化 Auth Service
	authService := service.NewAuthService(cfg)

	// 初始化 Vocabulary Service & Handler
	vocabService := service.NewVocabularyService()
	vocabHandler := api.NewVocabularyHandler(vocabService)

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "AI English Learning Backend is running",
		})
	})

	// 认证接口
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", api.HandleLogin(authService))
		authGroup.GET("/getUserInfo", api.HandleGetUserInfo(authService, cfg.Auth.JWTSecret))
	}

	// 业务接口 (需要鉴权)
	apiGroup := r.Group("/api")
	apiGroup.Use(api.AuthMiddleware(cfg.Auth.JWTSecret))
	{
		// AI 聊天流式接口
		apiGroup.POST("/chat", api.HandleChatStream(aiService))

		// 生词本接口
		vocabGroup := apiGroup.Group("/vocabulary")
		{
			vocabGroup.POST("", vocabHandler.HandleAddWord)
			vocabGroup.GET("", vocabHandler.HandleListWords)
			vocabGroup.PUT("/:id", vocabHandler.HandleUpdateWord)
			vocabGroup.DELETE("/:id", vocabHandler.HandleDeleteWord)
		}
	}

	// 启动服务器
	r.Run(":8080")
}
