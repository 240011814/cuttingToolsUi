package service

import (
	"backend/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm/clause"
)

// StockSyncService 股票数据同步服务
// 注意: baostock 代理有每日调用限额, 且要求串行调用, 所有同步任务必须复用同一 HTTP client 并按批次执行
type StockSyncService struct {
	baostockURL string
	httpClient  *http.Client

	mu         sync.Mutex
	running    bool
	task       string
	startedAt  time.Time
	finishedAt time.Time
	lastError  string
	progress   int
	total      int
}

// SyncStatus 同步任务状态
type SyncStatus struct {
	Running    bool      `json:"running"`
	Task       string    `json:"task"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	LastError  string    `json:"lastError"`
	Progress   int       `json:"progress"`
	Total      int       `json:"total"`
}

func NewStockSyncService(baostockURL string) *StockSyncService {
	return &StockSyncService{
		baostockURL: baostockURL,
		httpClient: &http.Client{
			// baostock 代理每个请求都要登录登出且有全局锁, 大查询耗时较长, 超时给足余量
			Timeout: 300 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:          20,
				IdleConnTimeout:       120 * time.Second,
				TLSHandshakeTimeout:   15 * time.Second,
				ResponseHeaderTimeout: 240 * time.Second,
			},
		},
	}
}

// StartTask 启动同步任务(防重入), 返回 false 表示已有任务在运行
func (s *StockSyncService) StartTask(task string, fn func() error) bool {
	s.mu.Lock()
	if s.running {
		current := s.task
		s.mu.Unlock()
		log.Printf("[StockSync] 任务 %s 被拒绝: %s 正在运行", task, current)
		return false
	}
	s.running = true
	s.task = task
	s.startedAt = time.Now()
	s.progress, s.total = 0, 0
	s.lastError = ""
	s.mu.Unlock()

	go func() {
		err := fn()
		s.mu.Lock()
		s.running = false
		s.finishedAt = time.Now()
		if err != nil {
			s.lastError = err.Error()
		}
		s.mu.Unlock()
		if err != nil {
			log.Printf("[StockSync] 任务 %s 失败: %v", task, err)
		} else {
			log.Printf("[StockSync] 任务 %s 完成", task)
		}
	}()
	return true
}

// GetStatus 获取当前同步任务状态
func (s *StockSyncService) GetStatus() SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SyncStatus{
		Running:    s.running,
		Task:       s.task,
		StartedAt:  s.startedAt,
		FinishedAt: s.finishedAt,
		LastError:  s.lastError,
		Progress:   s.progress,
		Total:      s.total,
	}
}

func (s *StockSyncService) setProgress(done, total int) {
	s.mu.Lock()
	s.progress, s.total = done, total
	s.mu.Unlock()
}

// fetchRecentTradeDays 查询最近的交易日列表(从新到旧, 1次API调用)
func (s *StockSyncService) fetchRecentTradeDays(limit int) ([]string, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	url := fmt.Sprintf("%s/query_trade_dates?start_date=%s&end_date=%s",
		s.baostockURL, start.Format("2006-01-02"), end.Format("2006-01-02"))
	body, err := s.httpGetWithDelay(url)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Ok   bool `json:"ok"`
		Data struct {
			Items []map[string]string `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析交易日数据失败: %v", err)
	}
	if !resp.Ok {
		return nil, fmt.Errorf("查询交易日失败")
	}

	days := make([]string, 0, limit)
	for i := len(resp.Data.Items) - 1; i >= 0 && len(days) < limit; i-- {
		if resp.Data.Items[i]["is_trading_day"] == "1" {
			days = append(days, resp.Data.Items[i]["calendar_date"])
		}
	}
	if len(days) == 0 {
		return nil, fmt.Errorf("近30天无交易日数据")
	}
	return days, nil
}

// fetchLatestTradeDay 查询最近一个交易日
func (s *StockSyncService) fetchLatestTradeDay() (string, error) {
	days, err := s.fetchRecentTradeDays(1)
	if err != nil {
		return "", err
	}
	return days[0], nil
}

