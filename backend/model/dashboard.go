package model

type DashboardStats struct {
	TodayMessages     int64            `json:"today_messages"`
	TotalMessages     int64            `json:"total_messages"`
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
