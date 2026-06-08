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

	aiService, err := service.NewAIService(cfg.AI.TimeoutMinutes)
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

	customTrainingService := service.NewCustomTrainingService()
	customTrainingHandler := api.NewCustomTrainingHandler(customTrainingService)

	dashboardService := service.NewDashboardService()
	dashboardHandler := api.NewDashboardHandler(dashboardService)

	cutService := service.NewCutService()
	cutHandler := api.NewCutHandler(cutService)

	mem0Service := service.NewMem0Service(cfg.Mem0)
	mem0Handler := api.NewMem0Handler(mem0Service)

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
		// User Profile APIs
		apiGroup.GET("/user/profile", api.HandleGetUserProfile(authService))
		apiGroup.PUT("/user/profile", api.HandleUpdateProfile(authService))
		apiGroup.PUT("/user/password", api.HandleChangePassword(authService))

		historyService := service.NewHistoryService()
		historyHandler := api.NewHistoryHandler(historyService)

		apiGroup.GET("/dashboard/stats", dashboardHandler.GetStats)
		apiGroup.GET("/ai/models", api.RequirePermission("ai:model:view"), api.HandleListModels(aiService))
		apiGroup.POST("/chat", api.RequirePermission("ai:chat:send"), api.HandleChatStream(aiService, historyService, mem0Service))

		// User specific AI prompt management
		promptGroup := apiGroup.Group("/user-prompts")
		{
			promptGroup.GET("/:moduleKey", api.RequirePermission("ai:prompt:view"), promptHandler.GetUserPrompt)
			promptGroup.POST("/:moduleKey", api.RequirePermission("ai:prompt:save"), promptHandler.SaveUserPrompt)
			promptGroup.PUT("/:moduleKey/switch", api.RequirePermission("ai:prompt:switch"), promptHandler.SwitchUserPrompt)
			promptGroup.DELETE("/:moduleKey/versions/:versionId", api.RequirePermission("ai:prompt:delete"), promptHandler.HandleDeleteVersion)
			promptGroup.DELETE("/:moduleKey", api.RequirePermission("ai:prompt:reset"), promptHandler.ResetUserPrompt)
		}

		vocabGroup := apiGroup.Group("/vocabulary")
		vocabGroup.Use(api.RequirePermission("ai:vocabulary:view"))
		{
			vocabGroup.POST("", api.RequirePermission("ai:vocabulary:add"), vocabHandler.HandleAddWord)
			vocabGroup.GET("", vocabHandler.HandleListWords)
			vocabGroup.PUT("/:id", api.RequirePermission("ai:vocabulary:edit"), vocabHandler.HandleUpdateWord)
			vocabGroup.DELETE("/:id", api.RequirePermission("ai:vocabulary:delete"), vocabHandler.HandleDeleteWord)
		}

		noteGroup := apiGroup.Group("/notes")
		noteGroup.Use(api.RequirePermission("ai:note:view"))
		{
			noteGroup.POST("", api.RequirePermission("ai:note:create"), noteHandler.HandleCreateNote)
			noteGroup.GET("", noteHandler.HandleListNotes)
			noteGroup.PUT("/:id", api.RequirePermission("ai:note:edit"), noteHandler.HandleUpdateNote)
			noteGroup.DELETE("/:id", api.RequirePermission("ai:note:delete"), noteHandler.HandleDeleteNote)
		}

		historyGroup := apiGroup.Group("/histories")
		historyGroup.Use(api.RequirePermission("ai:history:view"))
		{
			historyGroup.GET("", historyHandler.ListHistory)
			historyGroup.GET("/:id", historyHandler.GetHistory)
			historyGroup.PUT("/:id/favorite", api.RequirePermission("ai:history:favorite"), historyHandler.UpdateFavorite)
			historyGroup.PUT("/:id/title", api.RequirePermission("ai:history:edit"), historyHandler.UpdateTitle)
			historyGroup.DELETE("/:id", api.RequirePermission("ai:history:delete"), historyHandler.DeleteHistory)
		}

		customTrainingGroup := apiGroup.Group("/custom-trainings")
		customTrainingGroup.Use(api.RequirePermission("ai:custom-training:view"))
		{
			customTrainingGroup.GET("", customTrainingHandler.ListCustomTrainings)
			customTrainingGroup.GET("/:id", customTrainingHandler.GetCustomTraining)
			customTrainingGroup.POST("", api.RequirePermission("ai:custom-training:create"), customTrainingHandler.CreateCustomTraining)
			customTrainingGroup.PUT("/:id", api.RequirePermission("ai:custom-training:edit"), customTrainingHandler.UpdateCustomTraining)
			customTrainingGroup.DELETE("/:id", api.RequirePermission("ai:custom-training:delete"), customTrainingHandler.DeleteCustomTraining)
		}

		// Memory APIs (mem0)
		memoryGroup := apiGroup.Group("/memories")
		{
			memoryGroup.GET("", mem0Handler.HandleListMemories)
			memoryGroup.POST("", mem0Handler.HandleAddMemory)
			memoryGroup.POST("/search", mem0Handler.HandleSearchMemories)
			memoryGroup.DELETE("/:id", mem0Handler.HandleDeleteMemory)
		}

		// Cut APIs
		cutGroup := apiGroup.Group("/cut")
		cutGroup.Use(api.RequirePermission("cut:menu:view"))
		{
			cutGroup.POST("/bar", api.RequirePermission("cut:bar:compute"), cutHandler.HandleBarCut)
			cutGroup.POST("/plane", api.RequirePermission("cut:plane:compute"), cutHandler.HandlePlaneCut)
		}

		cutRecordGroup := apiGroup.Group("/cutRecord")
		cutRecordGroup.Use(api.RequirePermission("cut:menu:view"))
		{
			cutRecordGroup.POST("/add", api.RequirePermission("cut:record:create"), cutHandler.HandleAddRecord)
			cutRecordGroup.GET("/list", api.RequirePermission("cut:record:view"), cutHandler.HandleListRecords)
			cutRecordGroup.POST("/delete/:id", api.RequirePermission("cut:record:delete"), cutHandler.HandleDeleteRecord)
		}

		adminGroup := apiGroup.Group("/admin")
		{
			// User Management
			adminGroup.GET("/users", api.RequirePermission("system:user:list"), adminHandler.HandleListUsers)
			adminGroup.POST("/users", api.RequirePermission("system:user:create"), adminHandler.HandleCreateUser)
			adminGroup.PUT("/users/:id", api.RequirePermission("system:user:update"), adminHandler.HandleUpdateUser)
			adminGroup.DELETE("/users/:id", api.RequirePermission("system:user:delete"), adminHandler.HandleDeleteUser)

			// Role Management
			adminGroup.GET("/roles", api.RequirePermission("system:role:list"), adminHandler.HandleListRoles)
			adminGroup.POST("/roles", api.RequirePermission("system:role:create"), adminHandler.HandleCreateRole)
			adminGroup.DELETE("/roles/:roleCode", api.RequirePermission("system:role:delete"), adminHandler.HandleDeleteRole)

			// Permission Management
			adminGroup.GET("/permissions", api.RequirePermission("system:permission:view"), adminHandler.HandleListPermissions)
			adminGroup.POST("/permissions", api.RequirePermission("system:permission:create"), adminHandler.HandleCreatePermission)
			adminGroup.PUT("/permissions/:id", api.RequirePermission("system:permission:update"), adminHandler.HandleUpdatePermission)
			adminGroup.DELETE("/permissions/:id", api.RequirePermission("system:permission:delete"), adminHandler.HandleDeletePermission)

			// Role Permission Management
			adminGroup.GET("/roles/:roleCode/permissions", api.RequirePermission("system:role:permission:view"), adminHandler.HandleGetRolePermissions)
			adminGroup.PUT("/roles/:roleCode/permissions", api.RequirePermission("system:role:permission:update"), adminHandler.HandleUpdateRolePermissions)

			// AI Config Management
			adminGroup.GET("/ai-providers", api.RequirePermission("system:ai-provider:view"), adminHandler.HandleListAIProviders)
			adminGroup.POST("/ai-providers", api.RequirePermission("system:ai-provider:create"), adminHandler.HandleCreateAIProvider)
			adminGroup.PUT("/ai-providers/:id", api.RequirePermission("system:ai-provider:update"), adminHandler.HandleUpdateAIProvider)
			adminGroup.DELETE("/ai-providers/:id", api.RequirePermission("system:ai-provider:delete"), adminHandler.HandleDeleteAIProvider)

			adminGroup.GET("/ai-models", api.RequirePermission("system:ai-model:view"), adminHandler.HandleListAIModels)
			adminGroup.POST("/ai-models", api.RequirePermission("system:ai-model:create"), adminHandler.HandleCreateAIModel)
			adminGroup.PUT("/ai-models/:id", api.RequirePermission("system:ai-model:update"), adminHandler.HandleUpdateAIModel)
			adminGroup.DELETE("/ai-models/:id", api.RequirePermission("system:ai-model:delete"), adminHandler.HandleDeleteAIModel)
		}
	}

	r.Run(":8080")
}
