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

type StockSyncService struct {
	baostockURL string
}

func NewStockSyncService(baostockURL string) *StockSyncService {
	return &StockSyncService{baostockURL: baostockURL}
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

	// 1. 获取所有股票列表 (需要传入交易日参数)
	today := time.Now().Format("2006-01-02")
	stockListURL := fmt.Sprintf("%s/query_all_stock?day=%s", s.baostockURL, today)
	body, err := s.httpGetWithDelay(stockListURL)
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	log.Printf("[StockSync] 股票列表响应长度: %d", len(body))

	var stockListResp struct {
		Ok    bool `json:"ok"`
		Data  struct {
			Total int `json:"total"`
			Items []struct {
				Code        string `json:"code"`
				CodeName    string `json:"code_name"`
				TradeStatus string `json:"trade_status"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &stockListResp); err != nil {
		log.Printf("[StockSync] 解析股票列表失败, 响应前200字符: %s", string(body[:min(len(body), 200)]))
		return fmt.Errorf("解析股票列表失败: %v", err)
	}

	log.Printf("[StockSync] 股票列表解析结果: Ok=%v, Total=%d, Items数量=%d", stockListResp.Ok, stockListResp.Data.Total, len(stockListResp.Data.Items))

	if !stockListResp.Ok {
		return fmt.Errorf("获取股票列表失败")
	}

	// 2. 获取行业信息 (失败不影响股票列表同步)
	industryMap := make(map[string]string)
	industryURL := fmt.Sprintf("%s/query_stock_industry", s.baostockURL)
	industryBody, err := s.httpGetWithDelay(industryURL)
	if err != nil {
		log.Printf("[StockSync] 获取行业信息失败: %v, 将跳过行业信息", err)
	} else {
		var industryResp struct {
			Ok   bool `json:"ok"`
			Data struct {
				Total  int `json:"total"`
				Fields []string `json:"fields"`
				Items  []map[string]string `json:"items"`
			} `json:"data"`
		}

		if err := json.Unmarshal(industryBody, &industryResp); err == nil && industryResp.Ok {
			for _, item := range industryResp.Data.Items {
				code := item["code"]
				industry := item["industry"]
				if code != "" && industry != "" {
					industryMap[code] = industry
				}
			}
			log.Printf("[StockSync] 获取行业信息成功: %d 条", len(industryMap))
		}
	}

	// 3. 处理每只股票
	totalCount := 0
	skipCount := 0
	for _, item := range stockListResp.Data.Items {
		if item.Code == "" {
			skipCount++
			continue
		}

		// 解析市场代码 (sh.600000 -> SH, sz.000001 -> SZ)
		market := "SZ"
		if len(item.Code) > 3 && item.Code[:3] == "sh." {
			market = "SH"
		}

		// 提取纯数字代码
		code := item.Code
		if idx := strings.Index(code, "."); idx >= 0 {
			code = code[idx+1:]
		}

		stock := model.StockInfo{
			Code:   code,
			Name:   item.CodeName,
			Market: market,
		}

		// 设置行业信息
		if industry, ok := industryMap[item.Code]; ok && industry != "" {
			stock.Industry = industry
		}

		stock.IsST = strings.Contains(item.CodeName, "ST")
		stock.IsActive = item.TradeStatus == "1"

		result := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "market", "industry", "is_st", "is_active"}),
		}).Create(&stock)
		if result.Error != nil {
			log.Printf("[StockSync] 写入股票 %s 失败: %v", code, result.Error)
		} else {
			totalCount++
		}
	}

	log.Printf("[StockSync] 同步股票列表完成: %d 只 (跳过 %d 只)", totalCount, skipCount)
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
	// 转换代码格式: 600000 -> sh.600000
	baostockCode := s.convertToBaostockCode(code)

	// 获取最近30个交易日的K线数据
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -45).Format("2006-01-02") // 多取一些天数确保有30个交易日

	url := fmt.Sprintf("%s/query_history_k_data_plus?code=%s&fields=date,open,high,low,close,volume,amount,turn,pctChg&start_date=%s&end_date=%s&frequency=d&adjustflag=3",
		s.baostockURL, baostockCode, startDate, endDate)

	body, err := s.httpGetWithDelay(url)
	if err != nil {
		return fmt.Errorf("获取K线数据失败: %v", err)
	}

	var resp struct {
		Ok   bool `json:"ok"`
		Data struct {
			Total int `json:"total"`
			Items []struct {
				Date   string `json:"date"`
				Open   string `json:"open"`
				High   string `json:"high"`
				Low    string `json:"low"`
				Close  string `json:"close"`
				Volume string `json:"volume"`
				Amount string `json:"amount"`
				Turn   string `json:"turn"`
				PctChg string `json:"pctChg"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("解析K线数据失败: %v", err)
	}

	if !resp.Ok {
		return fmt.Errorf("获取K线数据失败")
	}

	count := 0
	for _, item := range resp.Data.Items {
		if item.Date == "" {
			continue
		}

		tradeDate, _ := time.Parse("2006-01-02", item.Date)
		open, _ := strconv.ParseFloat(item.Open, 64)
		close, _ := strconv.ParseFloat(item.Close, 64)
		high, _ := strconv.ParseFloat(item.High, 64)
		low, _ := strconv.ParseFloat(item.Low, 64)
		volume, _ := strconv.ParseFloat(item.Volume, 64)
		amount, _ := strconv.ParseFloat(item.Amount, 64)
		turnoverRate, _ := strconv.ParseFloat(item.Turn, 64)
		changePct, _ := strconv.ParseFloat(item.PctChg, 64)

		daily := model.StockDaily{
			Code:         code,
			TradeDate:    tradeDate,
			Open:         &open,
			Close:        &close,
			High:         &high,
			Low:          &low,
			Volume:       &volume,
			Amount:       &amount,
			ChangePct:    &changePct,
			TurnoverRate: &turnoverRate,
		}

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}, {Name: "trade_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "amount", "turnover_rate", "change_pct"}),
		}).Create(&daily)
		count++
	}

	log.Printf("[StockSync] 同步 %s K线数据: %d 条", code, count)
	return nil
}

