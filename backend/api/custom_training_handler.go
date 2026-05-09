package api

import (
	"backend/model"
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomTrainingHandler struct {
	customTrainingService *service.CustomTrainingService
}

func NewCustomTrainingHandler(customTrainingService *service.CustomTrainingService) *CustomTrainingHandler {
	return &CustomTrainingHandler{
		customTrainingService: customTrainingService,
	}
}

func (h *CustomTrainingHandler) ListCustomTrainings(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	trainings, err := h.customTrainingService.ListCustomTrainings(userID.(uint))
	if err != nil {
		SendError(c, "500", "Failed to fetch custom trainings")
		return
	}

	SendSuccess(c, trainings)
}

func (h *CustomTrainingHandler) GetCustomTraining(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "Invalid training ID")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	training, err := h.customTrainingService.GetCustomTrainingByID(userID.(uint), uint(id))
	if err != nil {
		SendError(c, "404", "Training not found")
		return
	}

	SendSuccess(c, training)
}

func (h *CustomTrainingHandler) CreateCustomTraining(c *gin.Context) {
	var req model.CreateCustomTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "Invalid request body: "+err.Error())
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	training, err := h.customTrainingService.CreateCustomTraining(userID.(uint), req)
	if err != nil {
		SendError(c, "500", "Failed to create custom training")
		return
	}

	SendSuccess(c, training)
}

func (h *CustomTrainingHandler) UpdateCustomTraining(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "Invalid training ID")
		return
	}

	var req model.UpdateCustomTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "Invalid request body: "+err.Error())
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	if err := h.customTrainingService.UpdateCustomTraining(userID.(uint), uint(id), req); err != nil {
		SendError(c, "500", "Failed to update custom training")
		return
	}

	SendSuccess(c, nil)
}

func (h *CustomTrainingHandler) DeleteCustomTraining(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "Invalid training ID")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	if err := h.customTrainingService.DeleteCustomTraining(userID.(uint), uint(id)); err != nil {
		SendError(c, "500", "Failed to delete custom training")
		return
	}

	SendSuccess(c, nil)
}
