package api

import (
	"backend/model"
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SkillHandler struct {
	svc *service.AISkillService
}

func NewSkillHandler(svc *service.AISkillService) *SkillHandler {
	return &SkillHandler{svc: svc}
}

// HandleList 获取 Skill 列表
func (h *SkillHandler) HandleList(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		SendError(c, "500", "获取Skill列表失败")
		return
	}
	SendSuccess(c, items)
}

// HandleCreate 创建 Skill
func (h *SkillHandler) HandleCreate(c *gin.Context) {
	var req model.CreateAISkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	item, err := h.svc.Create(&req)
	if err != nil {
		SendError(c, "500", "创建失败: "+err.Error())
		return
	}
	SendSuccess(c, item)
}

// HandleUpdate 更新 Skill
func (h *SkillHandler) HandleUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的ID")
		return
	}
	var req model.UpdateAISkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		SendError(c, "500", "更新失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

// HandleDelete 删除 Skill
func (h *SkillHandler) HandleDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}
