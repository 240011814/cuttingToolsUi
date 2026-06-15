package api

import (
	"backend/service"

	"github.com/gin-gonic/gin"
)

type UserPreferenceHandler struct {
	svc *service.UserPreferenceService
}

func NewUserPreferenceHandler(svc *service.UserPreferenceService) *UserPreferenceHandler {
	return &UserPreferenceHandler{svc: svc}
}

// GetThemePreference 获取当前用户的主题配置
func (h *UserPreferenceHandler) GetThemePreference(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	pref, err := h.svc.GetPreference(userID, "themeSettings")
	if err != nil {
		// 未找到返回空对象，不报错
		SendSuccess(c, nil)
		return
	}

	SendSuccess(c, pref.PrefValue)
}

// SaveThemePreference 保存当前用户的主题配置
func (h *UserPreferenceHandler) SaveThemePreference(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	var body interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	if err := h.svc.SavePreference(userID, "themeSettings", body); err != nil {
		SendError(c, "500", "保存失败")
		return
	}

	SendSuccess(c, nil)
}