// SyncFinanceData 同步财务数据
func (s *StockSyncService) SyncFinanceData(code string) error {
	baostockCode := s.convertToBaostockCode(code)

	// 获取最近4个季度的财务数据
	currentYear := time.Now().Year()
	financeData := make(map[string]*model.StockFinance)

	// 获取盈利能力数据
	for year := currentYear; year >= currentYear-1; year-- {
		for quarter := 4; quarter >= 1; quarter-- {
			url := fmt.Sprintf("%s/query_profit_data?code=%s&year=%d&quarter=%d",
				s.baostockURL, baostockCode, year, quarter)

			body, err := s.httpGetWithDelay(url)
			if err != nil {
				log.Printf("[StockSync] 获取 %s 盈利数据失败: %v", code, err)
				continue
			}

			var resp struct {
				Ok   bool `json:"ok"`
				Data struct {
					Total int `json:"total"`
					Items []struct {
						Code        string `json:"code"`
						StatDate    string `json:"statDate"`
						RoeAvg      string `json:"roeAvg"`
						NpMargin    string `json:"npMargin"`
						GpMargin    string `json:"gpMargin"`
						NetProfit   string `json:"netProfit"`
						EpsTTM      string `json:"epsTTM"`
						MBRevenue   string `json:"MBRevenue"`
						TotalShare  string `json:"totalShare"`
						LiqaShare   string `json:"liqaShare"`
					} `json:"items"`
				} `json:"data"`
			}

			if err := json.Unmarshal(body, &resp); err != nil {
				continue
			}

			if !resp.Ok || resp.Data.Total == 0 {
				continue
			}

			for _, item := range resp.Data.Items {
				if item.StatDate == "" {
					continue
				}

				reportDate, _ := time.Parse("2006-01-02", item.StatDate)
				key := fmt.Sprintf("%s_%s", code, item.StatDate)

				if _, exists := financeData[key]; !exists {
					financeData[key] = &model.StockFinance{
						Code:       code,
						ReportDate: reportDate,
					}
				}

				roe, _ := strconv.ParseFloat(item.RoeAvg, 64)
				grossMargin, _ := strconv.ParseFloat(item.GpMargin, 64)
				netMargin, _ := strconv.ParseFloat(item.NpMargin, 64)
				eps, _ := strconv.ParseFloat(item.EpsTTM, 64)
				revenue, _ := strconv.ParseFloat(item.MBRevenue, 64)
				netProfit, _ := strconv.ParseFloat(item.NetProfit, 64)
				totalShare, _ := strconv.ParseFloat(item.TotalShare, 64)
				floatShare, _ := strconv.ParseFloat(item.LiqaShare, 64)

				financeData[key].Roe = &roe
				financeData[key].GrossMargin = &grossMargin
				financeData[key].NetMargin = &netMargin
				financeData[key].Eps = &eps
				revenueInWan := revenue / 10000
				financeData[key].Revenue = &revenueInWan
				netProfitInWan := netProfit / 10000
				financeData[key].NetProfit = &netProfitInWan

				// 更新股本信息
				if totalShare > 0 {
					totalShareInWan := totalShare / 10000
					DB.Model(&model.StockInfo{}).Where("code = ?", code).Update("total_share", totalShareInWan)
				}
				if floatShare > 0 {
					floatShareInWan := floatShare / 10000
					DB.Model(&model.StockInfo{}).Where("code = ?", code).Update("float_share", floatShareInWan)
				}
			}
		}
	}

	// 获取偿债能力数据
	for year := currentYear; year >= currentYear-1; year-- {
		for quarter := 4; quarter >= 1; quarter-- {
			url := fmt.Sprintf("%s/query_balance_data?code=%s&year=%d&quarter=%d",
				s.baostockURL, baostockCode, year, quarter)

			body, err := s.httpGetWithDelay(url)
			if err != nil {
				continue
			}

			var resp struct {
				Ok   bool `json:"ok"`
				Data struct {
					Total int `json:"total"`
					Items []struct {
						Code             string `json:"code"`
						StatDate         string `json:"statDate"`
						CurrentRatio     string `json:"currentRatio"`
						QuickRatio       string `json:"quickRatio"`
						CashRatio        string `json:"cashRatio"`
						YOYLiability     string `json:"yoyLiability"`
						LiabilityToAsset string `json:"liabilityToAsset"`
						AssetToEquity    string `json:"assetToEquity"`
					} `json:"items"`
				} `json:"data"`
			}

			if err := json.Unmarshal(body, &resp); err != nil {
				continue
			}

			if !resp.Ok || resp.Data.Total == 0 {
				continue
			}

			for _, item := range resp.Data.Items {
				if item.StatDate == "" {
					continue
				}

				key := fmt.Sprintf("%s_%s", code, item.StatDate)

				if _, exists := financeData[key]; !exists {
					continue
				}

				currentRatio, _ := strconv.ParseFloat(item.CurrentRatio, 64)
				quickRatio, _ := strconv.ParseFloat(item.QuickRatio, 64)
				debtRatio, _ := strconv.ParseFloat(item.LiabilityToAsset, 64)

				financeData[key].CurrentRatio = &currentRatio
				financeData[key].QuickRatio = &quickRatio
				financeData[key].DebtRatio = &debtRatio
			}
		}
	}

	// 保存到数据库
	for _, finance := range financeData {
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}, {Name: "report_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"roe", "gross_margin", "net_margin", "eps", "revenue", "net_profit", "debt_ratio", "current_ratio", "quick_ratio"}),
		}).Create(finance)
	}

	log.Printf("[StockSync] 同步 %s 财务数据: %d 条", code, len(financeData))
	return nil
}

