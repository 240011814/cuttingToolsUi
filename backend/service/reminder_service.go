package service

import (
	"backend/model"
	iface "backend/interface"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

type ReminderService struct {
	jobScheduler *JobScheduler
	notifiers    []iface.Notifier
}

func NewReminderService(jobScheduler *JobScheduler, notifiers ...iface.Notifier) *ReminderService {
	s := &ReminderService{
		jobScheduler: jobScheduler,
		notifiers:    notifiers,
	}
	jobScheduler.RegisterCallback(model.JobTypeReminder, s.processReminder)
	return s
}

func (s *ReminderService) ListByMonth(userID uint, year int, month time.Month) ([]model.Job, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var jobs []model.Job
	err := DB.Where("user_id = ? AND job_type = ? AND scheduled_at >= ? AND scheduled_at < ?",
		userID, model.JobTypeReminder, start, end).
		Order("scheduled_at ASC").
		Find(&jobs).Error
	return jobs, err
}

func (s *ReminderService) GetByID(userID, id uint) (*model.Job, error) {
	var job model.Job
	err := DB.Where("user_id = ? AND id = ? AND job_type = ?", userID, id, model.JobTypeReminder).First(&job).Error
	if err != nil {
		return nil, errors.New("备忘不存在")
	}
	return &job, nil
}

func (s *ReminderService) Create(userID uint, req model.CreateReminderRequest) (*model.Job, error) {
	repeatType := req.RepeatType
	if repeatType == "" {
		repeatType = "none"
	}
	repeatInterval := req.RepeatInterval
	if repeatInterval <= 0 {
		repeatInterval = 1
	}

	advanceMinutes := req.AdvanceMinutes
	if advanceMinutes < 0 {
		advanceMinutes = 0
	}

	params := model.ReminderParams{
		Title:   req.Title,
		Content: req.Content,
	}
	paramsJSON, _ := json.Marshal(params)

	job := model.Job{
		JobType:        model.JobTypeReminder,
		JobID:          0,
		UserID:         &userID,
		ScheduledAt:    req.RemindAt,
		AdvanceMinutes: advanceMinutes,
		Status:         model.JobStatusPending,
		MaxRetries:     3,
		Params:         paramsJSON,
		RepeatType:     repeatType,
		RepeatInterval: repeatInterval,
		RepeatEndAt:    req.RepeatEndAt,
	}

	if err := DB.Create(&job).Error; err != nil {
		return nil, errors.New("创建备忘失败: " + err.Error())
	}

	DB.Model(&job).Update("job_id", job.ID)

	if !job.ScheduledAt.Before(time.Now()) {
		s.jobScheduler.scheduleJob(job)
	}

	return &job, nil
}

func (s *ReminderService) Update(userID, id uint, req model.UpdateReminderRequest) (*model.Job, error) {
	var job model.Job
	if err := DB.Where("user_id = ? AND id = ? AND job_type = ?", userID, id, model.JobTypeReminder).First(&job).Error; err != nil {
		return nil, errors.New("备忘不存在")
	}

	var params model.ReminderParams
	if err := json.Unmarshal(job.Params, &params); err != nil {
		return nil, errors.New("解析参数失败")
	}

	if req.Title != "" {
		params.Title = req.Title
	}
	if req.Content != "" {
		params.Content = req.Content
	}

	paramsJSON, _ := json.Marshal(params)

	updates := map[string]interface{}{
		"params": paramsJSON,
	}
	if !req.RemindAt.IsZero() {
		updates["scheduled_at"] = req.RemindAt
	}
	if req.AdvanceMinutes != nil {
		advanceMinutes := *req.AdvanceMinutes
		if advanceMinutes < 0 {
			advanceMinutes = 0
		}
		updates["advance_minutes"] = advanceMinutes
	}
	if req.RepeatType != "" {
		updates["repeat_type"] = req.RepeatType
	}
	if req.RepeatInterval > 0 {
		updates["repeat_interval"] = req.RepeatInterval
	}
	updates["repeat_end_at"] = req.RepeatEndAt

	if err := DB.Model(&job).Updates(updates).Error; err != nil {
		return nil, errors.New("更新备忘失败: " + err.Error())
	}

	DB.First(&job, job.ID)

	s.jobScheduler.UnscheduleJob(job.ID, model.JobTypeReminder)
	if !job.ScheduledAt.Before(time.Now()) {
		s.jobScheduler.scheduleJob(job)
	}

	return &job, nil
}

func (s *ReminderService) Delete(userID, id uint) error {
	result := DB.Where("user_id = ? AND id = ? AND job_type = ?", userID, id, model.JobTypeReminder).Delete(&model.Job{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("备忘不存在")
	}
	s.jobScheduler.UnscheduleJob(id, model.JobTypeReminder)
	return nil
}

func (s *ReminderService) processReminder(job model.Job) error {
	var params model.ReminderParams
	if err := json.Unmarshal(job.Params, &params); err != nil {
		return fmt.Errorf("解析参数失败: %v", err)
	}

	log.Printf("[ReminderService] 执行备忘任务: %s (job=%d)", params.Title, job.ID)

	if job.UserID == nil {
		return fmt.Errorf("用户ID为空")
	}

	var user model.User
	if err := DB.First(&user, *job.UserID).Error; err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}

	// 获取用户通知偏好
	channels := s.getUserNotificationChannels(user.ID)

	msg := iface.NotifyMessage{
		Subject: params.Title,
		Body:    params.Content,
	}

	var lastErr error
	for _, notifier := range s.notifiers {
		// 检查该通知渠道是否启用
		if !s.isChannelEnabled(channels, notifier.Name()) {
			continue
		}

		to := s.getRecipientForNotifier(notifier.Name(), &user)
		if to == "" {
			log.Printf("[ReminderService] 用户 %d 没有 %s 渠道的接收地址，跳过", user.ID, notifier.Name())
			continue
		}

		if err := notifier.Send(to, msg); err != nil {
			log.Printf("[ReminderService] 通过 %s 发送失败: %v", notifier.Name(), err)
			lastErr = err
		} else {
			log.Printf("[ReminderService] 已通过 %s 发送: %s -> %s", notifier.Name(), params.Title, to)
		}
	}

	return lastErr
}

// getUserNotificationChannels 获取用户的通知渠道偏好
func (s *ReminderService) getUserNotificationChannels(userID uint) []string {
	var pref model.UserPreference
	err := DB.Where("user_id = ? AND pref_key = ?", userID, "notification_channels").First(&pref).Error
	if err != nil {
		// 默认使用邮件
		return []string{"email"}
	}

	var channels []string
	if err := json.Unmarshal(pref.PrefValue, &channels); err != nil {
		return []string{"email"}
	}

	if len(channels) == 0 {
		return []string{"email"}
	}

	return channels
}

// isChannelEnabled 检查渠道是否在启用列表中
func (s *ReminderService) isChannelEnabled(channels []string, channelName string) bool {
	for _, ch := range channels {
		if ch == channelName {
			return true
		}
	}
	return false
}

// getRecipientForNotifier 获取通知的接收地址
func (s *ReminderService) getRecipientForNotifier(notifierName string, user *model.User) string {
	switch notifierName {
	case "email":
		return user.Email
	case "telegram":
		if user.TelegramChatID != nil {
			return fmt.Sprintf("%d", *user.TelegramChatID)
		}
		return ""
	default:
		return ""
	}
}
