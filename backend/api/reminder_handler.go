package api

import (
	"backend/model"
	"backend/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReminderHandler struct {
	reminderSvc *service.ReminderService
	scheduler   *service.ReminderScheduler
}

func NewReminderHandler(reminderSvc *service.ReminderService, scheduler *service.ReminderScheduler) *ReminderHandler {
	return &ReminderHandler{reminderSvc: reminderSvc, scheduler: scheduler}
}

// List 查询当前用户某月备忘
func (h *ReminderHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	year, _ := strconv.Atoi(c.DefaultQuery("year", ""))
	month, _ := strconv.Atoi(c.DefaultQuery("month", ""))

	now := time.Now()
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}

	reminders, err := h.reminderSvc.ListByMonth(userID, year, time.Month(month))
	if err != nil {
		SendError(c, "500", "查询备忘失败: "+err.Error())
		return
	}

	SendSuccess(c, reminders)
}

// Create 创建备忘
func (h *ReminderHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	var req model.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}

	reminder, err := h.reminderSvc.Create(userID, req)
	if err != nil {
		SendError(c, "500", err.Error())
		return
	}

	// 注册定时任务（仅未来时间）
	if h.scheduler != nil && !reminder.RemindAt.Before(time.Now()) {
		h.scheduler.ScheduleReminder(*reminder)
	}

	SendSuccess(c, reminder)
}

// Update 更新备忘
func (h *ReminderHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		SendError(c, "400", "无效的备忘ID")
		return
	}

	var req model.UpdateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}

	reminder, err := h.reminderSvc.Update(userID, uint(id), req)
	if err != nil {
		SendError(c, "500", err.Error())
		return
	}

	// 更新定时任务：先移除旧的，再注册新的
	if h.scheduler != nil {
		h.scheduler.RemoveReminder(reminder.ID)
		if !reminder.Notified {
			if reminder.RemindAt.Before(time.Now()) {
				// 过去时间的重复备忘：跳过通知，计算下一次
				if reminder.RepeatType != "none" {
					go h.scheduler.HandlePastRepeat(*reminder)
				}
			} else {
				h.scheduler.ScheduleReminder(*reminder)
			}
		}
	}

	SendSuccess(c, reminder)
}

// Delete 删除备忘
func (h *ReminderHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		SendError(c, "401", "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		SendError(c, "400", "无效的备忘ID")
		return
	}

	if err := h.reminderSvc.Delete(userID, uint(id)); err != nil {
		SendError(c, "500", err.Error())
		return
	}

	// 移除定时任务
	if h.scheduler != nil {
		h.scheduler.RemoveReminder(uint(id))
	}

	SendSuccess(c, nil)
}
