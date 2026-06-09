package api

import (
	"backend/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type PromptHandler struct {
	promptSvc *service.PromptService
}

func NewPromptHandler(promptSvc *service.PromptService) *PromptHandler {
	return &PromptHandler{promptSvc: promptSvc}
}

// GetUserPrompt 获取用户针对某个模块的当前生效提示词及所有版本
func (h *PromptHandler) GetUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	moduleKey := c.Param("moduleKey")

	// 1. 获取所有版本
	versions, err := h.promptSvc.ListVersions(userID, moduleKey)
	if err != nil {
		SendError(c, "500", "获取版本列表失败: "+err.Error())
		return
	}

	// 2. 获取生效提示词
	effectivePrompt, memorySearchQuery, err := h.promptSvc.GetEffectivePrompt(userID, moduleKey)
	if err != nil {
		SendError(c, "500", "获取提示词失败: "+err.Error())
		return
	}

	SendSuccess(c, gin.H{
		"effective_prompt":      effectivePrompt,
		"memory_search_query":  memorySearchQuery,
		"default_prompt":       "",
		"versions":             versions,
		"is_customized":        len(versions) > 0,
	})
}

// SaveUserPrompt 保存用户的自定义提示词（存为新版本）
func (h *PromptHandler) SaveUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	moduleKey := c.Param("moduleKey")

	var req struct {
		Prompt            string `json:"prompt" binding:"required"`
		Remark            string `json:"remark"`
		MemorySearchQuery string `json:"memory_search_query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "无效的请求参数")
		return
	}

	if err := h.promptSvc.SaveUserPrompt(userID, moduleKey, req.Prompt, req.Remark, req.MemorySearchQuery); err != nil {
		SendError(c, "500", "保存失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

// SwitchUserPrompt 切换启用的提示词版本
func (h *PromptHandler) SwitchUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	moduleKey := c.Param("moduleKey")

	var req struct {
		VersionID uint `json:"version_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "无效的请求参数")
		return
	}

	if err := h.promptSvc.SwitchVersion(userID, moduleKey, req.VersionID); err != nil {
		SendError(c, "500", "切换失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

// HandleDeleteVersion 删除特定的提示词版本
func (h *PromptHandler) HandleDeleteVersion(c *gin.Context) {
	userID := GetUserID(c)
	moduleKey := c.Param("moduleKey")

	versionID, err := strconv.ParseUint(c.Param("versionId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的版本 ID")
		return
	}

	if err := h.promptSvc.DeleteVersion(userID, moduleKey, uint(versionID)); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

// ResetUserPrompt 重置提示词到系统默认
func (h *PromptHandler) ResetUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	moduleKey := c.Param("moduleKey")

	if err := h.promptSvc.ResetUserPrompt(userID, moduleKey); err != nil {
		SendError(c, "500", "重置失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}
