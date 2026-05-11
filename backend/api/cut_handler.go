package api

import (
	"backend/model"
	"backend/service"

	"github.com/gin-gonic/gin"
)

type CutHandler struct {
	svc *service.CutService
}

func NewCutHandler(svc *service.CutService) *CutHandler {
	return &CutHandler{svc: svc}
}

// HandleBarCut 一维切割
func (h *CutHandler) HandleBarCut(c *gin.Context) {
	var req model.BarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	result, err := h.svc.BarCut(req)
	if err != nil {
		SendError(c, "500", "切割失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandlePlaneCut 平面切割
func (h *CutHandler) HandlePlaneCut(c *gin.Context) {
	var req model.BinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	result, err := h.svc.PlaneCut(req)
	if err != nil {
		SendError(c, "500", "切割失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleAddRecord 添加切割记录
func (h *CutHandler) HandleAddRecord(c *gin.Context) {
	userID := GetUserID(c)
	var req model.RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	record, err := h.svc.SaveCutRecord(userID, req)
	if err != nil {
		SendError(c, "500", "保存失败: "+err.Error())
		return
	}

	SendSuccess(c, record)
}

// HandleListRecords 查询切割记录列表
func (h *CutHandler) HandleListRecords(c *gin.Context) {
	userID := GetUserID(c)
	var params model.CutRecordSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	result, err := h.svc.ListCutRecords(userID, params)
	if err != nil {
		SendError(c, "500", "查询失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleDeleteRecord 删除切割记录
func (h *CutHandler) HandleDeleteRecord(c *gin.Context) {
	userID := GetUserID(c)
	id := c.Param("id")

	if id == "" {
		SendError(c, "400", "记录ID不能为空")
		return
	}

	if err := h.svc.DeleteCutRecord(userID, id); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}

	SendSuccess(c, true)
}
