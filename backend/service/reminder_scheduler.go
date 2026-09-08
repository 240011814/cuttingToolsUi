package service

import (
	"backend/model"
	iface "backend/interface"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type ReminderScheduler struct {
	scheduler   gocron.Scheduler
	notifier    iface.Notifier
	reminderSvc *ReminderService
	mu          sync.Mutex
	jobMap      map[uint]gocron.Job // reminderID -> job
}

func NewReminderScheduler(notifier iface.Notifier, reminderSvc *ReminderService) (*ReminderScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	rs := &ReminderScheduler{
		scheduler:   s,
		notifier:    notifier,
		reminderSvc: reminderSvc,
		jobMap:      make(map[uint]gocron.Job),
	}

	return rs, nil
}

func (rs *ReminderScheduler) Start() {
	rs.scheduler.Start()
	log.Println("[ReminderScheduler] 已启动")
}

func (rs *ReminderScheduler) GetScheduler() gocron.Scheduler {
	return rs.scheduler
}

func (rs *ReminderScheduler) Shutdown() error {
	return rs.scheduler.Shutdown()
}

// LoadAll 启动时加载所有未通知的备忘，注册定时任务
func (rs *ReminderScheduler) LoadAll() error {
	var reminders []model.Reminder
	if err := DB.Where("notified = false AND remind_at > ?", time.Now()).Find(&reminders).Error; err != nil {
		return err
	}

	for _, r := range reminders {
		rs.scheduleReminder(r)
	}

	log.Printf("[ReminderScheduler] 已加载 %d 个备忘任务", len(reminders))
	return nil
}

// ScheduleReminder 创建备忘时调用，注册定时任务
func (rs *ReminderScheduler) ScheduleReminder(reminder model.Reminder) {
	rs.scheduleReminder(reminder)
}

// RemoveReminder 删除备忘时调用，移除定时任务
func (rs *ReminderScheduler) RemoveReminder(reminderID uint) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if job, ok := rs.jobMap[reminderID]; ok {
		rs.scheduler.RemoveJob(job.ID())
		delete(rs.jobMap, reminderID)
	}
}

func (rs *ReminderScheduler) scheduleReminder(r model.Reminder) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// 如果已有任务，先移除
	if existingJob, ok := rs.jobMap[r.ID]; ok {
		rs.scheduler.RemoveJob(existingJob.ID())
		delete(rs.jobMap, r.ID)
	}

	// 如果已过期，直接处理
	if r.RemindAt.Before(time.Now()) {
		go rs.processReminder(r)
		return
	}

	delay := time.Until(r.RemindAt)
	job, err := rs.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(r.RemindAt),
		),
		gocron.NewTask(rs.processReminder, r),
		gocron.WithName("备忘:"+r.Title),
		gocron.WithTags("reminder", "reminder:"+fmt.Sprintf("%d", r.ID)),
	)
	if err != nil {
		log.Printf("[ReminderScheduler] 注册任务失败 (reminder=%d): %v", r.ID, err)
		return
	}

	rs.jobMap[r.ID] = job
	log.Printf("[ReminderScheduler] 已注册任务: %s, %v 后执行", r.Title, delay)
}

func (rs *ReminderScheduler) processReminder(r model.Reminder) {
	log.Printf("[ReminderScheduler] 执行备忘任务: %s (reminder=%d)", r.Title, r.ID)

	// 查询用户邮箱
	var user model.User
	if err := DB.First(&user, r.UserID).Error; err != nil || user.Email == "" {
		log.Printf("[ReminderScheduler] 用户 %d 邮箱不存在，跳过", r.UserID)
		rs.reminderSvc.MarkNotified(r.ID)
		return
	}

	// 发送邮件
	msg := iface.NotifyMessage{
		Subject: r.Title,
		Body:    r.Content,
	}

	if err := rs.notifier.Send(user.Email, msg); err != nil {
		log.Printf("[ReminderScheduler] 发送失败 (reminder=%d): %v", r.ID, err)
		return
	}

	rs.reminderSvc.MarkNotified(r.ID)
	log.Printf("[ReminderScheduler] 已发送: %s -> %s", r.Title, user.Email)

	// 创建下一次重复备忘
	if r.RepeatType != "none" {
		nextAt := calcNextTime(r.RemindAt, r.RepeatType, r.RepeatInterval)
		if !nextAt.IsZero() && (r.RepeatEndAt == nil || !nextAt.After(*r.RepeatEndAt)) {
			newReminder := model.Reminder{
				UserID:         r.UserID,
				Title:          r.Title,
				Content:        r.Content,
				RemindAt:       nextAt,
				RepeatType:     r.RepeatType,
				RepeatInterval: r.RepeatInterval,
				RepeatEndAt:    r.RepeatEndAt,
			}
			if err := DB.Create(&newReminder).Error; err != nil {
				log.Printf("[ReminderScheduler] 创建下次备忘失败: %v", err)
				return
			}
			// 注册新任务
			rs.scheduleReminder(newReminder)
		}
	}

	// 从 map 中移除已执行的任务
	rs.mu.Lock()
	delete(rs.jobMap, r.ID)
	rs.mu.Unlock()
}

func calcNextTime(at time.Time, repeatType string, interval int) time.Time {
	switch repeatType {
	case "daily":
		return at.AddDate(0, 0, interval)
	case "weekly":
		return at.AddDate(0, 0, 7*interval)
	case "monthly":
		return at.AddDate(0, interval, 0)
	case "yearly":
		return at.AddDate(interval, 0, 0)
	default:
		return time.Time{}
	}
}
