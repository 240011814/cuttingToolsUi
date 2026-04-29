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

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config.yaml: %v", err)
		cfg = &config.Config{}
	}

	_, err = service.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Database: %v", err)
	}

	aiService, err := service.NewAIService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize AI Service: %v", err)
	}

	authService := service.NewAuthService(cfg)
	adminService := service.NewAdminService()
	adminHandler := api.NewAdminHandler(adminService)

	vocabService := service.NewVocabularyService()
	vocabHandler := api.NewVocabularyHandler(vocabService)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "AI English Learning Backend is running",
		})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", api.HandleLogin(authService))
		authGroup.GET("/getUserInfo", api.HandleGetUserInfo(authService, cfg.Auth.JWTSecret))
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(api.AuthMiddleware(cfg.Auth.JWTSecret))
	{
		apiGroup.POST("/chat", api.HandleChatStream(aiService))

		vocabGroup := apiGroup.Group("/vocabulary")
		{
			vocabGroup.POST("", vocabHandler.HandleAddWord)
			vocabGroup.GET("", vocabHandler.HandleListWords)
			vocabGroup.PUT("/:id", vocabHandler.HandleUpdateWord)
			vocabGroup.DELETE("/:id", vocabHandler.HandleDeleteWord)
		}

		adminGroup := apiGroup.Group("/admin")
		adminGroup.Use(api.RequireRole("R_SUPER"))
		{
			adminGroup.GET("/users", adminHandler.HandleListUsers)
			adminGroup.POST("/users", adminHandler.HandleCreateUser)
			adminGroup.PUT("/users/:id", adminHandler.HandleUpdateUser)
			adminGroup.DELETE("/users/:id", adminHandler.HandleDeleteUser)
			adminGroup.GET("/roles", adminHandler.HandleListRoles)
			adminGroup.GET("/permissions", adminHandler.HandleListPermissions)
			adminGroup.GET("/roles/:roleCode/permissions", adminHandler.HandleGetRolePermissions)
			adminGroup.PUT("/roles/:roleCode/permissions", adminHandler.HandleUpdateRolePermissions)
		}
	}

	r.Run(":8080")
}
