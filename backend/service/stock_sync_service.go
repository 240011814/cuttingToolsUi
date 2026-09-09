package service

import (
	"backend/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type StockSyncService struct{}

func NewStockSyncService() *StockSyncService {
	return &StockSyncService{}
}

// SyncAll 全量同步
func (s *StockSyncService) SyncAll() error {
	log.Println("[StockSync] 开始全量同步...")

	if err := s.SyncStockList(); err != nil {
		log.Printf("[StockSync] 同步股票列表失败: %v", err)
	}

	if err := s.SyncDailyQuotes(); err != nil {
		log.Printf("[StockSync] 同步行情数据失败: %v", err)
	}

	log.Println("[StockSync] 全量同步完成")
	return nil
}

// SyncStockList 同步股票列表
func (s *StockSyncService) SyncStockList() error {
	log.Println("[StockSync] 同步股票列表...")

	pageSize := 500
	page := 1
	totalCount := 0

	for {
		// 使用东方财富另一个接口
		url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=SECURITY_CODE&sortTypes=1&pageSize=%d&pageNumber=%d&reportName=RPT_LICO_FN_CPD&columns=ALL",
			pageSize, page)

		body, err := s.httpGet(url)
		if err != nil {
			return err
		}

		var resp struct {
			Result struct {
				Count int `json:"count"`
				Data  []struct {
					Code       string  `json:"SECURITY_CODE"`
					Name       string  `json:"SECURITY_NAME_ABBR"`
					MarketCode string  `json:"TRADE_MARKET_CODE"`
					MarketCap  float64 `json:"TOTAL_MARKET_CAP"`
					FreeCap    float64 `json:"FREE_CAP"`
				} `json:"data"`
			} `json:"result"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("解析股票列表失败: %v", err)
		}

		if resp.Message != "" && resp.Message != "ok" {
			return fmt.Errorf("API错误: %s", resp.Message)
		}

		if len(resp.Result.Data) == 0 {
			break
		}

		for _, item := range resp.Result.Data {
			if item.Code == "" {
				continue
			}

			market := "SZ"
			if item.MarketCode == "069001001001" || strings.HasPrefix(item.Code, "6") {
				market = "SH"
			}

			stock := model.StockInfo{
				Code:   item.Code,
				Name:   item.Name,
				Market: market,
			}

			if item.MarketCap > 0 {
				totalShare := item.MarketCap / 10000
				stock.TotalShare = &totalShare
			}
			if item.FreeCap > 0 {
				floatShare := item.FreeCap / 10000
				stock.FloatShare = &floatShare
			}

			stock.IsST = strings.Contains(item.Name, "ST")
			stock.IsActive = true

			var existing model.StockInfo
			result := DB.Where("code = ?", item.Code).First(&existing)
			if result.Error == nil {
				DB.Model(&existing).Updates(map[string]interface{}{
					"name":        stock.Name,
					"market":      stock.Market,
					"is_st":       stock.IsST,
					"total_share": stock.TotalShare,
					"float_share": stock.FloatShare,
				})
			} else {
				DB.Create(&stock)
			}
			totalCount++
		}

		log.Printf("[StockSync] 已同步 %d/%d 只股票", totalCount, resp.Result.Count)

		if len(resp.Result.Data) < pageSize {
			break
		}
		page++
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("[StockSync] 同步股票列表完成: %d 只", totalCount)
	return nil
}

// SyncDailyQuotes 同步行情数据
func (s *StockSyncService) SyncDailyQuotes() error {
	log.Println("[StockSync] 同步行情数据...")

	var stocks []model.StockInfo
	DB.Where("is_active = ?", true).Find(&stocks)

	count := 0
	for _, stock := range stocks {
		if err := s.syncSingleStockDaily(stock.Code); err != nil {
			log.Printf("[StockSync] 同步 %s 失败: %v", stock.Code, err)
			continue
		}
		count++
		if count%100 == 0 {
			log.Printf("[StockSync] 已同步 %d/%d", count, len(stocks))
			time.Sleep(100 * time.Millisecond) // 避免请求过快
		}
	}

	log.Printf("[StockSync] 同步行情数据完成: %d 只", count)
	return nil
}

// syncSingleStockDaily 同步单只股票日K线
func (s *StockSyncService) syncSingleStockDaily(code string) error {
	// 东方财富日K接口
	url := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?cb=jQuery&secid=%s.%s&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&klt=101&fqt=1&end=20500101&lmt=30",
		s.getSecId(code), code)

	body, err := s.httpGet(url)
	if err != nil {
		return err
	}

	// 去掉 JSONP 包装
	jsonStr := s.stripJSONP(string(body))

	var resp struct {
		Data struct {
			Klines []string `json:"klines"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return fmt.Errorf("解析K线数据失败: %v", err)
	}

	for _, line := range resp.Data.Klines {
		// 格式: 日期,开盘,收盘,最高,最低,成交量,成交额,振幅,涨跌幅,涨跌额,换手率
		parts := s.splitKline(line)
		if len(parts) < 11 {
			continue
		}

		tradeDate, _ := time.Parse("2006-01-02", parts[0])
		open, _ := strconv.ParseFloat(parts[1], 64)
		close, _ := strconv.ParseFloat(parts[2], 64)
		high, _ := strconv.ParseFloat(parts[3], 64)
		low, _ := strconv.ParseFloat(parts[4], 64)
		volume, _ := strconv.ParseFloat(parts[5], 64)
		amount, _ := strconv.ParseFloat(parts[6], 64)
		amplitude, _ := strconv.ParseFloat(parts[7], 64)
		changePct, _ := strconv.ParseFloat(parts[8], 64)
		turnoverRate, _ := strconv.ParseFloat(parts[10], 64)

		daily := model.StockDaily{
			Code:         code,
			TradeDate:    tradeDate,
			Open:         &open,
			Close:        &close,
			High:         &high,
			Low:          &low,
			Volume:       &volume,
			Amount:       &amount,
			Amplitude:    &amplitude,
			ChangePct:    &changePct,
			TurnoverRate: &turnoverRate,
		}

		// UPSERT
		var existing model.StockDaily
		result := DB.Where("code = ? AND trade_date = ?", code, tradeDate).First(&existing)
		if result.Error == nil {
			DB.Model(&existing).Updates(map[string]interface{}{
				"open":          daily.Open,
				"close":         daily.Close,
				"high":          daily.High,
				"low":           daily.Low,
				"volume":        daily.Volume,
				"amount":        daily.Amount,
				"amplitude":     daily.Amplitude,
				"change_pct":    daily.ChangePct,
				"turnover_rate": daily.TurnoverRate,
			})
		} else {
			DB.Create(&daily)
		}
	}

	return nil
}

// SyncFinanceData 同步财务数据
func (s *StockSyncService) SyncFinanceData(code string) error {
	url := fmt.Sprintf("https://emweb.securities.eastmoney.com/PC_HSF10/NewFinanceAnalysis/ZYZBAjaxNew?type=0&code=%s.%s", s.getSecId(code), code)

	body, err := s.httpGet(url)
	if err != nil {
		return err
	}

	var resp struct {
		Data []struct {
			Date       string  `json:"date"`
			JBMGSY     float64 `json:"jbmgsy"`     // 每股收益
			XSMGSY     float64 `json:"xsmgsy"`     // 每股净资产
			ZYYSZSR    float64 `json:"zyysnzsr"`   // 营收
			YYZSR_TBZZ float64 `json:"yyzsr_tbzz"` // 营收同比增长
			JLR        float64 `json:"jlr"`        // 净利润
			JLR_TBZZ   float64 `json:"jlr_tbzz"`   // 净利润同比增长
			ROEJQ      float64 `json:"roejq"`      // ROE
			MLL        float64 `json:"mll"`        // 毛利率
			JLL        float64 `json:"jll"`        // 净利率
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	for _, item := range resp.Data {
		reportDate, _ := time.Parse("2006-12-31", item.Date)
		if reportDate.IsZero() {
			reportDate, _ = time.Parse("2006-09-30", item.Date)
		}
		if reportDate.IsZero() {
			reportDate, _ = time.Parse("2006-06-30", item.Date)
		}
		if reportDate.IsZero() {
			reportDate, _ = time.Parse("2006-03-31", item.Date)
		}

		finance := model.StockFinance{
			Code:         code,
			ReportDate:   reportDate,
			Eps:          &item.JBMGSY,
			Bps:          &item.XSMGSY,
			RevenueYoy:   &item.YYZSR_TBZZ,
			NetProfitYoy: &item.JLR_TBZZ,
			Roe:          &item.ROEJQ,
			GrossMargin:  &item.MLL,
			NetMargin:    &item.JLL,
		}

		// UPSERT
		var existing model.StockFinance
		result := DB.Where("code = ? AND report_date = ?", code, reportDate).First(&existing)
		if result.Error == nil {
			DB.Model(&existing).Updates(map[string]interface{}{
				"eps":            finance.Eps,
				"bps":            finance.Bps,
				"revenue_yoy":    finance.RevenueYoy,
				"net_profit_yoy": finance.NetProfitYoy,
				"roe":            finance.Roe,
				"gross_margin":   finance.GrossMargin,
				"net_margin":     finance.NetMargin,
			})
		} else {
			DB.Create(&finance)
		}
	}

	return nil
}

// SyncConcepts 同步概念板块
func (s *StockSyncService) SyncConcepts() error {
	log.Println("[StockSync] 同步概念板块...")

	// 获取概念列表
	url := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=500&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:90+t:3+f:!50&fields=f12,f14"
	body, err := s.httpGet(url)
	if err != nil {
		return err
	}

	var conceptResp struct {
		Data struct {
			Diff []struct {
				Code string `json:"f12"`
				Name string `json:"f14"`
			} `json:"diff"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &conceptResp); err != nil {
		return err
	}

	// 清空旧数据
	DB.Where("concept_type = ?", "concept").Delete(&model.StockConcept{})

	for _, concept := range conceptResp.Data.Diff {
		// 获取概念成分股
		stockUrl := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=2000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=b:BK%s+f:!50&fields=f12,f14",
			concept.Code)
		stockBody, err := s.httpGet(stockUrl)
		if err != nil {
			continue
		}

		var stockResp struct {
			Data struct {
				Diff []struct {
					Code string `json:"f12"`
				} `json:"diff"`
			} `json:"data"`
		}

		if err := json.Unmarshal(stockBody, &stockResp); err != nil {
			continue
		}

		for _, stock := range stockResp.Data.Diff {
			sc := model.StockConcept{
				Code:        stock.Code,
				ConceptName: concept.Name,
				ConceptCode: concept.Code,
				ConceptType: "concept",
			}
			DB.Create(&sc)
		}
	}

	log.Println("[StockSync] 同步概念板块完成")
	return nil
}

// getSecId 获取东方财富 secid
func (s *StockSyncService) getSecId(code string) string {
	if len(code) == 6 {
		if code[:1] == "6" {
			return "1" // 上海
		}
		return "0" // 深圳
	}
	return "0"
}

// stripJSONP 去掉 JSONP 包装
func (s *StockSyncService) stripJSONP(data string) string {
	if len(data) > 0 {
		start := strings.Index(data, "(")
		end := strings.LastIndex(data, ")")
		if start >= 0 && end > start {
			return data[start+1 : end]
		}
	}
	return data
}

// splitKline 分割K线数据
func (s *StockSyncService) splitKline(line string) []string {
	var result []string
	current := ""
	for _, ch := range line {
		if ch == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// httpGet 发送HTTP GET请求
func (s *StockSyncService) httpGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	var lastErr error
	for retry := 0; retry < 3; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * time.Second)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Referer", "https://quote.eastmoney.com/")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// 调试：打印前200个字符
		if len(data) > 200 {
			log.Printf("[StockSync] 响应前200字符: %s", string(data[:200]))
		} else {
			log.Printf("[StockSync] 响应内容: %s", string(data))
		}

		return data, nil
	}

	return nil, fmt.Errorf("请求失败(重试3次): %v", lastErr)
}