// SyncConcepts 同步概念板块
// 注意: baostock 没有概念板块接口，此处保留东方财富API实现
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

// convertToBaostockCode 将纯数字代码转换为baostock格式 (600000 -> sh.600000)
func (s *StockSyncService) convertToBaostockCode(code string) string {
	if len(code) == 6 {
		if code[:1] == "6" {
			return "sh." + code
		}
		return "sz." + code
	}
	return "sz." + code
}

// FetchRealtimeKline 获取实时K线数据
// 注意: baostock 不提供实时数据，此处保留东方财富API实现
func (s *StockSyncService) FetchRealtimeKline(code string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?cb=jQuery&secid=%s.%s&fields1=f1,f2,f3,f4,f5,f6&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&klt=101&fqt=1&end=20500101&lmt=120",
		s.getSecId(code), code)

	body, err := s.httpGetWithDelay(url)
	if err != nil {
		return nil, err
	}

	jsonStr := s.stripJSONP(string(body))

	var resp struct {
		Data struct {
			Name   string   `json:"name"`
			Code   string   `json:"code"`
			Klines []string `json:"klines"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("解析K线数据失败: %v", err)
	}

	var result []map[string]interface{}
	for _, line := range resp.Data.Klines {
		parts := s.splitKline(line)
		if len(parts) < 11 {
			continue
		}
		item := map[string]interface{}{
			"date":        parts[0],
			"open":        parts[1],
			"close":       parts[2],
			"high":        parts[3],
			"low":         parts[4],
			"volume":      parts[5],
			"amount":      parts[6],
			"amplitude":   parts[7],
			"changePct":   parts[8],
			"changeAmt":   parts[9],
			"turnoverRate": parts[10],
		}
		result = append(result, item)
	}

	return result, nil
}

// FetchRealtimeQuote 获取实时行情
// 注意: baostock 不提供实时数据，此处保留东方财富API实现
func (s *StockSyncService) FetchRealtimeQuote(code string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/get?cb=jQuery&secid=%s.%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f60,f71,f116,f117,f162,f167,f168,f169,f170,f171,f292",
		s.getSecId(code), code)

	body, err := s.httpGetWithDelay(url)
	if err != nil {
		return nil, err
	}

	jsonStr := s.stripJSONP(string(body))

	var resp struct {
		Data struct {
			Name       string  `json:"f58"`
			Code       string  `json:"f57"`
			Price      float64 `json:"f43"`
			Open       float64 `json:"f46"`
			High       float64 `json:"f44"`
			Low        float64 `json:"f45"`
			PreClose   float64 `json:"f60"`
			Volume     float64 `json:"f47"`
			Amount     float64 `json:"f48"`
			ChangePct  float64 `json:"f170"`
			ChangeAmt  float64 `json:"f169"`
			TurnoverRate float64 `json:"f168"`
			PeTtm      float64 `json:"f162"`
			Pb         float64 `json:"f167"`
			TotalCap   float64 `json:"f116"`
			FloatCap   float64 `json:"f117"`
			High52W    float64 `json:"f51"`
			Low52W     float64 `json:"f52"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("解析行情数据失败: %v", err)
	}

	d := resp.Data
	result := map[string]interface{}{
		"name":         d.Name,
		"code":         d.Code,
		"price":        d.Price / 100,
		"open":         d.Open / 100,
		"high":         d.High / 100,
		"low":          d.Low / 100,
		"preClose":     d.PreClose / 100,
		"volume":       d.Volume,
		"amount":       d.Amount,
		"changePct":    d.ChangePct / 100,
		"changeAmt":    d.ChangeAmt / 100,
		"turnoverRate": d.TurnoverRate / 100,
		"peTtm":        d.PeTtm / 100,
		"pb":           d.Pb / 100,
		"totalCap":     d.TotalCap / 100000000,
		"floatCap":     d.FloatCap / 100000000,
		"high52W":      d.High52W / 100,
		"low52W":       d.Low52W / 100,
	}

	return result, nil
}

