package api

import (
	"backend/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendError(c, "401", "Unauthorized")
		return
	}

	stats, err := h.dashboardService.GetStats(userID.(uint))
	if err != nil {
		SendError(c, "500", "Failed to fetch dashboard stats")
		return
	}

	SendSuccess(c, stats)
}
