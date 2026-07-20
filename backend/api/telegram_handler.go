package api

import (
	"backend/service"

	"github.com/gin-gonic/gin"
)

type TelegramHandler struct {
	telegramService *service.TelegramService
	configService   *service.SystemConfigService
}

func NewTelegramHandler(telegramService *service.TelegramService, configService *service.SystemConfigService) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
		configService:   configService,
	}
}

// HandleGetTelegramConfig 获取 Telegram 配置状态（是否配置了 Bot Token）
func (h *TelegramHandler) HandleGetTelegramConfig(c *gin.Context) {
	token := h.configService.GetTelegramBotToken()
	SendSuccess(c, gin.H{"configured": token != ""})
}

// HandleGenerateBindCode 生成 Telegram 绑定码
func (h *TelegramHandler) HandleGenerateBindCode(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	result, err := h.telegramService.GenerateBindCode(userID)
	if err != nil {
		SendError(c, "1001", err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleGetTelegramStatus 获取 Telegram 绑定状态
func (h *TelegramHandler) HandleGetTelegramStatus(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	result, err := h.telegramService.GetTelegramStatus(userID)
	if err != nil {
		SendError(c, "1001", err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleUnbindTelegram 解绑 Telegram
func (h *TelegramHandler) HandleUnbindTelegram(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	if err := h.telegramService.UnbindTelegram(userID); err != nil {
		SendError(c, "1001", err.Error())
		return
	}

	SendSuccess(c, nil)
}
