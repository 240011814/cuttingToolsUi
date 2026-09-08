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
	notifier     iface.Notifier
}

func NewReminderService(jobScheduler *JobScheduler, notifier iface.Notifier) *ReminderService {
	s := &ReminderService{
		jobScheduler: jobScheduler,
		notifier:     notifier,
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

	// 从调度器移除旧任务，重新注册新任务
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
	// 从调度器移除
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
	if err := DB.First(&user, *job.UserID).Error; err != nil || user.Email == "" {
		log.Printf("[ReminderService] 用户 %d 邮箱不存在，跳过", *job.UserID)
		return nil
	}

	msg := iface.NotifyMessage{
		Subject: params.Title,
		Body:    params.Content,
	}

	if err := s.notifier.Send(user.Email, msg); err != nil {
		return fmt.Errorf("发送失败: %v", err)
	}

	log.Printf("[ReminderService] 已发送: %s -> %s", params.Title, user.Email)
	return nil
}