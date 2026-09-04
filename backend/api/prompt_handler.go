package api

import (
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PromptHandler struct {
	promptSvc *service.PromptService
}

func NewPromptHandler(promptSvc *service.PromptService) *PromptHandler {
	return &PromptHandler{promptSvc: promptSvc}
}

func (h *PromptHandler) GetUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	agentID, err := strconv.ParseUint(c.Param("agentId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的 agent ID")
		return
	}

	versions, err := h.promptSvc.ListVersions(userID, uint(agentID))
	if err != nil {
		SendError(c, "500", "获取版本列表失败: "+err.Error())
		return
	}

	effectivePrompt, _, _, err := h.promptSvc.GetEffectivePrompt(userID, uint(agentID))
	if err != nil {
		SendError(c, "500", "获取提示词失败: "+err.Error())
		return
	}

	SendSuccess(c, gin.H{
		"effective_prompt":      effectivePrompt,
		"default_prompt":       "",
		"versions":             versions,
		"is_customized":        len(versions) > 0,
	})
}

func (h *PromptHandler) SaveUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	agentID, err := strconv.ParseUint(c.Param("agentId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的 agent ID")
		return
	}

	var req struct {
		Prompt string `json:"prompt" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "无效的请求参数")
		return
	}

	if err := h.promptSvc.SaveUserPrompt(userID, uint(agentID), req.Prompt, req.Remark); err != nil {
		SendError(c, "500", "保存失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

func (h *PromptHandler) SwitchUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	agentID, err := strconv.ParseUint(c.Param("agentId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的 agent ID")
		return
	}

	var req struct {
		VersionID uint `json:"version_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "无效的请求参数")
		return
	}

	if err := h.promptSvc.SwitchVersion(userID, uint(agentID), req.VersionID); err != nil {
		SendError(c, "500", "切换失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

func (h *PromptHandler) HandleDeleteVersion(c *gin.Context) {
	userID := GetUserID(c)
	agentID, err := strconv.ParseUint(c.Param("agentId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的 agent ID")
		return
	}

	versionID, err := strconv.ParseUint(c.Param("versionId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的版本 ID")
		return
	}

	if err := h.promptSvc.DeleteVersion(userID, uint(agentID), uint(versionID)); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

func (h *PromptHandler) ResetUserPrompt(c *gin.Context) {
	userID := GetUserID(c)
	agentID, err := strconv.ParseUint(c.Param("agentId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的 agent ID")
		return
	}

	if err := h.promptSvc.ResetUserPrompt(userID, uint(agentID)); err != nil {
		SendError(c, "500", "重置失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}
