package api

import (
	"strconv"

	"backend/model"
	"backend/service"
	"github.com/gin-gonic/gin"
)

type VocabularyHandler struct {
	svc *service.VocabularyService
}

func NewVocabularyHandler(svc *service.VocabularyService) *VocabularyHandler {
	return &VocabularyHandler{svc: svc}
}

// HandleAddWord 添加生词
func (h *VocabularyHandler) HandleAddWord(c *gin.Context) {
	userId := GetUserID(c)
	var req model.CreateVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	res, err := h.svc.AddWord(userId, req)
	if err != nil {
		SendError(c, "500", "保存失败: "+err.Error())
		return
	}

	SendSuccess(c, res)
}

// HandleListWords 获取生词列表
func (h *VocabularyHandler) HandleListWords(c *gin.Context) {
	userId := GetUserID(c)
	keyword := c.Query("keyword")
	isMasteredStr := c.Query("isMastered")

	var isMastered *bool
	if isMasteredStr != "" {
		b, err := strconv.ParseBool(isMasteredStr)
		if err == nil {
			isMastered = &b
		}
	}

	list, err := h.svc.GetUserVocabulary(userId, keyword, isMastered)
	if err != nil {
		SendError(c, "500", "获取列表失败: "+err.Error())
		return
	}

	SendSuccess(c, list)
}

// HandleDeleteWord 删除生词
func (h *VocabularyHandler) HandleDeleteWord(c *gin.Context) {
	userId := GetUserID(c)
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if err := h.svc.DeleteWord(userId, uint(id)); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}

// HandleUpdateWord 更新生词
func (h *VocabularyHandler) HandleUpdateWord(c *gin.Context) {
	userId := GetUserID(c)
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.UpdateVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	if err := h.svc.UpdateWord(userId, uint(id), req); err != nil {
		SendError(c, "500", "更新失败: "+err.Error())
		return
	}

	SendSuccess(c, nil)
}
