package api

import (
	"backend/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ModelScenarioHandler 模型和场景处理器
type ModelScenarioHandler struct {
	svc interface {
		List(typ string) ([]model.ModelScenario, error)
		Get(id uint) (*model.ModelScenario, error)
		Create(req *model.CreateModelScenarioRequest) (*model.ModelScenario, error)
		Update(id uint, req *model.UpdateModelScenarioRequest) error
		Delete(id uint) error
	}
}

// NewModelScenarioHandler 创建处理器
func NewModelScenarioHandler(svc interface {
	List(typ string) ([]model.ModelScenario, error)
	Get(id uint) (*model.ModelScenario, error)
	Create(req *model.CreateModelScenarioRequest) (*model.ModelScenario, error)
	Update(id uint, req *model.UpdateModelScenarioRequest) error
	Delete(id uint) error
}) *ModelScenarioHandler {
	return &ModelScenarioHandler{svc: svc}
}

// HandleList 获取列表
func (h *ModelScenarioHandler) HandleList(c *gin.Context) {
	typ := c.Query("type")
	items, err := h.svc.List(typ)
	if err != nil {
		SendError(c, "500", "获取列表失败")
		return
	}
	SendSuccess(c, items)
}

// HandleCreate 创建
func (h *ModelScenarioHandler) HandleCreate(c *gin.Context) {
	var req model.CreateModelScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	item, err := h.svc.Create(&req)
	if err != nil {
		SendError(c, "500", "创建失败: "+err.Error())
		return
	}
	SendSuccess(c, item)
}

// HandleUpdate 更新
func (h *ModelScenarioHandler) HandleUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的ID")
		return
	}
	var req model.UpdateModelScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		SendError(c, "500", "更新失败")
		return
	}
	SendSuccess(c, nil)
}

// HandleDelete 删除
func (h *ModelScenarioHandler) HandleDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		SendError(c, "500", "删除失败")
		return
	}
	SendSuccess(c, nil)
}
