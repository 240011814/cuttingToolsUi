package api

import (
	iface "backend/interface"
	"backend/model"
	"backend/service"

	"github.com/gin-gonic/gin"
)

type SystemConfigHandler struct {
	configSvc       *service.SystemConfigService
	telegramService *service.TelegramService
	aiAgentSvc      *service.AIAgentService
	notifier        iface.Notifier
}

func NewSystemConfigHandler(configSvc *service.SystemConfigService, telegramService *service.TelegramService, notifier iface.Notifier, aiAgentSvc ...*service.AIAgentService) *SystemConfigHandler {
	h := &SystemConfigHandler{
		configSvc:       configSvc,
		telegramService: telegramService,
		notifier:        notifier,
	}
	if len(aiAgentSvc) > 0 {
		h.aiAgentSvc = aiAgentSvc[0]
	}
	return h
}

func (h *SystemConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.configSvc.GetAll()
	if err != nil {
		SendError(c, "500", "获取配置失败: "+err.Error())
		return
	}
	SendSuccess(c, configs)
}

func (h *SystemConfigHandler) Update(c *gin.Context) {
	var req struct {
		Key    string `json:"key" binding:"required"`
		Value  string `json:"value"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	if err := h.configSvc.SetValue(req.Key, req.Value, req.Remark); err != nil {
		SendError(c, "500", "更新配置失败: "+err.Error())
		return
	}

	// 当关闭 2FA 时，清除所有用户的 TOTP 密钥，确保重新开启时需要重新绑定
	if req.Key == "admin_2fa_enabled" && req.Value == "false" {
		service.DB.Model(&model.User{}).Where("totp_secret IS NOT NULL").Update("totp_secret", nil)
	}

	// Telegram Bot Token 变更后重启 Bot
	if req.Key == "telegram_bot_token" || req.Key == "telegram_enabled" || req.Key == "telegram_webhook_url" {
		if h.telegramService != nil {
			go h.telegramService.RestartBot()
		}
	}

	// 超时配置变更后重新加载配置缓存和 AI 服务
	timeoutKeys := []string{"ai_timeout_minutes", "ai_tls_handshake_timeout", "ai_response_header_timeout", "http_timeout_seconds"}
	for _, key := range timeoutKeys {
		if req.Key == key {
			h.configSvc.ReloadTimeoutConfig()
		if h.aiAgentSvc != nil {
			go h.aiAgentSvc.ReloadConfig()
			}
			break
		}
	}

	SendSuccess(c, nil)
}

// GetRegisterStatus 公开接口，供登录页检查注册是否开启
func (h *SystemConfigHandler) GetRegisterStatus(c *gin.Context) {
	val, err := h.configSvc.GetValue("register_enabled")
	if err != nil {
		SendSuccess(c, gin.H{"enabled": true})
		return
	}
	SendSuccess(c, gin.H{"enabled": val == "true"})
}

// SendTestEmail 发送测试邮件
func (h *SystemConfigHandler) SendTestEmail(c *gin.Context) {
	var req struct {
		Email   string `json:"email" binding:"required,email"`
		Subject string `json:"subject"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请输入有效的邮箱地址")
		return
	}

	subject := req.Subject
	if subject == "" {
		subject = "测试邮件"
	}
	content := req.Content
	if content == "" {
		content = "这是一封 SMTP 邮件服务的测试邮件。\n\n如果您收到此邮件，说明 SMTP 配置正确。"
	}

	msg := iface.NotifyMessage{
		Subject: subject,
		Body:    content,
	}

	if err := h.notifier.Send(req.Email, msg); err != nil {
		SendError(c, "500", "发送测试邮件失败: "+err.Error())
		return
	}

	SendSuccess(c, gin.H{"message": "测试邮件已发送"})
}
