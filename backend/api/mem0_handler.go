package api

import (
	"backend/service"

	"github.com/gin-gonic/gin"
)

type Mem0Handler struct {
	mem0Service *service.Mem0Service
}

func NewMem0Handler(mem0Service *service.Mem0Service) *Mem0Handler {
	return &Mem0Handler{
		mem0Service: mem0Service,
	}
}

func (h *Mem0Handler) HandleSearchMemories(c *gin.Context) {
	if !h.mem0Service.IsConfigured() {
		SendError(c, "503", "记忆服务未配置")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	var req struct {
		Query string `json:"query" binding:"required"`
		TopK  int    `json:"top_k"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", err.Error())
		return
	}

	memories, err := h.mem0Service.SearchMemories(userID.(uint), req.Query, req.TopK)
	if err != nil {
		SendError(c, "500", "搜索记忆失败: "+err.Error())
		return
	}

	SendSuccess(c, memories)
}

func (h *Mem0Handler) HandleListMemories(c *gin.Context) {
	if !h.mem0Service.IsConfigured() {
		SendError(c, "503", "记忆服务未配置")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	memories, err := h.mem0Service.ListMemories(userID.(uint))
	if err != nil {
		SendError(c, "500", "获取记忆列表失败: "+err.Error())
		return
	}

	SendSuccess(c, memories)
}

func (h *Mem0Handler) HandleDeleteMemory(c *gin.Context) {
	if !h.mem0Service.IsConfigured() {
		SendError(c, "503", "记忆服务未配置")
		return
	}

	memoryID := c.Param("id")
	if memoryID == "" {
		SendError(c, "400", "记忆ID不能为空")
		return
	}

	if err := h.mem0Service.DeleteMemory(memoryID); err != nil {
		SendError(c, "500", "删除记忆失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}
