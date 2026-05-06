package api

import (
	"backend/model"
	"backend/service"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	historyService *service.HistoryService
}

func NewHistoryHandler(historyService *service.HistoryService) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
	}
}

func (h *HistoryHandler) ListHistory(c *gin.Context) {
	var req model.ListHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, "400", "Invalid request parameters: "+err.Error())
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	histories, total, err := h.historyService.ListHistory(userID.(uint), req.Page, req.PageSize, req.Title, req.RecordType)
	if err != nil {
		SendError(c, "500", "Failed to fetch histories")
		return
	}

	SendSuccess(c, model.ListHistoryResponse{
		Total: total,
		Items: histories,
	})
}

func (h *HistoryHandler) ArchiveHistory(c *gin.Context) {
	var req model.ArchiveHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "Invalid request body: "+err.Error())
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	title := req.Title
	if title == "" {
		title = "手动归档对话"
	}

	historyID, err := h.historyService.SaveHistory(userID.(uint), 0, req.TrainingType, title, req.Messages, "manual")
	if err != nil {
		SendError(c, "500", "Failed to archive history")
		return
	}

	SendSuccess(c, gin.H{"id": historyID})
}
