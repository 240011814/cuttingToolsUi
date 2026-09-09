package service

import (
	"backend/model"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"
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
		// 东方财富行情接口，包含行业和股本
		url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?cb=jQuery&pn=%d&pz=%d&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f12,f14,f13,f100,f20,f21",
			page, pageSize)

		body, err := s.httpGetWithDelay(url)
		if err != nil {
			return err
		}

		jsonStr := s.stripJSONP(string(body))

		var resp struct {
			Data struct {
				Total int `json:"total"`
				Diff  []struct {
					Code       string  `json:"f12"`
					Name       string  `json:"f14"`
					Market     int     `json:"f13"`
					Industry   string  `json:"f100"`
					TotalShare float64 `json:"f20"`
					FloatShare float64 `json:"f21"`
				} `json:"diff"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
			return fmt.Errorf("解析股票列表失败: %v", err)
		}

		if len(resp.Data.Diff) == 0 {
			break
		}

		for _, item := range resp.Data.Diff {
			if item.Code == "" {
				continue
			}

			market := "SZ"
			if item.Market == 1 {
				market = "SH"
			}

			stock := model.StockInfo{
				Code:   item.Code,
				Name:   item.Name,
				Market: market,
			}

			if item.Industry != "" && item.Industry != "-" {
				stock.Industry = item.Industry
			}

			// f20/f21 是市值(元)，转为万股需要除以股价，但这里没股价
			// 直接存原始值，在查询时用市值
			if item.TotalShare > 0 {
				totalShare := item.TotalShare / 10000
				stock.TotalShare = &totalShare
			}
			if item.FloatShare > 0 {
				floatShare := item.FloatShare / 10000
				stock.FloatShare = &floatShare
			}

			stock.IsST = strings.Contains(item.Name, "ST")
			stock.IsActive = true

			// f20/f21 是市值(元)
			if item.TotalShare > 0 {
				stock.TotalMarketCap = &item.TotalShare
			}
			if item.FloatShare > 0 {
				stock.FloatMarketCap = &item.FloatShare
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "code"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "market", "industry", "is_st", "total_share", "float_share", "total_market_cap", "float_market_cap"}),
			}).Create(&stock)
			totalCount++
		}

		log.Printf("[StockSync] 已同步 %d/%d 只股票", totalCount, resp.Data.Total)

		if len(resp.Data.Diff) < pageSize {
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
		if err := s.SyncSingleStockDaily(stock.Code); err != nil {
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

// SyncSingleStockDaily 同步单只股票日K线
func (s *StockSyncService) SyncSingleStockDaily(code string) error {
	url := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?cb=jQuery&secid=%s.%s&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&klt=101&fqt=1&end=20500101&lmt=30",
		s.getSecId(code), code)

	body, err := s.httpGetWithDelay(url)
	if err != nil {
		return err
	}

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

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}, {Name: "trade_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "amount", "turnover_rate", "change_pct", "amplitude"}),
		}).Create(&daily)
	}

	return nil
}

// SyncFinanceData 同步财务数据
func (s *StockSyncService) SyncFinanceData(code string) error {
	// 使用东方财富财务数据接口
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_F10_FINANCE_MAINFINADATA&columns=ALL&filter=(SECURITY_CODE%%3D%%22%s%%22)&pageSize=4&sortColumns=REPORT_DATE&sortTypes=-1",
		code)

	body, err := s.httpGetWithDelay(url)
	if err != nil {
		return err
	}

	var resp struct {
		Result struct {
			Data []struct {
				ReportDate       string  `json:"REPORT_DATE"`
				EPSJB            float64 `json:"EPSJB"`
				BPS              float64 `json:"BPS"`
				ROEJQ            float64 `json:"ROEJQ"`
				RevenueYoy       float64 `json:"TOTALOPERATEREVETZ"`
				NetProfitYoy     float64 `json:"PARENTNETPROFITTZ"`
				GrossProfitRatio float64 `json:"XSMLL"`
				NetProfitRatio   float64 `json:"XSJLL"`
			} `json:"data"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	for _, item := range resp.Result.Data {
		reportDate, _ := time.Parse("2006-01-02", item.ReportDate[:10])

		finance := model.StockFinance{
			Code:         code,
			ReportDate:   reportDate,
			Eps:          &item.EPSJB,
			Bps:          &item.BPS,
			Roe:          &item.ROEJQ,
			RevenueYoy:   &item.RevenueYoy,
			NetProfitYoy: &item.NetProfitYoy,
			GrossMargin:  &item.GrossProfitRatio,
			NetMargin:    &item.NetProfitRatio,
		}

		// UPSERT
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}, {Name: "report_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"roe", "gross_margin", "net_margin", "revenue_yoy", "net_profit_yoy", "eps", "bps"}),
		}).Create(&finance)
	}

	return nil
}

// SyncConcepts 同步概念板块
func (s *StockSyncService) SyncConcepts() error {
	log.Println("[StockSync] 同步概念板块...")

	// 获取概念列表
	url := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=500&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:90+t:3+f:!50&fields=f12,f14"
	body, err := s.httpGetWithDelay(url)
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
		stockBody, err := s.httpGetWithDelay(stockUrl)
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
	var lastErr error
	for retry := 0; retry < 3; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * time.Second)
		}

		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Referer", "https://quote.eastmoney.com/")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()

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

// httpGetWithDelay 带延迟的HTTP请求
func (s *StockSyncService) httpGetWithDelay(url string) ([]byte, error) {
	data, err := s.httpGet(url)
	if err == nil {
		time.Sleep(80 * time.Millisecond)
	}
	return data, err
}
