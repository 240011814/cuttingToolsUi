package model

import (
	"time"
)

// StockInfo 股票基础信息
type StockInfo struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Code            string     `gorm:"size:10;not null;uniqueIndex" json:"code"`
	Name            string     `gorm:"size:50;not null" json:"name"`
	Market          string     `gorm:"size:10;default:SZ" json:"market"`
	Industry        string     `gorm:"size:50" json:"industry"`
	Area            string     `gorm:"size:50" json:"area"`
	ListDate        *time.Time `json:"listDate"`
	IsST            bool       `gorm:"default:false" json:"isSt"`
	IsActive        bool       `gorm:"default:true" json:"isActive"`
	TotalShare      *float64   `json:"totalShare"`
	FloatShare      *float64   `json:"floatShare"`
	TotalMarketCap  *float64   `json:"totalMarketCap"`
	FloatMarketCap  *float64   `json:"floatMarketCap"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (StockInfo) TableName() string {
	return "stock_info"
}

// StockDaily 股票日行情
type StockDaily struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Code         string     `gorm:"size:10;not null" json:"code"`
	TradeDate    time.Time  `gorm:"not null" json:"tradeDate"`
	Open         *float64   `json:"open"`
	High         *float64   `json:"high"`
	Low          *float64   `json:"low"`
	Close        *float64   `json:"close"`
	Volume       *float64   `json:"volume"`
	Amount       *float64   `json:"amount"`
	TurnoverRate *float64   `json:"turnoverRate"`
	ChangePct    *float64   `json:"changePct"`
	Amplitude    *float64   `json:"amplitude"`
	CreatedAt    time.Time  `json:"createdAt"`
}

func (StockDaily) TableName() string {
	return "stock_daily"
}

// StockFinance 股票财务数据
type StockFinance struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Code           string     `gorm:"size:10;not null" json:"code"`
	ReportDate     time.Time  `gorm:"not null" json:"reportDate"`
	ReportType     string     `gorm:"size:20" json:"reportType"`
	PeTtm          *float64   `json:"peTtm"`
	Pb             *float64   `json:"pb"`
	PsTtm          *float64   `json:"psTtm"`
	Roe            *float64   `json:"roe"`
	Roa            *float64   `json:"roa"`
	GrossMargin    *float64   `json:"grossMargin"`
	NetMargin      *float64   `json:"netMargin"`
	Revenue        *float64   `json:"revenue"`
	RevenueYoy     *float64   `json:"revenueYoy"`
	NetProfit      *float64   `json:"netProfit"`
	NetProfitYoy   *float64   `json:"netProfitYoy"`
	DebtRatio      *float64   `json:"debtRatio"`
	CurrentRatio   *float64   `json:"currentRatio"`
	QuickRatio     *float64   `json:"quickRatio"`
	Eps            *float64   `json:"eps"`
	EpsDeducted    *float64   `json:"epsDeducted"`
	Bps            *float64   `json:"bps"`
	OcfPerShare    *float64   `json:"ocfPerShare"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (StockFinance) TableName() string {
	return "stock_finance"
}

// StockConcept 股票概念关联
type StockConcept struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:10;not null" json:"code"`
	ConceptName string    `gorm:"size:50;not null" json:"conceptName"`
	ConceptCode string    `gorm:"size:20" json:"conceptCode"`
	ConceptType string    `gorm:"size:20;default:industry" json:"conceptType"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (StockConcept) TableName() string {
	return "stock_concept"
}

// StockFilterCondition 用户保存的筛选条件
type StockFilterCondition struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Conditions  string    `gorm:"type:json;not null" json:"conditions"`
	ResultCount int       `json:"resultCount"`
	IsPinned    bool      `gorm:"default:false" json:"isPinned"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (StockFilterCondition) TableName() string {
	return "stock_filter_condition"
}

// StockWatchlist 自选股
type StockWatchlist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	Name      string    `gorm:"size:50" json:"name"`
	GroupName string    `gorm:"size:50;default:默认" json:"groupName"`
	Note      string    `gorm:"size:200" json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

func (StockWatchlist) TableName() string {
	return "stock_watchlist"
}

// === 筛选请求/响应结构 ===

// FilterCondition 单个筛选条件
type FilterCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// StockScreenRequest 筛选请求
type StockScreenRequest struct {
	Conditions   []FilterCondition `json:"conditions"`
	SortBy       string            `json:"sortBy"`
	SortOrder    string            `json:"sortOrder"`
	Page         int               `json:"page"`
	PageSize     int               `json:"pageSize"`
	ConceptNames []string          `json:"conceptNames"`
	Industries   []string          `json:"industries"`
	Markets      []string          `json:"markets"`
	ExcludeST    bool              `json:"excludeSt"`
	Keyword      string            `json:"keyword"`
}

// StockScreenResult 筛选结果项
type StockScreenResult struct {
	Code           string   `json:"code"`
	Name           string   `json:"name"`
	Market         string   `json:"market"`
	Industry       string   `json:"industry"`
	Price          *float64 `json:"price"`
	ChangePct      *float64 `json:"changePct"`
	TurnoverRate   *float64 `json:"turnoverRate"`
	Amount         *float64 `json:"amount"`
	MarketCap      *float64 `json:"marketCap"`
	FloatMarketCap *float64 `json:"floatMarketCap"`
	PeTtm          *float64 `json:"peTtm"`
	Pb             *float64 `json:"pb"`
	Roe            *float64 `json:"roe"`
	RevenueYoy     *float64 `json:"revenueYoy"`
	NetProfitYoy   *float64 `json:"netProfitYoy"`
	GrossMargin    *float64 `json:"grossMargin"`
	NetMargin      *float64 `json:"netMargin"`
	DebtRatio      *float64 `json:"debtRatio"`
	CurrentRatio   *float64 `json:"currentRatio"`
	Change5d       *float64 `json:"change5d"`
	Change20d      *float64 `json:"change20d"`
	Ma5            *float64 `json:"ma5"`
	Ma10           *float64 `json:"ma10"`
	Ma20           *float64 `json:"ma20"`
	Ma60           *float64 `json:"ma60"`
	Rsi6           *float64 `json:"rsi6"`
	Rsi12          *float64 `json:"rsi12"`
	Rsi24          *float64 `json:"rsi24"`
	Concepts       []string `json:"concepts"`
}

// StockScreenResponse 筛选响应
type StockScreenResponse struct {
	List       []StockScreenResult `json:"list"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	TotalPages int                 `json:"totalPages"`
}

// StockFilterConditionRequest 保存筛选条件请求
type StockFilterConditionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Conditions  string `json:"conditions" binding:"required"`
}

// StockWatchlistRequest 自选股请求
type StockWatchlistRequest struct {
	Code      string `json:"code" binding:"required"`
	Name      string `json:"name"`
	GroupName string `json:"groupName"`
	Note      string `json:"note"`
}

// StockKlineRequest K线请求
type StockKlineRequest struct {
	Period string `form:"period" json:"period"`
	Count  int    `form:"count" json:"count"`
}
