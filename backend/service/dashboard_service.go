package service

import (
	"backend/model"
	"time"
)

type DashboardService struct{}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

func (s *DashboardService) GetStats(userID uint) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	// 今日训练次数
	today := time.Now().Format("2006-01-02")
	DB.Model(&model.TrainingHistory{}).
		Where("user_id = ? AND DATE(created_at) = ?", userID, today).
		Count(&stats.TodayTrainings)

	// 累计训练次数
	DB.Model(&model.TrainingHistory{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalTrainings)

	// 生词本词汇量
	DB.Model(&model.Vocabulary{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalVocabulary)

	// 笔记数量
	DB.Model(&model.Note{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalNotes)

	// 收藏数量
	DB.Model(&model.TrainingHistory{}).
		Where("user_id = ? AND is_favorite = ?", userID, true).
		Count(&stats.TotalFavorites)

	// 近7天训练趋势
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		DB.Model(&model.TrainingHistory{}).
			Where("user_id = ? AND DATE(created_at) = ?", userID, date).
			Count(&count)
		stats.TrainingTrend = append(stats.TrainingTrend, model.TrendItem{
			Date:  date,
			Count: count,
		})
	}

	// 训练类型分布
	type TypeResult struct {
		TrainingType string `gorm:"column:training_type"`
		Count        int64  `gorm:"column:count"`
	}
	var typeResults []TypeResult
	DB.Model(&model.TrainingHistory{}).
		Select("training_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("training_type").
		Scan(&typeResults)

	typeNameMap := map[string]string{
		"ai_chat":      "英语训练",
		"ai_decision":  "决策训练",
		"ai_social":    "社交训练",
		"ai_emergency": "应急训练",
	}

	for _, r := range typeResults {
		name := typeNameMap[r.TrainingType]
		if name == "" {
			name = r.TrainingType
		}
		stats.TrainingTypeStats = append(stats.TrainingTypeStats, model.TypeStatItem{
			Type:  name,
			Count: r.Count,
		})
	}

	return stats, nil
}