// SyncStockList 同步股票列表 (API调用: 交易日1次 + 股票列表1~2次 + 行业1次)
func (s *StockSyncService) SyncStockList() error {
	log.Println("[StockSync] 同步股票列表...")

	// 1. 先查最近交易日列表; query_all_stock 需要传交易日, 且当日数据约17:30后才发布,
	// 当天为交易日但数据未发布时列表为空, 需要回退到上一交易日
	tradeDays, err := s.fetchRecentTradeDays(5)
	if err != nil {
		log.Printf("[StockSync] 查询交易日失败, 将按自然日回退尝试: %v", err)
		tradeDays = nil
	}

	// 2. 获取所有股票列表
	var items []struct {
		Code        string `json:"code"`
		CodeName    string `json:"code_name"`
		TradeStatus string `json:"trade_status"`
	}
	for i := 0; i < 15 && len(items) == 0; i++ {
		day := ""
		if i < len(tradeDays) {
			day = tradeDays[i]
		} else {
			day = time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		}
		stockListURL := fmt.Sprintf("%s/query_all_stock?day=%s", s.baostockURL, day)
		body, err := s.httpGetWithDelay(stockListURL)
		if err != nil {
			return fmt.Errorf("获取股票列表失败: %v", err)
		}

		var resp struct {
			Ok   bool `json:"ok"`
			Data struct {
				Items []struct {
					Code        string `json:"code"`
					CodeName    string `json:"code_name"`
					TradeStatus string `json:"trade_status"`
				} `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("解析股票列表失败: %v", err)
		}
		if resp.Ok && len(resp.Data.Items) > 0 {
			items = resp.Data.Items
			log.Printf("[StockSync] 股票列表取自 %s", day)
			break
		}
		log.Printf("[StockSync] %s 无股票列表数据, 尝试前一交易日", day)
	}
	if len(items) == 0 {
		return fmt.Errorf("未获取到股票列表数据")
	}
	log.Printf("[StockSync] 获取股票列表: %d 只", len(items))

	// 3. 获取行业信息 (失败不影响股票列表同步)
	industryMap := make(map[string]string)
	industryURL := fmt.Sprintf("%s/query_stock_industry", s.baostockURL)
	industryBody, err := s.httpGetWithDelay(industryURL)
	if err != nil {
		log.Printf("[StockSync] 获取行业信息失败, 将保留已有行业数据: %v", err)
	} else {
		var industryResp struct {
			Ok   bool `json:"ok"`
			Data struct {
				Items []map[string]string `json:"items"`
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

	// 4. 组装数据, 有行业的和无行业的分开, 避免行业接口失败或缺失时把已有行业清空
	withIndustry := make([]model.StockInfo, 0, len(items))
	withoutIndustry := make([]model.StockInfo, 0)
	for _, item := range items {
		if item.Code == "" {
			continue
		}

		market := "SZ"
		if strings.HasPrefix(item.Code, "sh.") {
			market = "SH"
		} else if strings.HasPrefix(item.Code, "bj.") {
			market = "BJ"
		}

		code := item.Code
		if idx := strings.Index(code, "."); idx >= 0 {
			code = code[idx+1:]
		}

		stock := model.StockInfo{
			Code:     code,
			Name:     item.CodeName,
			Market:   market,
			IsST:     strings.Contains(item.CodeName, "ST"),
			IsActive: item.TradeStatus == "1",
		}
		if industry, ok := industryMap[item.Code]; ok && industry != "" {
			stock.Industry = industry
			withIndustry = append(withIndustry, stock)
		} else {
			withoutIndustry = append(withoutIndustry, stock)
		}
	}

	if len(withIndustry) > 0 {
		if err := s.batchUpsertStocks(withIndustry, []string{"name", "market", "industry", "is_st", "is_active"}); err != nil {
			return err
		}
	}
	if len(withoutIndustry) > 0 {
		if err := s.batchUpsertStocks(withoutIndustry, []string{"name", "market", "is_st", "is_active"}); err != nil {
			return err
		}
	}

	log.Printf("[StockSync] 同步股票列表完成: %d 只 (含行业 %d 只)", len(withIndustry)+len(withoutIndustry), len(withIndustry))
	return nil
}

// batchUpsertStocks 按批次 upsert 股票列表, 减少数据库交互次数
func (s *StockSyncService) batchUpsertStocks(stocks []model.StockInfo, updateCols []string) error {
	const batchSize = 500
	for i := 0; i < len(stocks); i += batchSize {
		end := min(i+batchSize, len(stocks))
		batch := stocks[i:end]
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoUpdates: clause.AssignmentColumns(updateCols),
		}).Create(&batch).Error; err != nil {
			return fmt.Errorf("批量写入股票失败(第%d批): %v", i/batchSize+1, err)
		}
	}
	return nil
}

// klineFrequencies 同步的K线周期
var klineFrequencies = []string{"daily", "weekly", "monthly"}

// baostockFreq 库内周期标识 -> baostock frequency 参数
func baostockFreq(freq string) string {
	switch freq {
	case "weekly":
		return "w"
	case "monthly":
		return "m"
	default:
		return "d"
	}
}

// freqIsCurrent 判断某周期K线是否已到最新 (周/月线只含已完成周期, 按6天/31天窗口判断)
func freqIsCurrent(st time.Time, freq string, latestTradeDay time.Time) bool {
	if st.IsZero() || latestTradeDay.IsZero() {
		return false
	}
	switch freq {
	case "weekly":
		return !st.Before(latestTradeDay.AddDate(0, 0, -6))
	case "monthly":
		return !st.Before(latestTradeDay.AddDate(0, 0, -31))
	default:
		return isSameDate(st, latestTradeDay)
	}
}

// GetStoredKlineStarts 查询某只股票各周期K线在库中的最新交易日 (零值表示该周期无数据)
func (s *StockSyncService) GetStoredKlineStarts(code string) map[string]time.Time {
	type row struct {
		Frequency string    `gorm:"column:frequency"`
		TradeDate time.Time `gorm:"column:trade_date"`
	}
	var rows []row
	DB.Model(&model.StockDaily{}).
		Select("frequency, MAX(trade_date) AS trade_date").
		Where("code = ?", code).
		Group("frequency").
		Scan(&rows)
	starts := make(map[string]time.Time, len(rows))
	for _, r := range rows {
		starts[r.Frequency] = r.TradeDate
	}
	return starts
}

// LatestTradeDay 查询最近一个交易日 (失败返回零值)
func (s *StockSyncService) LatestTradeDay() time.Time {
	day, err := s.fetchLatestTradeDay()
	if err != nil {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", day)
	if err != nil {
		return time.Time{}
	}
	return t
}

// HasStoredStockData 判断该股票在库中是否已有行情或财务数据
func (s *StockSyncService) HasStoredStockData(code string) bool {
	var kCount, fCount int64
	DB.Model(&model.StockDaily{}).Where("code = ?", code).Limit(1).Count(&kCount)
	DB.Model(&model.StockFinance{}).Where("code = ?", code).Limit(1).Count(&fCount)
	return kCount > 0 || fCount > 0
}

// SyncDailyQuotes 同步行情数据 (串行, 日/周/月K线增量同步: 从各周期库中最新交易日起更新, 无数据则拉全部历史; force=true 强制全量)
func (s *StockSyncService) SyncDailyQuotes(force bool) error {
	log.Printf("[StockSync] 同步行情数据... (force=%v)", force)

	var stocks []model.StockInfo
	if err := DB.Where("is_active = ?", true).Find(&stocks).Error; err != nil {
		return fmt.Errorf("查询股票列表失败: %v", err)
	}

	// 1次API调用获取最近交易日, 用于跳过已是最新的周期
	var latestTradeDay time.Time
	if !force {
		latestTradeDay = s.LatestTradeDay()
		if latestTradeDay.IsZero() {
			log.Printf("[StockSync] 查询最近交易日失败, 将不跳过已有数据")
		}
	}

	// 每只股票每个周期在库中的最新交易日
	type latestRow struct {
		Code      string    `gorm:"column:code"`
		Frequency string    `gorm:"column:frequency"`
		TradeDate time.Time `gorm:"column:trade_date"`
	}
	var latestRows []latestRow
	if err := DB.Model(&model.StockDaily{}).
		Select("code, frequency, MAX(trade_date) AS trade_date").
		Group("code, frequency").
		Scan(&latestRows).Error; err != nil {
		return fmt.Errorf("查询已有行情日期失败: %v", err)
	}
	latestMap := make(map[string]map[string]time.Time, len(latestRows))
	for _, r := range latestRows {
		if latestMap[r.Code] == nil {
			latestMap[r.Code] = make(map[string]time.Time, 3)
		}
		latestMap[r.Code][r.Frequency] = r.TradeDate
	}

	updated := 0
	skipped := 0
	failed := 0
	bjSkipped := 0
	consecutiveFails := 0
	done := 0

	for _, stock := range stocks {
		done++
		s.setProgress(done, len(stocks))

		if isBJCode(stock.Code) {
			bjSkipped++
			continue
		}

		starts := latestMap[stock.Code]
		if starts == nil {
			starts = map[string]time.Time{}
		}
		if force {
			// 强制全量: 各周期清零起点
			starts = map[string]time.Time{}
		}

		rows, err := s.SyncSingleStockDaily(stock.Code, stock.Market, starts, latestTradeDay)
		if err != nil {
			failed++
			consecutiveFails++
			log.Printf("[StockSync] 同步 %s 失败: %v", stock.Code, err)
			if consecutiveFails >= 20 {
				return fmt.Errorf("连续20只股票同步失败, 中止本轮同步(已处理 %d/%d), 请检查 baostock 服务", done, len(stocks))
			}
			continue
		}
		consecutiveFails = 0
		if rows == 0 {
			skipped++
		} else {
			updated++
		}
	}

	log.Printf("[StockSync] 同步行情数据完成: 共 %d 只, 更新 %d 只, 已最新跳过 %d 只, 北交所跳过 %d 只, 失败 %d 只",
		len(stocks), updated, skipped, bjSkipped, failed)
	return nil
}

// SyncSingleStockDaily 同步单只股票/指数的日/周/月K线 (每周期1次API调用, 批量写入), 返回总写入条数
// starts 为各周期在库中的最新交易日(零值=该周期拉取全部历史); latestTradeDay 为最近交易日(用于跳过已最新周期, 零值=不跳过)
func (s *StockSyncService) SyncSingleStockDaily(code, market string, starts map[string]time.Time, latestTradeDay time.Time) (int, error) {
	if isBJCode(code) {
		return 0, fmt.Errorf("北交所股票 %s 暂不支持同步(baostock 无该市场数据)", code)
	}
	market = s.resolveMarket(code, market)
	baostockCode := s.convertToBaostockCode(code, market)
	isIndex := isIndexCode(code, market)
	endDate := time.Now().Format("2006-01-02")

	total := 0
	fetched := make([]string, 0, len(klineFrequencies))
	for _, freq := range klineFrequencies {
		start := starts[freq]
		if freqIsCurrent(start, freq, latestTradeDay) {
			continue
		}
		startDate := "1990-01-01"
		if !start.IsZero() {
			startDate = start.Format("2006-01-02")
		}

		// 指数走指数接口(无换手率字段), 个股走股票接口
		var url string
		if isIndex {
			url = fmt.Sprintf("%s/query_history_index_k_data_plus?code=%s&fields=date,open,high,low,close,volume,amount,pctChg&start_date=%s&end_date=%s&frequency=%s",
				s.baostockURL, baostockCode, startDate, endDate, baostockFreq(freq))
		} else {
			url = fmt.Sprintf("%s/query_history_k_data_plus?code=%s&fields=date,open,high,low,close,volume,amount,turn,pctChg&start_date=%s&end_date=%s&frequency=%s&adjustflag=3",
				s.baostockURL, baostockCode, startDate, endDate, baostockFreq(freq))
		}

		body, err := s.httpGetWithDelay(url)
		if err != nil {
			return total, fmt.Errorf("获取%sK线数据失败: %v", freq, err)
		}

		var resp struct {
			Ok   bool `json:"ok"`
			Data struct {
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
			return total, fmt.Errorf("解析%sK线数据失败: %v", freq, err)
		}
		if !resp.Ok {
			return total, fmt.Errorf("获取%sK线数据失败", freq)
		}

		dailies := make([]model.StockDaily, 0, len(resp.Data.Items))
		for _, item := range resp.Data.Items {
			if item.Date == "" {
				continue
			}
			tradeDate, err := time.Parse("2006-01-02", item.Date)
			if err != nil {
				continue
			}
			dailies = append(dailies, model.StockDaily{
				Code:         code,
				Frequency:    freq,
				TradeDate:    tradeDate,
				Open:         parseFloatPtr(item.Open),
				Close:        parseFloatPtr(item.Close),
				High:         parseFloatPtr(item.High),
				Low:          parseFloatPtr(item.Low),
				Volume:       parseFloatPtr(item.Volume),
				Amount:       parseFloatPtr(item.Amount),
				ChangePct:    parseFloatPtr(item.PctChg),
				TurnoverRate: parseFloatPtr(item.Turn),
			})
		}

		if len(dailies) == 0 {
			log.Printf("[StockSync] 同步 %s %s K线: 0 条(%s 起无新数据)", code, freq, startDate)
			continue
		}
		// 全量历史可能上万条, 分批写入避免超过 max_allowed_packet
		const batchSize = 2000
		for i := 0; i < len(dailies); i += batchSize {
			end := min(i+batchSize, len(dailies))
			batch := dailies[i:end]
			if err := DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "code"}, {Name: "frequency"}, {Name: "trade_date"}},
				DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "amount", "turnover_rate", "change_pct"}),
			}).Create(&batch).Error; err != nil {
				return total, fmt.Errorf("写入 %s %s K线数据失败: %v", code, freq, err)
			}
		}

		fetched = append(fetched, fmt.Sprintf("%s=%d条(%s~%s)", freq, len(dailies), dailies[0].TradeDate.Format("2006-01-02"), dailies[len(dailies)-1].TradeDate.Format("2006-01-02")))
		total += len(dailies)
	}
	if len(fetched) > 0 {
		log.Printf("[StockSync] 同步 %s K线数据: %s", code, strings.Join(fetched, ", "))
	}
	return total, nil
}

// financeEndpoints baostock 五类季频财务数据来源: DB字段更新列与接口一一对应
var financeEndpoints = []struct {
	Code    string   // 来源标记: P盈利 G成长 O营运 C现金流 B偿债
	Path    string   // 代理接口路径
	Columns []string // 该来源覆盖的 DB 更新列
}{
	{"P", "query_profit_data", []string{"roe", "gross_margin", "net_margin", "eps", "revenue", "net_profit"}},
	{"G", "query_growth_data", []string{"net_profit_yoy", "yoy_equity", "yoy_asset", "yoy_eps"}},
	{"O", "query_operation_data", []string{"nr_turn_ratio", "inv_turn_ratio", "ca_turn_ratio", "asset_turn_ratio"}},
	{"C", "query_cash_flow_data", []string{"cfo_to_or", "cfo_to_np"}},
	{"B", "query_balance_data", []string{"current_ratio", "quick_ratio", "debt_ratio", "cash_ratio"}},
}

// SyncFinanceData 同步单只股票财务数据(2007Q1起全部历史)
// 先批量查本地(finance_sources 标记各报告期已同步的来源), 只对缺失的报告期和来源调用接口(单季度调用)
func (s *StockSyncService) SyncFinanceData(code, market string) (int, error) {
	if isBJCode(code) {
		return 0, fmt.Errorf("北交所股票 %s 暂不支持同步(baostock 无该市场数据)", code)
	}
	market = s.resolveMarket(code, market)
	if isIndexCode(code, market) {
		log.Printf("[StockSync] %s 是指数, 无财务数据, 跳过", code)
		return 0, nil
	}
	baostockCode := s.convertToBaostockCode(code, market)

	// 1. 批量查本地: 已有报告期、各报告期已同步的数据来源、已入库营收
	type storedRow struct {
		ReportDate     time.Time `gorm:"column:report_date"`
		Revenue        *float64  `gorm:"column:revenue"`
		FinanceSources string    `gorm:"column:finance_sources"`
	}
	var stored []storedRow
	if err := DB.Model(&model.StockFinance{}).Select("report_date, revenue, finance_sources").Where("code = ?", code).Find(&stored).Error; err != nil {
		return 0, fmt.Errorf("查询已有财务数据失败: %v", err)
	}
	quarterOf := func(t time.Time) [2]int {
		return [2]int{t.Year(), (int(t.Month()) + 2) / 3}
	}
	// 兼容旧数据: 无来源标记的行视为已有利润+偿债(旧逻辑只同步这两类)
	storedSources := make(map[[2]int]string, len(stored))
	dateByQuarter := make(map[[2]int]time.Time, len(stored))
	revenueByQuarter := make(map[[2]int]float64, len(stored))
	for _, r := range stored {
		pair := quarterOf(r.ReportDate)
		src := r.FinanceSources
		if src == "" {
			src = "PB"
		}
		storedSources[pair] = src
		dateByQuarter[pair] = r.ReportDate
		if r.Revenue != nil {
			revenueByQuarter[pair] = *r.Revenue
		}
	}

	currentYear := time.Now().Year()
	currentQuarter := (int(time.Now().Month()) + 2) / 3
	const financeStartYear = 2007
	const maxConsecutiveEmpty = 8
	const maxConsecutiveErrors = 5

	// 2. 计算各季度缺失的数据来源(从新到旧排列)
	type quarterRef struct{ y, q int }
	missing := make(map[string][]quarterRef, len(financeEndpoints))
	for y, q := currentYear, currentQuarter; y >= financeStartYear; {
		pair := [2]int{y, q}
		have := storedSources[pair]
		for _, ep := range financeEndpoints {
			if !strings.Contains(have, ep.Code) {
				missing[ep.Code] = append(missing[ep.Code], quarterRef{y, q})
			}
		}
		if q == 1 {
			y--
			q = 4
		} else {
			q--
		}
	}

	financeData := make(map[string]*model.StockFinance)
	touchedSources := make(map[string]string) // statDate -> 本次已同步来源
	touchedDates := make(map[string]time.Time)
	var totalShareVal, floatShareVal float64
	var shareDate time.Time
	var abortErr error

	getFinanceRow := func(statDate string, reportDate time.Time) *model.StockFinance {
		f, exists := financeData[statDate]
		if !exists {
			f = &model.StockFinance{Code: code, ReportDate: reportDate}
			financeData[statDate] = f
		}
		return f
	}

	// 3. 逐来源补缺, 每个来源只请求缺失的报告期
	// 注意区分: 连续空季度=数据自然终点(正常结束), 连续请求失败=服务异常(报错中止)
	for _, ep := range financeEndpoints {
		if abortErr != nil {
			break
		}
		pending := missing[ep.Code]
		if len(pending) == 0 {
			continue
		}
		emptyStreak := 0
		errStreak := 0
		var epStatDates []string
		for _, qr := range pending {
			if emptyStreak >= maxConsecutiveEmpty || errStreak >= maxConsecutiveErrors {
				break
			}
			url := fmt.Sprintf("%s/%s?code=%s&year=%d&quarter=%d",
				s.baostockURL, ep.Path, baostockCode, qr.y, qr.q)
			body, err := s.httpGetWithDelay(url)
			if err != nil {
				log.Printf("[StockSync] 获取 %s %d年Q%d %s数据失败: %v", code, qr.y, qr.q, ep.Code, err)
				errStreak++
				if errStreak >= maxConsecutiveErrors {
					abortErr = fmt.Errorf("baostock 服务连续异常, 财务同步中止")
					break
				}
				continue
			}
			errStreak = 0

			var resp struct {
				Ok   bool `json:"ok"`
				Data struct {
					Items []map[string]string `json:"items"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &resp); err != nil || !resp.Ok || len(resp.Data.Items) == 0 {
				emptyStreak++
				continue
			}
			emptyStreak = 0

			for _, item := range resp.Data.Items {
				statDate := item["statDate"]
				if statDate == "" {
					continue
				}
				reportDate, err := time.Parse("2006-01-02", statDate)
				if err != nil {
					continue
				}
				f := getFinanceRow(statDate, reportDate)
				switch ep.Code {
				case "P":
					f.Roe = parseFloatPtr(item["roeAvg"])
					f.GrossMargin = parseFloatPtr(item["gpMargin"])
					f.NetMargin = parseFloatPtr(item["npMargin"])
					f.Eps = parseFloatPtr(item["epsTTM"])
					if v := parseFloatPtr(item["MBRevenue"]); v != nil {
						revenue := *v / 10000
						f.Revenue = &revenue
					}
					if v := parseFloatPtr(item["netProfit"]); v != nil {
						netProfit := *v / 10000
						f.NetProfit = &netProfit
					}
					// 股本取最新报告期的数据
					if reportDate.After(shareDate) {
						if ts := parseFloatPtr(item["totalShare"]); ts != nil && *ts > 0 {
							totalShareVal = *ts
							shareDate = reportDate
						}
						if fs := parseFloatPtr(item["liqaShare"]); fs != nil && *fs > 0 {
							floatShareVal = *fs
							shareDate = reportDate
						}
					}
				case "G":
					// baostock 返回小数(0.047=4.7%), 统一转成百分比存储, 与营收同比口径一致
					if v := parseFloatPtr(item["YOYNI"]); v != nil {
						yoy := *v * 100
						f.NetProfitYoy = &yoy
					}
					if v := parseFloatPtr(item["YOYEquity"]); v != nil {
						yoy := *v * 100
						f.YoyEquity = &yoy
					}
					if v := parseFloatPtr(item["YOYAsset"]); v != nil {
						yoy := *v * 100
						f.YoyAsset = &yoy
					}
					if v := parseFloatPtr(item["YOYEPSBasic"]); v != nil {
						yoy := *v * 100
						f.YoyEps = &yoy
					}
				case "O":
					f.NrTurnRatio = parseFloatPtr(item["NRTurnRatio"])
					f.InvTurnRatio = parseFloatPtr(item["INVTurnRatio"])
					f.CaTurnRatio = parseFloatPtr(item["CATurnRatio"])
					f.AssetTurnRatio = parseFloatPtr(item["AssetTurnRatio"])
				case "C":
					f.CfoToOr = parseFloatPtr(item["CFOToOR"])
					f.CfoToNp = parseFloatPtr(item["CFOToNP"])
				case "B":
					f.CurrentRatio = parseFloatPtr(item["currentRatio"])
					f.QuickRatio = parseFloatPtr(item["quickRatio"])
					f.DebtRatio = parseFloatPtr(item["liabilityToAsset"])
					f.CashRatio = parseFloatPtr(item["cashRatio"])
				}
				touchedSources[statDate] += ep.Code
				touchedDates[statDate] = reportDate
				dateByQuarter[quarterOf(reportDate)] = reportDate
				if f.Revenue != nil {
					revenueByQuarter[quarterOf(reportDate)] = *f.Revenue
				}
				epStatDates = append(epStatDates, statDate)
			}
		}

		// 该来源按自有字段批量入库, 避免覆盖其他来源已写入的字段
		if len(epStatDates) > 0 {
			list := make([]*model.StockFinance, 0, len(epStatDates))
			for _, sd := range epStatDates {
				list = append(list, financeData[sd])
			}
			if err := DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "code"}, {Name: "report_date"}},
				DoUpdates: clause.AssignmentColumns(ep.Columns),
			}).Create(&list).Error; err != nil {
				return 0, fmt.Errorf("写入 %s %s 财务数据失败: %v", code, ep.Code, err)
			}
		}
	}

	// 4. 更新股本信息
	updates := map[string]interface{}{}
	if totalShareVal > 0 {
		updates["total_share"] = totalShareVal / 10000
	}
	if floatShareVal > 0 {
		updates["float_share"] = floatShareVal / 10000
	}
	if len(updates) > 0 {
		if err := DB.Model(&model.StockInfo{}).Where("code = ?", code).Updates(updates).Error; err != nil {
			log.Printf("[StockSync] 更新 %s 股本信息失败: %v", code, err)
		}
	}

	// 5. 计算营收同比增长: (本期营收-上年同期营收)/|上年同期营收|*100 (口径与baostock成长数据一致)
	revenueYoyRows := make([]*model.StockFinance, 0)
	for pair, rev := range revenueByQuarter {
		prevRev, ok := revenueByQuarter[[2]int{pair[0] - 1, pair[1]}]
		if !ok || prevRev == 0 {
			continue
		}
		rd, ok := dateByQuarter[pair]
		if !ok {
			continue
		}
		yoy := (rev - prevRev) / math.Abs(prevRev) * 100
		revenueYoyRows = append(revenueYoyRows, &model.StockFinance{Code: code, ReportDate: rd, RevenueYoy: &yoy})
	}
	if len(revenueYoyRows) > 0 {
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}, {Name: "report_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"revenue_yoy"}),
		}).Create(&revenueYoyRows).Error; err != nil {
			log.Printf("[StockSync] 更新 %s 营收同比增长失败: %v", code, err)
		}
	}

	// 6. 更新各报告期的数据来源标记
	if len(touchedSources) > 0 {
		list := make([]*model.StockFinance, 0, len(touchedSources))
		for sd, src := range touchedSources {
			list = append(list, &model.StockFinance{Code: code, ReportDate: touchedDates[sd], FinanceSources: src})
		}
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}, {Name: "report_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"finance_sources"}),
		}).Create(&list).Error; err != nil {
			log.Printf("[StockSync] 更新 %s 财务来源标记失败: %v", code, err)
		}
	}

	if abortErr != nil {
		// 已获取的部分已落库, 但整体按失败返回, 让前端感知到同步不完整
		log.Printf("[StockSync] 同步 %s 财务数据中止: %v (已处理 %d 期)", code, abortErr, len(touchedSources))
		return len(touchedSources), abortErr
	}

	log.Printf("[StockSync] 同步 %s 财务数据: %d 期", code, len(touchedSources))
	return len(touchedSources), nil
}

