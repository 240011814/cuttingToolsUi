package api

import (
	"backend/model"
	"backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// ListCourses 获取课程包列表
func (h *CourseHandler) ListCourses(c *gin.Context) {
	userID := c.GetUint("userId")
	showAll := c.Query("show_all") == "true"
	keyword := c.Query("keyword")
	tag := c.Query("tag")

	var isPublic *bool
	if v := c.Query("is_public"); v != "" {
		b := v == "true"
		isPublic = &b
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 12
	}

	courses, total, err := h.courseService.ListCourses(userID, showAll, keyword, isPublic, tag, page, pageSize)
	if err != nil {
		SendError(c, "500", "获取课程包列表失败")
		return
	}
	SendSuccess(c, gin.H{
		"list":      courses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCourse 获取课程包详情（包含条目）
func (h *CourseHandler) GetCourse(c *gin.Context) {
	userID := c.GetUint("userId")
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	course, err := h.courseService.GetCourseByID(userID, uint(courseID))
	if err != nil {
		SendError(c, "404", err.Error())
		return
	}
	// 获取课程条目
	items, err := h.courseService.GetCourseItems(uint(courseID))
	if err != nil {
		SendError(c, "500", "获取课程条目失败")
		return
	}
	SendSuccess(c, model.CourseResponse{
		Course: *course,
		Items:  items,
	})
}

// CreateCourse 创建课程包
func (h *CourseHandler) CreateCourse(c *gin.Context) {
	userID := c.GetUint("userId")
	var req model.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	course, err := h.courseService.CreateCourse(userID, req)
	if err != nil {
		SendError(c, "500", "创建课程包失败")
		return
	}
	SendSuccess(c, course)
}

// UpdateCourse 更新课程包
func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	userID := c.GetUint("userId")
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	var req model.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	course, err := h.courseService.UpdateCourse(userID, uint(courseID), req)
	if err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, course)
}

// DeleteCourse 删除课程包
func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	userID := c.GetUint("userId")
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	if err := h.courseService.DeleteCourse(userID, uint(courseID)); err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, nil)
}

// GetCourseItems 获取课程条目列表
func (h *CourseHandler) GetCourseItems(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	items, err := h.courseService.GetCourseItems(uint(courseID))
	if err != nil {
		SendError(c, "500", "获取课程条目失败")
		return
	}
	SendSuccess(c, items)
}

// CreateCourseItem 创建课程条目
func (h *CourseHandler) CreateCourseItem(c *gin.Context) {
	userID := c.GetUint("userId")
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	var req model.CreateCourseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	item, err := h.courseService.CreateCourseItem(userID, uint(courseID), req)
	if err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, item)
}

// UpdateCourseItem 更新课程条目
func (h *CourseHandler) UpdateCourseItem(c *gin.Context) {
	userID := c.GetUint("userId")
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的条目ID")
		return
	}
	var req model.UpdateCourseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	item, err := h.courseService.UpdateCourseItem(userID, uint(itemID), req)
	if err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, item)
}

// DeleteCourseItem 删除课程条目
func (h *CourseHandler) DeleteCourseItem(c *gin.Context) {
	userID := c.GetUint("userId")
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的条目ID")
		return
	}
	if err := h.courseService.DeleteCourseItem(userID, uint(itemID)); err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, nil)
}

// BatchCreateCourseItems 批量创建课程条目
func (h *CourseHandler) BatchCreateCourseItems(c *gin.Context) {
	userID := c.GetUint("userId")
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	var req model.BatchCreateCourseItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	items, err := h.courseService.BatchCreateCourseItems(userID, uint(courseID), req.Items)
	if err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, items)
}

// BatchDeleteCourseItems 批量删除课程条目
func (h *CourseHandler) BatchDeleteCourseItems(c *gin.Context) {
	userID := c.GetUint("userId")
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		SendError(c, "400", "无效的课程包ID")
		return
	}
	var req model.BatchDeleteCourseItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}
	if err := h.courseService.BatchDeleteCourseItems(userID, uint(courseID), req.Ids); err != nil {
		SendError(c, "500", err.Error())
		return
	}
	SendSuccess(c, nil)
}
