package service

import (
	"backend/model"
	"errors"
	"time"
)

type ReminderService struct{}

func NewReminderService() *ReminderService {
	return &ReminderService{}
}

func (s *ReminderService) ListByMonth(userID uint, year int, month time.Month) ([]model.Reminder, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var reminders []model.Reminder
	err := DB.Where("user_id = ? AND remind_at >= ? AND remind_at < ?", userID, start, end).
		Order("remind_at ASC").
		Find(&reminders).Error
	return reminders, err
}

func (s *ReminderService) GetByID(userID, id uint) (*model.Reminder, error) {
	var reminder model.Reminder
	err := DB.Where("user_id = ? AND id = ?", userID, id).First(&reminder).Error
	if err != nil {
		return nil, errors.New("备忘不存在")
	}
	return &reminder, nil
}

func (s *ReminderService) Create(userID uint, req model.CreateReminderRequest) (*model.Reminder, error) {
	repeatType := req.RepeatType
	if repeatType == "" {
		repeatType = "none"
	}
	repeatInterval := req.RepeatInterval
	if repeatInterval <= 0 {
		repeatInterval = 1
	}

	reminder := model.Reminder{
		UserID:         userID,
		Title:          req.Title,
		Content:        req.Content,
		RemindAt:       req.RemindAt,
		RepeatType:     repeatType,
		RepeatInterval: repeatInterval,
		RepeatEndAt:    req.RepeatEndAt,
	}

	if err := DB.Create(&reminder).Error; err != nil {
		return nil, errors.New("创建备忘失败: " + err.Error())
	}
	return &reminder, nil
}

func (s *ReminderService) Update(userID, id uint, req model.UpdateReminderRequest) (*model.Reminder, error) {
	var reminder model.Reminder
	if err := DB.Where("user_id = ? AND id = ?", userID, id).First(&reminder).Error; err != nil {
		return nil, errors.New("备忘不存在")
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if !req.RemindAt.IsZero() {
		updates["remind_at"] = req.RemindAt
		updates["notified"] = false
	}
	if req.RepeatType != "" {
		updates["repeat_type"] = req.RepeatType
	}
	if req.RepeatInterval > 0 {
		updates["repeat_interval"] = req.RepeatInterval
	}
	updates["repeat_end_at"] = req.RepeatEndAt

	if err := DB.Model(&reminder).Updates(updates).Error; err != nil {
		return nil, errors.New("更新备忘失败: " + err.Error())
	}

	DB.First(&reminder, reminder.ID)
	return &reminder, nil
}

func (s *ReminderService) Delete(userID, id uint) error {
	result := DB.Where("user_id = ? AND id = ?", userID, id).Delete(&model.Reminder{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("备忘不存在")
	}
	return nil
}

// MarkNotified 标记为已通知
func (s *ReminderService) MarkNotified(id uint) error {
	return DB.Model(&model.Reminder{}).Where("id = ?", id).Update("notified", true).Error
}
