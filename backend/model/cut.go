package model

import "time"

// CutRecord 切割记录实体
type CutRecord struct {
	ID         string    `gorm:"primaryKey;size:100" json:"id"`
	Type       string    `gorm:"size:20;index" json:"type"`
	Request    string    `gorm:"type:text" json:"request"`
	Response   string    `gorm:"type:text" json:"response"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	UserID     uint      `gorm:"not null;index" json:"userId"`
	Code       string    `gorm:"size:100" json:"code"`
	Name       string    `gorm:"size:100;not null" json:"name"`
}

func (CutRecord) TableName() string {
	return "cut_record"
}

// BarRequest 一维切割请求
type BarRequest struct {
	Items              []int     `json:"items" binding:"required,min=1"`
	Materials          []int     `json:"materials"`
	NewMaterialLength  int       `json:"newMaterialLength" binding:"required,min=1"`
	Loss               float64   `json:"loss"`
	UtilizationWeight  int       `json:"utilizationWeight"`
}

// BarResult 一维切割结果
type BarResult struct {
	Index       int     `json:"index"`
	TotalLength int     `json:"totalLength"`
	Cuts        []int   `json:"cuts"`
	Used        float64 `json:"used"`
	Remaining   float64 `json:"remaining"`
}

// Item 切割项目
type Item struct {
	Label    string  `json:"label"`
	Width    float64 `json:"width" binding:"required,min=1"`
	Height   float64 `json:"height" binding:"required,min=1"`
	Quantity int     `json:"quantity"`
}

// Piece 切割块
type Piece struct {
	Label   string  `json:"label"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	W       float64 `json:"w"`
	H       float64 `json:"h"`
	Rotated bool    `json:"rotated"`
}

// BinRequest 平面切割请求
type BinRequest struct {
	Items     []Item `json:"items" binding:"required,min=1"`
	Materials []Item `json:"materials"`
	Height    float64    `json:"height" binding:"required,min=1"`
	Width     float64    `json:"width" binding:"required,min=1"`
	Strategy  string `json:"strategy" binding:"required"`
}

// BinResult 平面切割结果
type BinResult struct {
	BinID          int     `json:"binId"`
	MaterialType   string  `json:"materialType"`
	MaterialWidth  float64 `json:"materialWidth"`
	MaterialHeight float64 `json:"materialHeight"`
	Pieces         []Piece `json:"pieces"`
	Utilization    float64 `json:"utilization"`
}

// RecordRequest 保存切割记录请求
type RecordRequest struct {
	Type     string `json:"type" binding:"required"`
	Request  string `json:"request" binding:"required"`
	Response string `json:"response" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// CutRecordSearchParams 切割记录搜索参数
type CutRecordSearchParams struct {
	Page      int    `form:"current" binding:"required,min=1"`
	PageSize  int    `form:"size" binding:"required,min=1,max=100"`
	Name      string `form:"name"`
	Type      string `form:"type"`
	StartTime *int64 `form:"startTime"`
	EndTime   *int64 `form:"endTime"`
}

// CutRecordListResponse 切割记录列表响应
type CutRecordListResponse struct {
	Total int64       `json:"total"`
	Items []CutRecord `json:"items"`
}