// convertToBaostockCode 将代码转换为baostock格式, 优先使用市场标识 (000003+SH -> sh.000003)
func (s *StockSyncService) convertToBaostockCode(code, market string) string {
	switch market {
	case "SH":
		return "sh." + code
	case "BJ":
		return "bj." + code
	case "SZ":
		return "sz." + code
	}
	// 无市场信息时按代码规则推断(仅对无歧义的代码有效)
	if len(code) == 6 && code[:1] == "6" {
		return "sh." + code
	}
	return "sz." + code
}

// resolveMarket 解析股票的市场标识: 优先用传入值, 其次查数据库, 最后按代码推断
func (s *StockSyncService) resolveMarket(code, market string) string {
	if market != "" {
		return market
	}
	var m string
	if err := DB.Model(&model.StockInfo{}).Select("market").Where("code = ?", code).Scan(&m).Error; err == nil && m != "" {
		return m
	}
	if len(code) == 6 {
		if code[0] == '6' {
			return "SH"
		}
		if isBJCode(code) {
			return "BJ"
		}
	}
	return "SZ"
}

// isIndexCode 指数代码: 上证指数 sh.000xxx / 深证指数 sz.399xxx (个股与指数代码有重叠, 必须结合市场判断)
func isIndexCode(code, market string) bool {
	if len(code) != 6 {
		return false
	}
	if market == "SH" && strings.HasPrefix(code, "000") {
		return true
	}
	if market == "SZ" && strings.HasPrefix(code, "399") {
		return true
	}
	return false
}

