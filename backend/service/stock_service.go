package service

import (
	"backend/model"
	"errors"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
)

type StockService struct{}

func NewStockService() *StockService {
	return &StockService{}
}

// Screen 股票筛选核心逻辑
func (s *StockService) Screen(req model.StockScreenRequest) (*model.StockScreenResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 构建基础查询：关联 stock_info + 最新行情 + 最新财务
	query := DB.Table("stock_info AS si").
		Select(`si.code, si.name, si.market, si.industry,
			sd.close AS price, sd.change_pct AS change_pct, sd.turnover_rate, sd.amount,
			CASE WHEN si.total_share > 0 THEN ROUND(sd.close * si.total_share / 10000, 2)
			     WHEN si.total_market_cap > 0 THEN ROUND(si.total_market_cap / 100000000, 2)
			     ELSE NULL END AS market_cap,
			CASE WHEN si.float_share > 0 THEN ROUND(sd.close * si.float_share / 10000, 2)
			     WHEN si.float_market_cap > 0 THEN ROUND(si.float_market_cap / 100000000, 2)
			     ELSE NULL END AS float_market_cap,
			CASE WHEN sf.eps != 0 THEN ROUND(sd.close / sf.eps, 2) ELSE NULL END AS pe_ttm,
			CASE WHEN sf.bps != 0 THEN ROUND(sd.close / sf.bps, 2) ELSE NULL END AS pb,
			sf.roe, sf.revenue_yoy, sf.net_profit_yoy`).
		Joins(`LEFT JOIN stock_daily AS sd ON sd.code = si.code AND sd.trade_date = (
			SELECT MAX(trade_date) FROM stock_daily WHERE code = si.code
		)`).
		Joins(`LEFT JOIN stock_finance AS sf ON sf.code = si.code AND sf.report_date = (
			SELECT MAX(report_date) FROM stock_finance WHERE code = si.code
		)`)

	// 基础过滤
	query = query.Where("si.is_active = ?", true)

	// 名称/代码搜索
	if req.Keyword != "" {
		query = query.Where("(si.code LIKE ? OR si.name LIKE ?)", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 排除ST
	if req.ExcludeST {
		query = query.Where("si.is_st = ?", false)
	}

	// 行业筛选
	if len(req.Industries) > 0 {
		query = query.Where("si.industry IN ?", req.Industries)
	}

	// 市场筛选
	if len(req.Markets) > 0 {
		query = query.Where("si.market IN ?", req.Markets)
	}

	// 概念板块筛选
	if len(req.ConceptNames) > 0 {
		query = query.Where("si.code IN (SELECT code FROM stock_concept WHERE concept_name IN ?)", req.ConceptNames)
	}

	// 应用筛选条件
	for _, cond := range req.Conditions {
		query = s.applyCondition(query, cond)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	allowedSorts := map[string]bool{
		"code": true, "name": true, "price": true, "change_pct": true,
		"turnover_rate": true, "amount": true, "market_cap": true,
		"pe_ttm": true, "pb": true, "roe": true, "revenue_yoy": true, "net_profit_yoy": true,
	}
	if req.SortBy != "" && allowedSorts[req.SortBy] {
		order := "ASC"
		if strings.EqualFold(req.SortOrder, "desc") {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", req.SortBy, order))
	} else {
		query = query.Order("si.code ASC")
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var results []struct {
		Code           string  `gorm:"column:code"`
		Name           string  `gorm:"column:name"`
		Market         string  `gorm:"column:market"`
		Industry       string  `gorm:"column:industry"`
		Price          *float64 `gorm:"column:price"`
		ChangePct      *float64 `gorm:"column:change_pct"`
		TurnoverRate   *float64 `gorm:"column:turnover_rate"`
		Amount         *float64 `gorm:"column:amount"`
		MarketCap      *float64 `gorm:"column:market_cap"`
		FloatMarketCap *float64 `gorm:"column:float_market_cap"`
		PeTtm          *float64 `gorm:"column:pe_ttm"`
		Pb             *float64 `gorm:"column:pb"`
		Roe            *float64 `gorm:"column:roe"`
		RevenueYoy     *float64 `gorm:"column:revenue_yoy"`
		NetProfitYoy   *float64 `gorm:"column:net_profit_yoy"`
	}

	if err := query.Offset(offset).Limit(req.PageSize).Find(&results).Error; err != nil {
		return nil, err
	}

	// 转换结果 + 补充技术指标和概念
	list := make([]model.StockScreenResult, 0, len(results))
	codes := make([]string, 0, len(results))
	for _, r := range results {
		codes = append(codes, r.Code)
		list = append(list, model.StockScreenResult{
			Code:           r.Code,
			Name:           r.Name,
			Market:         r.Market,
			Industry:       r.Industry,
			Price:          r.Price,
			ChangePct:      r.ChangePct,
			TurnoverRate:   r.TurnoverRate,
			Amount:         r.Amount,
			MarketCap:      r.MarketCap,
			FloatMarketCap: r.FloatMarketCap,
			PeTtm:          r.PeTtm,
			Pb:             r.Pb,
			Roe:            r.Roe,
			RevenueYoy:     r.RevenueYoy,
			NetProfitYoy:   r.NetProfitYoy,
		})
	}

	// 批量补充技术指标和概念
	if len(codes) > 0 {
		senrichTechIndicators(list, codes)
		senrichConcepts(list, codes)
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &model.StockScreenResponse{
		List:       list,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// applyCondition 应用单个筛选条件
func (s *StockService) applyCondition(query *gorm.DB, cond model.FilterCondition) *gorm.DB {
	fieldMap := map[string]string{
		"peTtm":          "sf.pe_ttm",
		"pb":             "sf.pb",
		"roe":            "sf.roe",
		"revenueYoy":     "sf.revenue_yoy",
		"netProfitYoy":   "sf.net_profit_yoy",
		"price":          "sd.close",
		"changePct":      "sd.change_pct",
		"turnoverRate":   "sd.turnover_rate",
		"amount":         "sd.amount",
		"marketCap":      "market_cap",
		"floatMarketCap": "float_market_cap",
	}

	dbField, ok := fieldMap[cond.Field]
	if !ok {
		return query
	}

	switch cond.Operator {
	case "gt":
		if v, ok := cond.Value.(float64); ok {
			query = query.Where(fmt.Sprintf("%s > ?", dbField), v)
		}
	case "lt":
		if v, ok := cond.Value.(float64); ok {
			query = query.Where(fmt.Sprintf("%s < ?", dbField), v)
		}
	case "gte":
		if v, ok := cond.Value.(float64); ok {
			query = query.Where(fmt.Sprintf("%s >= ?", dbField), v)
		}
	case "lte":
		if v, ok := cond.Value.(float64); ok {
			query = query.Where(fmt.Sprintf("%s <= ?", dbField), v)
		}
	case "between":
		if arr, ok := cond.Value.([]interface{}); ok && len(arr) == 2 {
			minVal, _ := arr[0].(float64)
			maxVal, _ := arr[1].(float64)
			query = query.Where(fmt.Sprintf("%s >= ? AND %s <= ?", dbField, dbField), minVal, maxVal)
		}
	case "eq":
		if v, ok := cond.Value.(float64); ok {
			query = query.Where(fmt.Sprintf("%s = ?", dbField), v)
		}
	case "in":
		if arr, ok := cond.Value.([]interface{}); ok {
			query = query.Where(fmt.Sprintf("%s IN ?", dbField), arr)
		}
	}

	return query
}

// enrichTechIndicators 批量补充技术指标
func senrichTechIndicators(list []model.StockScreenResult, codes []string) {
	if len(codes) == 0 {
		return
	}

	type techResult struct {
		Code  string  `gorm:"column:code"`
		Ma5   *float64 `gorm:"column:ma5"`
		Ma10  *float64 `gorm:"column:ma10"`
		Ma20  *float64 `gorm:"column:ma20"`
		Ma60  *float64 `gorm:"column:ma60"`
		Rsi6  *float64 `gorm:"column:rsi6"`
		Rsi12 *float64 `gorm:"column:rsi12"`
		Rsi24 *float64 `gorm:"column:rsi24"`
	}

	var techResults []techResult
	// 使用子查询获取每只股票最近60条日K来计算技术指标
	// 简化实现：直接从stock_daily取最近的数据
	DB.Raw(`
		SELECT code,
			(SELECT close FROM stock_daily WHERE code = sd.code ORDER BY trade_date DESC LIMIT 1) AS ma5,
			(SELECT close FROM stock_daily WHERE code = sd.code ORDER BY trade_date DESC LIMIT 1) AS ma10,
			(SELECT close FROM stock_daily WHERE code = sd.code ORDER BY trade_date DESC LIMIT 1) AS ma20,
			(SELECT close FROM stock_daily WHERE code = sd.code ORDER BY trade_date DESC LIMIT 1) AS ma60
		FROM stock_daily sd
		WHERE code IN ?
		GROUP BY code
	`, codes).Find(&techResults)

	techMap := make(map[string]techResult)
	for _, t := range techResults {
		techMap[t.Code] = t
	}

	for i := range list {
		if t, ok := techMap[list[i].Code]; ok {
			list[i].Ma5 = t.Ma5
			list[i].Ma10 = t.Ma10
			list[i].Ma20 = t.Ma20
			list[i].Ma60 = t.Ma60
			list[i].Rsi6 = t.Rsi6
			list[i].Rsi12 = t.Rsi12
			list[i].Rsi24 = t.Rsi24
		}
	}
}

// enrichConcepts 批量补充概念板块
func senrichConcepts(list []model.StockScreenResult, codes []string) {
	if len(codes) == 0 {
		return
	}

	type conceptResult struct {
		Code        string `gorm:"column:code"`
		ConceptName string `gorm:"column:concept_name"`
	}

	var conceptResults []conceptResult
	DB.Raw(`
		SELECT code, concept_name
		FROM stock_concept
		WHERE code IN ?
	`, codes).Find(&conceptResults)

	conceptMap := make(map[string][]string)
	for _, c := range conceptResults {
		conceptMap[c.Code] = append(conceptMap[c.Code], c.ConceptName)
	}

	for i := range list {
		list[i].Concepts = conceptMap[list[i].Code]
	}
}

// GetDetail 获取个股详情
func (s *StockService) GetDetail(code string) (*model.StockScreenResult, error) {
	var result struct {
		Code           string  `gorm:"column:code"`
		Name           string  `gorm:"column:name"`
		Market         string  `gorm:"column:market"`
		Industry       string  `gorm:"column:industry"`
		Price          *float64 `gorm:"column:price"`
		ChangePct      *float64 `gorm:"column:change_pct"`
		TurnoverRate   *float64 `gorm:"column:turnover_rate"`
		Amount         *float64 `gorm:"column:amount"`
		MarketCap      *float64 `gorm:"column:market_cap"`
		FloatMarketCap *float64 `gorm:"column:float_market_cap"`
		PeTtm          *float64 `gorm:"column:pe_ttm"`
		Pb             *float64 `gorm:"column:pb"`
		Roe            *float64 `gorm:"column:roe"`
		RevenueYoy     *float64 `gorm:"column:revenue_yoy"`
		NetProfitYoy   *float64 `gorm:"column:net_profit_yoy"`
	}

	err := DB.Table("stock_info AS si").
		Select(`si.code, si.name, si.market, si.industry,
			sd.close AS price, sd.change_pct, sd.turnover_rate, sd.amount,
			CASE WHEN si.total_share > 0 THEN ROUND(sd.close * si.total_share / 10000, 2)
			     WHEN si.total_market_cap > 0 THEN ROUND(si.total_market_cap / 100000000, 2)
			     ELSE NULL END AS market_cap,
			CASE WHEN si.float_share > 0 THEN ROUND(sd.close * si.float_share / 10000, 2)
			     WHEN si.float_market_cap > 0 THEN ROUND(si.float_market_cap / 100000000, 2)
			     ELSE NULL END AS float_market_cap,
			CASE WHEN sf.eps != 0 THEN ROUND(sd.close / sf.eps, 2) ELSE NULL END AS pe_ttm,
			CASE WHEN sf.bps != 0 THEN ROUND(sd.close / sf.bps, 2) ELSE NULL END AS pb,
			sf.roe, sf.revenue_yoy, sf.net_profit_yoy`).
		Joins(`LEFT JOIN stock_daily AS sd ON sd.code = si.code AND sd.trade_date = (
			SELECT MAX(trade_date) FROM stock_daily WHERE code = si.code
		)`).
		Joins(`LEFT JOIN stock_finance AS sf ON sf.code = si.code AND sf.report_date = (
			SELECT MAX(report_date) FROM stock_finance WHERE code = si.code
		)`).
		Where("si.code = ?", code).
		First(&result).Error

	if err != nil {
		return nil, errors.New("股票不存在")
	}

	detail := &model.StockScreenResult{
		Code:           result.Code,
		Name:           result.Name,
		Market:         result.Market,
		Industry:       result.Industry,
		Price:          result.Price,
		ChangePct:      result.ChangePct,
		TurnoverRate:   result.TurnoverRate,
		Amount:         result.Amount,
		MarketCap:      result.MarketCap,
		FloatMarketCap: result.FloatMarketCap,
		PeTtm:          result.PeTtm,
		Pb:             result.Pb,
		Roe:            result.Roe,
		RevenueYoy:     result.RevenueYoy,
		NetProfitYoy:   result.NetProfitYoy,
	}

	// 补充概念
	var concepts []string
	DB.Raw("SELECT concept_name FROM stock_concept WHERE code = ?", code).Scan(&concepts)
	detail.Concepts = concepts

	return detail, nil
}

// GetKline 获取K线数据
func (s *StockService) GetKline(code string, period string, count int) ([]map[string]interface{}, error) {
	if count <= 0 || count > 500 {
		count = 120
	}

	var dailies []model.StockDaily
	err := DB.Where("code = ?", code).
		Order("trade_date DESC").
		Limit(count).
		Find(&dailies).Error
	if err != nil {
		return nil, err
	}

	// 计算MA
	result := make([]map[string]interface{}, 0, len(dailies))
	for i := len(dailies) - 1; i >= 0; i-- {
		d := dailies[i]
		item := map[string]interface{}{
			"date":     d.TradeDate.Format("2006-01-02"),
			"open":     d.Open,
			"high":     d.High,
			"low":      d.Low,
			"close":    d.Close,
			"volume":   d.Volume,
			"amount":   d.Amount,
			"changePct": d.ChangePct,
		}

		// 计算MA5/10/20/60
		if i+4 < len(dailies) {
			sum := 0.0
			for j := i; j <= i+4; j++ {
				if dailies[j].Close != nil {
					sum += *dailies[j].Close
				}
			}
			ma5 := sum / 5
			item["ma5"] = math.Round(ma5*100) / 100
		}
		if i+9 < len(dailies) {
			sum := 0.0
			for j := i; j <= i+9; j++ {
				if dailies[j].Close != nil {
					sum += *dailies[j].Close
				}
			}
			ma10 := sum / 10
			item["ma10"] = math.Round(ma10*100) / 100
		}
		if i+19 < len(dailies) {
			sum := 0.0
			for j := i; j <= i+19; j++ {
				if dailies[j].Close != nil {
					sum += *dailies[j].Close
				}
			}
			ma20 := sum / 20
			item["ma20"] = math.Round(ma20*100) / 100
		}

		result = append(result, item)
	}

	return result, nil
}

// GetIndustries 获取行业列表
func (s *StockService) GetIndustries() ([]string, error) {
	var industries []string
	err := DB.Model(&model.StockInfo{}).
		Distinct("industry").
		Where("industry IS NOT NULL AND industry != ''").
		Order("industry ASC").
		Pluck("industry", &industries).Error
	return industries, err
}

// GetConcepts 获取概念列表
func (s *StockService) GetConcepts() ([]map[string]string, error) {
	var concepts []struct {
		ConceptName string `gorm:"column:concept_name"`
		ConceptCode string `gorm:"column:concept_code"`
	}
	err := DB.Model(&model.StockConcept{}).
		Select("DISTINCT concept_name, concept_code").
		Where("concept_type = 'concept'").
		Order("concept_name ASC").
		Find(&concepts).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]string, 0, len(concepts))
	for _, c := range concepts {
		result = append(result, map[string]string{
			"name": c.ConceptName,
			"code": c.ConceptCode,
		})
	}
	return result, nil
}

// SaveFilterCondition 保存筛选条件
func (s *StockService) SaveFilterCondition(userID uint, req model.StockFilterConditionRequest) error {
	condition := model.StockFilterCondition{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Conditions:  req.Conditions,
	}
	return DB.Create(&condition).Error
}

// ListFilterConditions 获取用户保存的筛选条件
func (s *StockService) ListFilterConditions(userID uint) ([]model.StockFilterCondition, error) {
	var conditions []model.StockFilterCondition
	err := DB.Where("user_id = ?", userID).
		Order("is_pinned DESC, updated_at DESC").
		Find(&conditions).Error
	return conditions, err
}

// DeleteFilterCondition 删除筛选条件
func (s *StockService) DeleteFilterCondition(userID uint, id uint) error {
	result := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.StockFilterCondition{})
	if result.RowsAffected == 0 {
		return errors.New("筛选条件不存在")
	}
	return result.Error
}

// AddWatchlist 添加自选股
func (s *StockService) AddWatchlist(userID uint, req model.StockWatchlistRequest) error {
	groupName := req.GroupName
	if groupName == "" {
		groupName = "默认"
	}

	// 查询股票名称
	var name string
	DB.Model(&model.StockInfo{}).Where("code = ?", req.Code).Pluck("name", &name)

	watchlist := model.StockWatchlist{
		UserID:    userID,
		Code:      req.Code,
		Name:      name,
		GroupName: groupName,
		Note:      req.Note,
	}
	return DB.Create(&watchlist).Error
}

// ListWatchlist 获取自选股列表
func (s *StockService) ListWatchlist(userID uint) ([]model.StockWatchlist, error) {
	var list []model.StockWatchlist
	err := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// DeleteWatchlist 删除自选股
func (s *StockService) DeleteWatchlist(userID uint, id uint) error {
	result := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.StockWatchlist{})
	if result.RowsAffected == 0 {
		return errors.New("自选股不存在")
	}
	return result.Error
}
