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

	systemConfigService := service.NewSystemConfigService()

	authService := service.NewAuthService(cfg)
	adminService := service.NewAdminService()

	vocabService := service.NewVocabularyService()
	vocabHandler := api.NewVocabularyHandler(vocabService)

	noteService := service.NewNoteService()
	noteHandler := api.NewNoteHandler(noteService)

	promptService := service.NewPromptService(service.DB, nil)

	aiAgentService, err := service.NewAIAgentService(cfg.AI.TimeoutMinutes, systemConfigService)
	if err != nil {
		log.Fatalf("Failed to initialize AI Agent Service: %v", err)
	}
	promptService = service.NewPromptService(service.DB, aiAgentService)
	aiAgentHandler := api.NewAIAgentHandler(aiAgentService)

	adminHandler := api.NewAdminHandler(adminService, aiAgentService, authService)

	dashboardService := service.NewDashboardService()
	dashboardHandler := api.NewDashboardHandler(dashboardService)

	cutService := service.NewCutService()
	cutHandler := api.NewCutHandler(cutService)

	promptHandler := api.NewPromptHandler(promptService, systemConfigService)

	// mem0 配置从数据库读取
	mem0Cfg := systemConfigService.GetMem0Config()
	timeoutConfig := systemConfigService.GetTimeoutConfig()
	mem0Service := service.NewMem0Service(mem0Cfg, timeoutConfig)
	mem0Handler := api.NewMem0Handler(mem0Service)

	// Telegram Bot
	telegramService := service.NewTelegramService(systemConfigService)
	telegramHandler := api.NewTelegramHandler(telegramService, systemConfigService)

	systemConfigHandler := api.NewSystemConfigHandler(systemConfigService, mem0Service, telegramService, aiAgentService)

	userPrefService := service.NewUserPreferenceService()
	userPrefHandler := api.NewUserPreferenceHandler(userPrefService)

	historyService := service.NewHistoryService()
	historyHandler := api.NewHistoryHandler(historyService)

	lotteryService := service.NewLotteryService()
	lotteryHandler := api.NewLotteryHandler(lotteryService)

	modelScenarioService := service.NewModelScenarioService()
	modelScenarioHandler := api.NewModelScenarioHandler(modelScenarioService)

	courseService := service.NewCourseService()
	courseHandler := api.NewCourseHandler(courseService)

	errorBookService := service.NewErrorBookService()
	errorBookHandler := api.NewErrorBookHandler(errorBookService)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "AI English Learning Backend is running",
		})
	})

	// Telegram Webhook (公开接口，无需认证)
	r.POST("/api/telegram/webhook", telegramHandler.HandleWebhook)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", api.HandleLogin(authService))
		authGroup.POST("/register", api.HandleRegister(authService))
		authGroup.GET("/getUserInfo", api.HandleGetUserInfo(authService, cfg.Auth.JWTSecret))
		authGroup.POST("/refreshToken", api.HandleRefreshToken(authService))
		authGroup.GET("/register-status", systemConfigHandler.GetRegisterStatus)

		// 2FA endpoints (temp token in Authorization header, no full auth required)
		authGroup.POST("/2fa/setup", api.Handle2FASetup(authService))
		authGroup.POST("/2fa/verify", api.Handle2FAVerify(authService))
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(api.AuthMiddleware(cfg.Auth.JWTSecret))
	{
		// User Profile APIs
		apiGroup.GET("/user/profile", api.HandleGetUserProfile(authService))
		apiGroup.PUT("/user/profile", api.HandleUpdateProfile(authService))
		apiGroup.PUT("/user/password", api.HandleChangePassword(authService))

		// User Preferences
		apiGroup.GET("/user/preferences/theme", userPrefHandler.GetThemePreference)
		apiGroup.PUT("/user/preferences/theme", userPrefHandler.SaveThemePreference)

		// Telegram Binding
		apiGroup.GET("/telegram/config", telegramHandler.HandleGetTelegramConfig)
		apiGroup.GET("/telegram/status", telegramHandler.HandleGetTelegramStatus)
		apiGroup.POST("/telegram/bind-code", telegramHandler.HandleGenerateBindCode)
		apiGroup.POST("/telegram/unbind", telegramHandler.HandleUnbindTelegram)

		apiGroup.GET("/dashboard/stats", dashboardHandler.GetStats)
		apiGroup.GET("/ai/models", api.RequirePermission("ai:model:view"), api.HandleListModels(aiAgentService))
		apiGroup.POST("/chat", api.RequirePermission("ai:chat:send"), api.HandleChatStream(aiAgentService, historyService, mem0Service))

		// User specific AI prompt management
		promptGroup := apiGroup.Group("/user-prompts")
		{
			promptGroup.GET("/:agentId", api.RequirePermission("ai:prompt:view"), promptHandler.GetUserPrompt)
			promptGroup.POST("/:agentId", api.RequirePermission("ai:prompt:save"), promptHandler.SaveUserPrompt)
			promptGroup.PUT("/:agentId/switch", api.RequirePermission("ai:prompt:switch"), promptHandler.SwitchUserPrompt)
			promptGroup.DELETE("/:agentId/versions/:versionId", api.RequirePermission("ai:prompt:delete"), promptHandler.HandleDeleteVersion)
			promptGroup.DELETE("/:agentId", api.RequirePermission("ai:prompt:reset"), promptHandler.ResetUserPrompt)
		}

		vocabGroup := apiGroup.Group("/vocabulary")
		vocabGroup.Use(api.RequirePermission("ai:vocabulary:view"))
		{
			vocabGroup.POST("", api.RequirePermission("ai:vocabulary:add"), vocabHandler.HandleAddWord)
			vocabGroup.GET("", vocabHandler.HandleListWords)
			vocabGroup.GET("/random", vocabHandler.HandleGetRandomWords)
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
		historyGroup.POST("/:id/share", api.RequirePermission("ai:history:edit"), historyHandler.GenerateShare)
		historyGroup.DELETE("/:id/share", api.RequirePermission("ai:history:edit"), historyHandler.RevokeShare)
		}

		aiAgentGroup := apiGroup.Group("/ai-agents")
		{
			aiAgentGroup.GET("", aiAgentHandler.ListAvailableAgents)
			aiAgentGroup.GET("/:id", aiAgentHandler.GetAIAgent)
			aiAgentGroup.POST("", api.RequirePermission("ai:custom-training:create"), aiAgentHandler.CreateAIAgent)
			aiAgentGroup.PUT("/:id", api.RequirePermission("ai:custom-training:edit"), aiAgentHandler.UpdateAIAgent)
			aiAgentGroup.DELETE("/:id", api.RequirePermission("ai:custom-training:delete"), aiAgentHandler.DeleteAIAgent)
		}

		// Memory APIs (mem0)
		memoryGroup := apiGroup.Group("/memories")
		{
			memoryGroup.GET("/status", mem0Handler.HandleStatus)
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

		// 模型和场景 APIs
		modelScenarioGroup := apiGroup.Group("/model-scenario")
		modelScenarioGroup.Use(api.RequirePermission("model_scenario:view"))
		{
			modelScenarioGroup.GET("", modelScenarioHandler.HandleList)
			modelScenarioGroup.POST("", api.RequirePermission("model_scenario:create"), modelScenarioHandler.HandleCreate)
			modelScenarioGroup.PUT("/:id", api.RequirePermission("model_scenario:update"), modelScenarioHandler.HandleUpdate)
			modelScenarioGroup.DELETE("/:id", api.RequirePermission("model_scenario:delete"), modelScenarioHandler.HandleDelete)
		}

		// Course APIs
		courseGroup := apiGroup.Group("/courses")
		courseGroup.Use(api.RequirePermission("ai:course:view"))
		{
			courseGroup.GET("", courseHandler.ListCourses)
			courseGroup.GET("/:id", courseHandler.GetCourse)
			courseGroup.POST("", api.RequirePermission("ai:course:create"), courseHandler.CreateCourse)
			courseGroup.PUT("/:id", api.RequirePermission("ai:course:edit"), courseHandler.UpdateCourse)
			courseGroup.DELETE("/:id", api.RequirePermission("ai:course:delete"), courseHandler.DeleteCourse)
			courseGroup.GET("/:id/items", courseHandler.GetCourseItems)
			courseGroup.POST("/:id/items", api.RequirePermission("ai:course:edit"), courseHandler.CreateCourseItem)
			courseGroup.POST("/:id/items/batch", api.RequirePermission("ai:course:edit"), courseHandler.BatchCreateCourseItems)
			courseGroup.DELETE("/:id/items/batch", api.RequirePermission("ai:course:delete"), courseHandler.BatchDeleteCourseItems)
			courseGroup.PUT("/:id/items/:itemId", api.RequirePermission("ai:course:edit"), courseHandler.UpdateCourseItem)
			courseGroup.DELETE("/:id/items/:itemId", api.RequirePermission("ai:course:edit"), courseHandler.DeleteCourseItem)
			courseGroup.GET("/:id/training", courseHandler.GetTrainingStatus)
			courseGroup.PUT("/:id/training", courseHandler.UpdateTrainingStatus)
			courseGroup.POST("/:id/training/increment", courseHandler.IncrementTrainingCount)
		}

		// Error Book APIs
		errorBookGroup := apiGroup.Group("/error-book")
		errorBookGroup.Use(api.RequirePermission("ai:error-book:view"))
		{
			errorBookGroup.POST("", api.RequirePermission("ai:error-book:add"), errorBookHandler.HandleAddErrorBook)
			errorBookGroup.GET("", errorBookHandler.HandleListErrorBooks)
			errorBookGroup.GET("/practice", api.RequirePermission("ai:error-book:practice"), errorBookHandler.HandleGetErrorBookForPractice)
			errorBookGroup.GET("/random", api.RequirePermission("ai:error-book:practice"), errorBookHandler.HandleGetRandomErrorBooks)
			errorBookGroup.GET("/stats", errorBookHandler.HandleGetErrorBookStats)
			errorBookGroup.PUT("/:id", api.RequirePermission("ai:error-book:edit"), errorBookHandler.HandleUpdateErrorBook)
			errorBookGroup.DELETE("/:id", api.RequirePermission("ai:error-book:delete"), errorBookHandler.HandleDeleteErrorBook)
		}

		// Lottery Admin APIs (需要登录+权限)
		lotteryGroup := apiGroup.Group("/lottery")
		lotteryGroup.Use(api.RequirePermission("lottery:menu:view"))
		{
			// 活动管理
			lotteryGroup.POST("/activities", api.RequirePermission("lottery:activity:create"), lotteryHandler.HandleCreateActivity)
			lotteryGroup.PUT("/activities/:id", api.RequirePermission("lottery:activity:update"), lotteryHandler.HandleUpdateActivity)
			lotteryGroup.DELETE("/activities/:id", api.RequirePermission("lottery:activity:delete"), lotteryHandler.HandleDeleteActivity)
			lotteryGroup.DELETE("/activities/:id/records", api.RequirePermission("lottery:record:delete"), lotteryHandler.HandleDeleteRecordsByActivityID)

			// 奖品管理
			lotteryGroup.POST("/activities/:id/prizes", api.RequirePermission("lottery:prize:create"), lotteryHandler.HandleCreatePrize)
			lotteryGroup.PUT("/prizes/:id", api.RequirePermission("lottery:prize:update"), lotteryHandler.HandleUpdatePrize)
			lotteryGroup.DELETE("/prizes/:id", api.RequirePermission("lottery:prize:delete"), lotteryHandler.HandleDeletePrize)

			// 记录管理
			lotteryGroup.DELETE("/records/:id", api.RequirePermission("lottery:record:delete"), lotteryHandler.HandleDeleteRecord)
		}

		adminGroup := apiGroup.Group("/admin")
		{
			// User Management
			adminGroup.GET("/users", api.RequirePermission("system:user:list"), adminHandler.HandleListUsers)
			adminGroup.POST("/users", api.RequirePermission("system:user:create"), adminHandler.HandleCreateUser)
			adminGroup.PUT("/users/:id", api.RequirePermission("system:user:update"), adminHandler.HandleUpdateUser)
			adminGroup.DELETE("/users/:id", api.RequirePermission("system:user:delete"), adminHandler.HandleDeleteUser)
			adminGroup.POST("/users/:id/proxy-login", api.RequireRole("R_SUPER"), adminHandler.HandleProxyLogin)

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
			adminGroup.POST("/ai-test", api.RequirePermission("system:ai-provider:view"), adminHandler.HandleTestAIConnection)

			adminGroup.GET("/ai-tools", api.RequirePermission("system:ai-tool:view"), adminHandler.HandleListAITools)
			adminGroup.GET("/ai-tools/:id", api.RequirePermission("system:ai-tool:view"), adminHandler.HandleGetAITool)
			adminGroup.POST("/ai-tools", api.RequirePermission("system:ai-tool:create"), adminHandler.HandleCreateAITool)
			adminGroup.PUT("/ai-tools/:id", api.RequirePermission("system:ai-tool:update"), adminHandler.HandleUpdateAITool)
			adminGroup.DELETE("/ai-tools/:id", api.RequirePermission("system:ai-tool:delete"), adminHandler.HandleDeleteAITool)

			// System Config (R_SUPER only)
			configGroup := adminGroup.Group("/system-config")
			configGroup.Use(api.RequireRole("R_SUPER"))
			{
				configGroup.GET("", systemConfigHandler.GetAll)
				configGroup.PUT("", systemConfigHandler.Update)
			}
		}
	}

	// Lottery Public APIs (无需登录)
	lotteryPublicGroup := r.Group("/api/lottery")
	{
		lotteryPublicGroup.GET("/activities", lotteryHandler.HandleListActivities)
		lotteryPublicGroup.GET("/activities/:id", lotteryHandler.HandleGetActivity)
		lotteryPublicGroup.GET("/activities/:id/prizes", lotteryHandler.HandleListPrizes)
		lotteryPublicGroup.GET("/activities/:id/limits", lotteryHandler.HandleGetDrawLimits)
		lotteryPublicGroup.POST("/draw/:activityId", lotteryHandler.HandleDraw)
		lotteryPublicGroup.GET("/records", lotteryHandler.HandleListRecords)
		lotteryPublicGroup.GET("/winners", lotteryHandler.HandleListWinners)
	}

	// Public share route (no auth required, under /api for reverse proxy compatibility)
	r.GET("/api/share/:token", api.HandleGetSharedHistory(historyService))

	// Start Telegram Bot
	go func() {
		if err := telegramService.StartBot(); err != nil {
			log.Printf("Warning: Failed to start Telegram Bot: %v", err)
		}
	}()

	r.Run(":8080")
}
