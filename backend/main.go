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

	aiService, err := service.NewAIService()
	if err != nil {
		log.Fatalf("Failed to initialize AI Service: %v", err)
	}

	authService := service.NewAuthService(cfg)
	adminService := service.NewAdminService()
	adminHandler := api.NewAdminHandler(adminService, aiService)

	vocabService := service.NewVocabularyService()
	vocabHandler := api.NewVocabularyHandler(vocabService)

	noteService := service.NewNoteService()
	noteHandler := api.NewNoteHandler(noteService)

	promptService := service.NewPromptService(service.DB)
	promptHandler := api.NewPromptHandler(promptService)

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
		authGroup.POST("/refreshToken", api.HandleRefreshToken(authService))
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(api.AuthMiddleware(cfg.Auth.JWTSecret))
	{
		historyService := service.NewHistoryService()
		historyHandler := api.NewHistoryHandler(historyService)

		apiGroup.GET("/ai/models", api.HandleListModels(aiService))
		apiGroup.POST("/chat", api.HandleChatStream(aiService, historyService))

		// User specific AI prompt management
		apiGroup.GET("/user-prompts/:moduleKey", promptHandler.GetUserPrompt)
		apiGroup.POST("/user-prompts/:moduleKey", promptHandler.SaveUserPrompt)
		apiGroup.PUT("/user-prompts/:moduleKey/switch", promptHandler.SwitchUserPrompt)
		apiGroup.DELETE("/user-prompts/:moduleKey/versions/:versionId", promptHandler.HandleDeleteVersion)
		apiGroup.DELETE("/user-prompts/:moduleKey", promptHandler.ResetUserPrompt)

		vocabGroup := apiGroup.Group("/vocabulary")
		{
			vocabGroup.POST("", vocabHandler.HandleAddWord)
			vocabGroup.GET("", vocabHandler.HandleListWords)
			vocabGroup.PUT("/:id", vocabHandler.HandleUpdateWord)
			vocabGroup.DELETE("/:id", vocabHandler.HandleDeleteWord)
		}

		noteGroup := apiGroup.Group("/notes")
		{
			noteGroup.POST("", noteHandler.HandleCreateNote)
			noteGroup.GET("", noteHandler.HandleListNotes)
			noteGroup.PUT("/:id", noteHandler.HandleUpdateNote)
			noteGroup.DELETE("/:id", noteHandler.HandleDeleteNote)
		}

		historyGroup := apiGroup.Group("/histories")
		{
			historyGroup.GET("", historyHandler.ListHistory)
			historyGroup.GET("/:id", historyHandler.GetHistory)
			historyGroup.PUT("/:id/favorite", historyHandler.UpdateFavorite)
			historyGroup.PUT("/:id/title", historyHandler.UpdateTitle)
		}

		adminGroup := apiGroup.Group("/admin")
		adminGroup.Use(api.RequireRole("R_SUPER", "R_ADMIN"))
		{
			adminGroup.GET("/users", adminHandler.HandleListUsers)
			adminGroup.POST("/users", adminHandler.HandleCreateUser)
			adminGroup.PUT("/users/:id", adminHandler.HandleUpdateUser)
			adminGroup.DELETE("/users/:id", adminHandler.HandleDeleteUser)
			adminGroup.GET("/roles", adminHandler.HandleListRoles)
			adminGroup.POST("/roles", adminHandler.HandleCreateRole)
			adminGroup.DELETE("/roles/:roleCode", adminHandler.HandleDeleteRole)
			adminGroup.GET("/permissions", adminHandler.HandleListPermissions)
			adminGroup.POST("/permissions", adminHandler.HandleCreatePermission)
			adminGroup.PUT("/permissions/:id", adminHandler.HandleUpdatePermission)
			adminGroup.DELETE("/permissions/:id", adminHandler.HandleDeletePermission)
			adminGroup.GET("/roles/:roleCode/permissions", adminHandler.HandleGetRolePermissions)
			adminGroup.PUT("/roles/:roleCode/permissions", adminHandler.HandleUpdateRolePermissions)

			// AI Config Management
			adminGroup.GET("/ai-providers", adminHandler.HandleListAIProviders)
			adminGroup.POST("/ai-providers", adminHandler.HandleCreateAIProvider)
			adminGroup.PUT("/ai-providers/:id", adminHandler.HandleUpdateAIProvider)
			adminGroup.DELETE("/ai-providers/:id", adminHandler.HandleDeleteAIProvider)

			adminGroup.GET("/ai-models", adminHandler.HandleListAIModels)
			adminGroup.POST("/ai-models", adminHandler.HandleCreateAIModel)
			adminGroup.PUT("/ai-models/:id", adminHandler.HandleUpdateAIModel)
			adminGroup.DELETE("/ai-models/:id", adminHandler.HandleDeleteAIModel)
		}
	}

	r.Run(":8080")
}