// stripJSONP 去掉 JSONP 包装
func (s *StockSyncService) stripJSONP(data string) string {
	if len(data) > 0 {
		// 检查是否是JSONP格式 (jQuery({...}))
		if strings.HasPrefix(data, "jQuery") || strings.HasPrefix(data, "jsonp") {
			start := strings.Index(data, "(")
			end := strings.LastIndex(data, ")")
			if start >= 0 && end > start {
				return data[start+1 : end]
			}
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
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		client := &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:          10,
				IdleConnTimeout:       60 * time.Second,
				TLSHandshakeTimeout:   30 * time.Second,
				ResponseHeaderTimeout: 90 * time.Second,
				DisableCompression:    false,
				ForceAttemptHTTP2:     false,
			},
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*; q=0.01")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://quote.eastmoney.com/")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-site")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[StockSync] 请求失败: %v", err)
			lastErr = err
			continue
		}

		log.Printf("[StockSync] 响应状态码: %d, ContentLength: %d", resp.StatusCode, resp.ContentLength)

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Printf("[StockSync] 读取响应失败: %v", err)
			lastErr = err
			continue
		}

		if resp.StatusCode != 200 {
			log.Printf("[StockSync] 非200状态码: %d, 响应: %s", resp.StatusCode, string(data[:min(len(data), 200)]))
			lastErr = fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
			continue
		}
		return data, nil
	}

	return nil, fmt.Errorf("请求失败(重试3次): %v", lastErr)
}

// httpGetWithDelay 带延迟的HTTP请求
func (s *StockSyncService) httpGetWithDelay(url string) ([]byte, error) {
	data, err := s.httpGet(url)
	log.Printf("[StockSync] 请求URL: %s, 响应长度: %d, 错误: %v", url, len(data), err)
	if err == nil {
		time.Sleep(80 * time.Millisecond)
	}
	return data, err
}
