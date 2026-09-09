package api

import (
	"backend/model"
	"backend/service"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	svc         *service.StockService
	syncService *service.StockSyncService
}

func NewStockHandler(svc *service.StockService) *StockHandler {
	return &StockHandler{svc: svc, syncService: service.NewStockSyncService()}
}

// HandleScreen 股票筛选
func (h *StockHandler) HandleScreen(c *gin.Context) {
	var req model.StockScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	result, err := h.svc.Screen(req)
	if err != nil {
		SendError(c, "500", "筛选失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleGetDetail 个股详情
func (h *StockHandler) HandleGetDetail(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		SendError(c, "400", "股票代码不能为空")
		return
	}

	result, err := h.svc.GetDetail(code)
	if err != nil {
		SendError(c, "404", err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleGetKline K线数据
func (h *StockHandler) HandleGetKline(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		SendError(c, "400", "股票代码不能为空")
		return
	}

	period := c.DefaultQuery("period", "daily")
	countStr := c.DefaultQuery("count", "120")
	count, _ := strconv.Atoi(countStr)

	result, err := h.svc.GetKline(code, period, count)
	if err != nil {
		SendError(c, "500", "获取K线失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleGetIndustries 行业列表
func (h *StockHandler) HandleGetIndustries(c *gin.Context) {
	result, err := h.svc.GetIndustries()
	if err != nil {
		SendError(c, "500", "获取行业列表失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleGetConcepts 概念列表
func (h *StockHandler) HandleGetConcepts(c *gin.Context) {
	result, err := h.svc.GetConcepts()
	if err != nil {
		SendError(c, "500", "获取概念列表失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleSaveFilterCondition 保存筛选条件
func (h *StockHandler) HandleSaveFilterCondition(c *gin.Context) {
	userID := GetUserID(c)
	var req model.StockFilterConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	if err := h.svc.SaveFilterCondition(userID, req); err != nil {
		SendError(c, "500", "保存失败: "+err.Error())
		return
	}

	SendSuccess(c, true)
}

// HandleListFilterConditions 获取筛选条件列表
func (h *StockHandler) HandleListFilterConditions(c *gin.Context) {
	userID := GetUserID(c)
	result, err := h.svc.ListFilterConditions(userID)
	if err != nil {
		SendError(c, "500", "查询失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleDeleteFilterCondition 删除筛选条件
func (h *StockHandler) HandleDeleteFilterCondition(c *gin.Context) {
	userID := GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "无效的ID")
		return
	}

	if err := h.svc.DeleteFilterCondition(userID, uint(id)); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}

	SendSuccess(c, true)
}

// HandleAddWatchlist 添加自选股
func (h *StockHandler) HandleAddWatchlist(c *gin.Context) {
	userID := GetUserID(c)
	var req model.StockWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误")
		return
	}

	if err := h.svc.AddWatchlist(userID, req); err != nil {
		SendError(c, "500", "添加失败: "+err.Error())
		return
	}

	SendSuccess(c, true)
}

// HandleListWatchlist 获取自选股列表
func (h *StockHandler) HandleListWatchlist(c *gin.Context) {
	userID := GetUserID(c)
	result, err := h.svc.ListWatchlist(userID)
	if err != nil {
		SendError(c, "500", "查询失败: "+err.Error())
		return
	}

	SendSuccess(c, result)
}

// HandleDeleteWatchlist 删除自选股
func (h *StockHandler) HandleDeleteWatchlist(c *gin.Context) {
	userID := GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendError(c, "400", "无效的ID")
		return
	}

	if err := h.svc.DeleteWatchlist(userID, uint(id)); err != nil {
		SendError(c, "500", "删除失败: "+err.Error())
		return
	}

	SendSuccess(c, true)
}

// HandleSyncStockList 同步股票列表
func (h *StockHandler) HandleSyncStockList(c *gin.Context) {
	go func() {
		if err := h.syncService.SyncStockList(); err != nil {
			log.Printf("同步股票列表失败: %v", err)
		}
	}()
	SendSuccess(c, true)
}

// HandleSyncDailyQuotes 同步行情数据(全量)
func (h *StockHandler) HandleSyncDailyQuotes(c *gin.Context) {
	go func() {
		if err := h.syncService.SyncDailyQuotes(); err != nil {
			log.Printf("同步行情数据失败: %v", err)
		}
	}()
	SendSuccess(c, true)
}

// HandleSyncSingleStock 同步单只股票行情+财务
func (h *StockHandler) HandleSyncSingleStock(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		SendError(c, "400", "股票代码不能为空")
		return
	}

	go func() {
		if err := h.syncService.SyncSingleStockDaily(code); err != nil {
			log.Printf("同步 %s 行情失败: %v", code, err)
		}
		time.Sleep(80 * time.Millisecond)
		if err := h.syncService.SyncFinanceData(code); err != nil {
			log.Printf("同步 %s 财务失败: %v", code, err)
		}
	}()
	SendSuccess(c, true)
}

// HandleSyncFinanceData 同步财务数据
func (h *StockHandler) HandleSyncFinanceData(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		SendError(c, "400", "股票代码不能为空")
		return
	}

	go func() {
		if err := h.syncService.SyncFinanceData(code); err != nil {
			log.Printf("同步财务数据失败: %v", err)
		}
	}()
	SendSuccess(c, true)
}

// HandleRealtimeKline 获取实时K线数据（直接调第三方API）
func (h *StockHandler) HandleRealtimeKline(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		SendError(c, "400", "股票代码不能为空")
		return
	}

	data, err := h.syncService.FetchRealtimeKline(code)
	if err != nil {
		SendError(c, "500", "获取数据失败: "+err.Error())
		return
	}

	SendSuccess(c, data)
}

// HandleRealtimeQuote 获取实时行情（直接调第三方API）
func (h *StockHandler) HandleRealtimeQuote(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		SendError(c, "400", "股票代码不能为空")
		return
	}

	data, err := h.syncService.FetchRealtimeQuote(code)
	if err != nil {
		SendError(c, "500", "获取数据失败: "+err.Error())
		return
	}

	SendSuccess(c, data)
}

// HandleSyncConcepts 同步概念板块
func (h *StockHandler) HandleSyncConcepts(c *gin.Context) {
	go func() {
		if err := h.syncService.SyncConcepts(); err != nil {
			log.Printf("同步概念板块失败: %v", err)
		}
	}()
	SendSuccess(c, true)
}
