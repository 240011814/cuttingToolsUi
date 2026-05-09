package api

import (
	"backend/model"
	"backend/service"
	"strconv"

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

func (h *HistoryHandler) GetHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "Invalid history ID")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	history, err := h.historyService.GetHistoryByID(userID.(uint), uint(id))
	if err != nil {
		SendError(c, "404", "History not found")
		return
	}

	SendSuccess(c, history)
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

	histories, total, err := h.historyService.ListHistory(userID.(uint), req.Page, req.PageSize, req.Title, req.IsFavorite)
	if err != nil {
		SendError(c, "500", "Failed to fetch histories")
		return
	}

	SendSuccess(c, model.ListHistoryResponse{
		Total: total,
		Items: histories,
	})
}

func (h *HistoryHandler) UpdateFavorite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "Invalid history ID")
		return
	}

	var req model.UpdateFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "Invalid request body: "+err.Error())
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	if err := h.historyService.UpdateFavorite(userID.(uint), uint(id), req.IsFavorite); err != nil {
		SendError(c, "500", "Failed to update favorite status")
		return
	}

	SendSuccess(c, nil)
}

func (h *HistoryHandler) UpdateTitle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "Invalid history ID")
		return
	}

	var req model.UpdateTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "Invalid request body: "+err.Error())
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	if err := h.historyService.UpdateTitle(userID.(uint), uint(id), req.Title); err != nil {
		SendError(c, "500", "Failed to update title")
		return
	}

	SendSuccess(c, nil)
}