// isBJCode 判断是否北交所代码 (baostock 无北交所数据)
func isBJCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	return code[0] == '4' || code[0] == '8' || strings.HasPrefix(code, "92")
}

// isSameDate 判断两个时间是否同一天
func isSameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// parseFloatPtr 解析浮点数, 空串或非法值返回 nil (存 NULL 而不是 0, 避免污染数据)
func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// httpGet 发送HTTP GET请求 (复用 client 连接, 429限额错误不重试)
func (s *StockSyncService) httpGet(url string) ([]byte, error) {
	var lastErr error
	for retry := 0; retry < 3; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * 5 * time.Second)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*; q=0.01")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Connection", "keep-alive")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("[StockSync] 请求失败: %v", err)
			lastErr = err
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("[StockSync] 读取响应失败: %v", err)
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("baostock 每日调用限额已用尽, 请明日再试或调整 BAOSTOCK_API_DAILY_LIMIT")
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

// httpGetWithDelay 带延迟的HTTP请求 (串行调用 + 成功后短暂延迟, 避免触发第三方限流)
func (s *StockSyncService) httpGetWithDelay(url string) ([]byte, error) {
	data, err := s.httpGet(url)
	if err != nil {
		log.Printf("[StockSync] 请求失败: %s, 错误: %v", url, err)
	} else {
		time.Sleep(80 * time.Millisecond)
	}
	return data, err
}
