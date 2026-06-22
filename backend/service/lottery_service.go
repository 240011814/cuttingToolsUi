package service

import (
	"errors"
	"math/rand"
	"time"

	"backend/model"

	"gorm.io/gorm"
)

type LotteryService struct{}

func NewLotteryService() *LotteryService {
	return &LotteryService{}
}

// ==================== 活动管理 ====================

// CreateActivity 创建抽奖活动
func (s *LotteryService) CreateActivity(userID uint, req model.CreateLotteryActivityRequest) (*model.LotteryActivity, error) {
	activity := model.LotteryActivity{
		Name:            req.Name,
		Description:     req.Description,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		DrawMode:        req.DrawMode,
		MaxParticipants: req.MaxParticipants,
		DailyLimit:      req.DailyLimit,
		TotalLimit:      req.TotalLimit,
		CreatedBy:       userID,
		Status:          0,
	}

	if err := DB.Create(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetActivity 获取抽奖活动详情
func (s *LotteryService) GetActivity(id uint) (*model.LotteryActivity, error) {
	var activity model.LotteryActivity
	if err := DB.First(&activity, id).Error; err != nil {
		return nil, err
	}

	// 根据当前时间自动更新状态
	now := time.Now()
	var newStatus int
	if now.Before(activity.StartTime) {
		newStatus = 0 // 未开始
	} else if now.After(activity.EndTime) {
		newStatus = 2 // 已结束
	} else {
		newStatus = 1 // 进行中
	}

	// 如果状态变化，更新数据库
	if activity.Status != newStatus {
		DB.Model(&model.LotteryActivity{}).Where("id = ?", activity.ID).Update("status", newStatus)
		activity.Status = newStatus
	}

	return &activity, nil
}

// ListActivities 获取抽奖活动列表
func (s *LotteryService) ListActivities(keyword string, status *int) ([]model.LotteryActivity, error) {
	var activities []model.LotteryActivity
	query := DB.Model(&model.LotteryActivity{})

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("created_at DESC").Find(&activities).Error; err != nil {
		return nil, err
	}

	// 根据当前时间自动更新状态
	now := time.Now()
	for i := range activities {
		var newStatus int
		if now.Before(activities[i].StartTime) {
			newStatus = 0 // 未开始
		} else if now.After(activities[i].EndTime) {
			newStatus = 2 // 已结束
		} else {
			newStatus = 1 // 进行中
		}

		// 如果状态变化，更新数据库
		if activities[i].Status != newStatus {
			DB.Model(&model.LotteryActivity{}).Where("id = ?", activities[i].ID).Update("status", newStatus)
			activities[i].Status = newStatus
		}
	}

	return activities, nil
}

// UpdateActivity 更新抽奖活动
func (s *LotteryService) UpdateActivity(id uint, req model.UpdateLotteryActivityRequest) error {
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if !req.StartTime.IsZero() {
		updates["start_time"] = req.StartTime
	}
	if !req.EndTime.IsZero() {
		updates["end_time"] = req.EndTime
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.DrawMode != nil {
		updates["draw_mode"] = *req.DrawMode
	}
	if req.MaxParticipants != nil {
		updates["max_participants"] = *req.MaxParticipants
	}
	if req.DailyLimit != nil {
		updates["daily_limit"] = *req.DailyLimit
	}
	if req.TotalLimit != nil {
		updates["total_limit"] = *req.TotalLimit
	}

	return DB.Model(&model.LotteryActivity{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteActivity 删除抽奖活动
func (s *LotteryService) DeleteActivity(id uint) error {
	// 检查是否有抽奖记录
	var count int64
	DB.Model(&model.LotteryRecord{}).Where("activity_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该活动已有抽奖记录，无法删除")
	}

	// 删除关联的奖品
	if err := DB.Where("activity_id = ?", id).Delete(&model.LotteryPrize{}).Error; err != nil {
		return err
	}

	return DB.Delete(&model.LotteryActivity{}, id).Error
}

// ==================== 奖品管理 ====================

// CreatePrize 创建奖品
func (s *LotteryService) CreatePrize(activityID uint, req model.CreateLotteryPrizeRequest) (*model.LotteryPrize, error) {
	// 检查活动是否存在
	var activity model.LotteryActivity
	if err := DB.First(&activity, activityID).Error; err != nil {
		return nil, errors.New("活动不存在")
	}

	// 检查概率总和
	var totalProb float64
	DB.Model(&model.LotteryPrize{}).Where("activity_id = ?", activityID).
		Select("COALESCE(SUM(probability), 0)").Scan(&totalProb)

	if totalProb+req.Probability > 1.0 {
		return nil, errors.New("奖品概率总和不能超过1")
	}

	prize := model.LotteryPrize{
		ActivityID:         activityID,
		Name:               req.Name,
		Description:        req.Description,
		ImageURL:           req.ImageURL,
		PrizeType:          req.PrizeType,
		PrizeValue:         req.PrizeValue,
		TotalCount:         req.TotalCount,
		RemainingCount:     req.TotalCount,
		Probability:        req.Probability,
		DisplayProbability: req.DisplayProbability,
		SortOrder:          req.SortOrder,
	}

	if err := DB.Create(&prize).Error; err != nil {
		return nil, err
	}
	return &prize, nil
}

// ListPrizes 获取奖品列表
func (s *LotteryService) ListPrizes(activityID uint) ([]model.LotteryPrize, error) {
	var prizes []model.LotteryPrize
	if err := DB.Where("activity_id = ?", activityID).
		Order("sort_order ASC, id ASC").Find(&prizes).Error; err != nil {
		return nil, err
	}
	return prizes, nil
}

// UpdatePrize 更新奖品
func (s *LotteryService) UpdatePrize(id uint, req model.UpdateLotteryPrizeRequest) error {
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.PrizeType != nil {
		updates["prize_type"] = *req.PrizeType
	}
	if req.PrizeValue != nil {
		updates["prize_value"] = *req.PrizeValue
	}
	if req.TotalCount != nil {
		updates["total_count"] = *req.TotalCount
		updates["remaining_count"] = *req.TotalCount // 重置剩余数量
	}
	if req.Probability != nil {
		updates["probability"] = *req.Probability
	}
	if req.DisplayProbability != nil {
		updates["display_probability"] = *req.DisplayProbability
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	return DB.Model(&model.LotteryPrize{}).Where("id = ?", id).Updates(updates).Error
}

// DeletePrize 删除奖品
func (s *LotteryService) DeletePrize(id uint) error {
	return DB.Delete(&model.LotteryPrize{}, id).Error
}

// ==================== 抽奖逻辑 ====================

// Draw 执行抽奖
func (s *LotteryService) Draw(userID, activityID uint, userName string) (*model.LotteryPrize, error) {
	// 使用事务确保数据一致性
	var result *model.LotteryPrize
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. 获取活动信息
		var activity model.LotteryActivity
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&activity, activityID).Error; err != nil {
			return errors.New("活动不存在")
		}

		// 2. 检查活动状态
		now := time.Now()
		if now.Before(activity.StartTime) {
			return errors.New("活动尚未开始")
		}
		if now.After(activity.EndTime) {
			return errors.New("活动已结束")
		}
		if activity.Status != 1 {
			return errors.New("活动未进行中")
		}

		// 3. 检查参与人数限制
		if activity.MaxParticipants > 0 && activity.CurrentParticipants >= activity.MaxParticipants {
			return errors.New("参与人数已满")
		}

		// 4. 检查每日抽奖次数限制
		if activity.DailyLimit > 0 {
			var dailyCount int64
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			query := tx.Model(&model.LotteryRecord{}).
				Where("activity_id = ? AND created_at >= ?", activityID, today)
			if userID > 0 {
				query = query.Where("user_id = ?", userID)
			} else {
				query = query.Where("user_name = ?", userName)
			}
			query.Count(&dailyCount)
			if int(dailyCount) >= activity.DailyLimit {
				return errors.New("今日抽奖次数已用完")
			}
		}

		// 5. 检查总抽奖次数限制
		if activity.TotalLimit > 0 {
			var totalCount int64
			query := tx.Model(&model.LotteryRecord{}).
				Where("activity_id = ?", activityID)
			if userID > 0 {
				query = query.Where("user_id = ?", userID)
			} else {
				query = query.Where("user_name = ?", userName)
			}
			query.Count(&totalCount)
			if int(totalCount) >= activity.TotalLimit {
				return errors.New("抽奖次数已用完")
			}
		}

		// 6. 获取所有奖品
		var prizes []model.LotteryPrize
		if err := tx.Where("activity_id = ? AND remaining_count > 0", activityID).
			Order("sort_order ASC").Find(&prizes).Error; err != nil {
			return errors.New("获取奖品失败")
		}

		if len(prizes) == 0 {
			return errors.New("奖品已抽完")
		}

		// 7. 根据概率抽奖
		var wonPrize *model.LotteryPrize
		random := rand.Float64()
		cumulativeProb := 0.0

		for i := range prizes {
			cumulativeProb += prizes[i].Probability
			if random < cumulativeProb {
				wonPrize = &prizes[i]
				break
			}
		}

		// 8. 创建抽奖记录
		record := model.LotteryRecord{
			ActivityID: activityID,
			UserID:     userID,
			UserName:   userName,
		}

		if wonPrize != nil {
			// 中奖
			record.PrizeID = &wonPrize.ID
			record.PrizeName = wonPrize.Name
			record.IsWinner = true

			// 减少奖品剩余数量
			if err := tx.Model(&model.LotteryPrize{}).Where("id = ?", wonPrize.ID).
				Update("remaining_count", gorm.Expr("remaining_count - 1")).Error; err != nil {
				return errors.New("更新奖品数量失败")
			}

			result = wonPrize
		}

		// 9. 保存抽奖记录
		if err := tx.Create(&record).Error; err != nil {
			return errors.New("保存抽奖记录失败")
		}

		// 10. 更新参与人数
		if err := tx.Model(&model.LotteryActivity{}).Where("id = ?", activityID).
			Update("current_participants", gorm.Expr("current_participants + 1")).Error; err != nil {
			return errors.New("更新参与人数失败")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ==================== 记录查询 ====================

// DeleteRecord 删除抽奖记录
func (s *LotteryService) DeleteRecord(id uint) error {
	return DB.Delete(&model.LotteryRecord{}, id).Error
}

// DeleteRecordsByActivityID 删除活动的所有抽奖记录
func (s *LotteryService) DeleteRecordsByActivityID(activityID uint) error {
	return DB.Where("activity_id = ?", activityID).Delete(&model.LotteryRecord{}).Error
}

// ListRecords 获取抽奖记录列表
func (s *LotteryService) ListRecords(activityID *uint, userID *uint, page, pageSize int) ([]model.LotteryRecord, int64, error) {
	var records []model.LotteryRecord
	var total int64

	query := DB.Model(&model.LotteryRecord{})

	if activityID != nil {
		query = query.Where("activity_id = ?", *activityID)
	}
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// ListWinners 获取中奖名单
func (s *LotteryService) ListWinners(activityID *uint, page, pageSize int) ([]model.LotteryRecord, int64, error) {
	var records []model.LotteryRecord
	var total int64

	query := DB.Model(&model.LotteryRecord{}).Where("is_winner = ?", true)

	if activityID != nil {
		query = query.Where("activity_id = ?", *activityID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetUserDrawCount 获取用户抽奖次数
func (s *LotteryService) GetUserDrawCount(activityID, userID uint) (int, error) {
	var count int64
	if err := DB.Model(&model.LotteryRecord{}).
		Where("activity_id = ? AND user_id = ?", activityID, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetUserDailyDrawCount 获取用户今日抽奖次数
func (s *LotteryService) GetUserDailyDrawCount(activityID, userID uint) (int, error) {
	var count int64
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if err := DB.Model(&model.LotteryRecord{}).
		Where("activity_id = ? AND user_id = ? AND created_at >= ?", activityID, userID, today).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetUserDailyDrawCountByName 根据用户名获取今日抽奖次数
func (s *LotteryService) GetUserDailyDrawCountByName(activityID uint, userName string) (int, error) {
	var count int64
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if err := DB.Model(&model.LotteryRecord{}).
		Where("activity_id = ? AND user_name = ? AND created_at >= ?", activityID, userName, today).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetTotalDailyDrawCount 获取今日所有用户总抽奖次数
func (s *LotteryService) GetTotalDailyDrawCount(activityID uint) (int, error) {
	var count int64
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if err := DB.Model(&model.LotteryRecord{}).
		Where("activity_id = ? AND created_at >= ?", activityID, today).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetUserTotalDrawCountByName 根据用户名获取总抽奖次数
func (s *LotteryService) GetUserTotalDrawCountByName(activityID uint, userName string) (int, error) {
	var count int64

	if err := DB.Model(&model.LotteryRecord{}).
		Where("activity_id = ? AND user_name = ?", activityID, userName).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
