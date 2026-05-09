package model

type DashboardStats struct {
	TodayTrainings    int64            `json:"today_trainings"`
	TotalTrainings    int64            `json:"total_trainings"`
	TotalVocabulary   int64            `json:"total_vocabulary"`
	TotalNotes        int64            `json:"total_notes"`
	TotalFavorites    int64            `json:"total_favorites"`
	TrainingTrend     []TrendItem      `json:"training_trend"`
	TrainingTypeStats []TypeStatItem   `json:"training_type_stats"`
}

type TrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TypeStatItem struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}
