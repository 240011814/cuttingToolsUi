import { request } from '../request'

/** 股票筛选 */
export function stockScreen(data: Api.Stock.ScreenRequest) {
  return request<Api.Stock.ScreenResponse>({
    url: '/api/stock/screen',
    method: 'post',
    data
  })
}

/** 个股详情 */
export function stockDetail(code: string) {
  return request<Api.Stock.ScreenResult>({
    url: `/api/stock/${code}`,
    method: 'get'
  })
}

/** K线数据 */
export function stockKline(code: string, params?: { period?: string; count?: number }) {
  return request<Api.Stock.KlineData[]>({
    url: `/api/stock/${code}/kline`,
    method: 'get',
    params
  })
}

/** 行业列表 */
export function stockIndustries() {
  return request<string[]>({
    url: '/api/stock/industries',
    method: 'get'
  })
}

/** 概念列表 */
export function stockConcepts() {
  return request<{ name: string; code: string }[]>({
    url: '/api/stock/concepts',
    method: 'get'
  })
}

/** 保存筛选条件 */
export function saveFilterCondition(data: { name: string; description?: string; conditions: string }) {
  return request<boolean>({
    url: '/api/stock/filters',
    method: 'post',
    data
  })
}

/** 获取筛选条件列表 */
export function getFilterConditions() {
  return request<Api.Stock.FilterConditionSave[]>({
    url: '/api/stock/filters',
    method: 'get'
  })
}

/** 删除筛选条件 */
export function deleteFilterCondition(id: number) {
  return request<boolean>({
    url: `/api/stock/filters/${id}`,
    method: 'delete'
  })
}

/** 添加自选股 */
export function addWatchlist(data: { code: string; groupName?: string; note?: string }) {
  return request<boolean>({
    url: '/api/stock/watchlist',
    method: 'post',
    data
  })
}

/** 获取自选股列表 */
export function getWatchlist() {
  return request<Api.Stock.WatchlistItem[]>({
    url: '/api/stock/watchlist',
    method: 'get'
  })
}

/** 删除自选股 */
export function deleteWatchlist(id: number) {
  return request<boolean>({
    url: `/api/stock/watchlist/${id}`,
    method: 'delete'
  })
}

/** 同步股票列表 */
export function syncStockList() {
  return request<boolean>({
    url: '/api/stock/sync/stock-list',
    method: 'post'
  })
}

/** 同步行情数据 */
export function syncDailyQuotes() {
  return request<boolean>({
    url: '/api/stock/sync/daily-quotes',
    method: 'post'
  })
}

/** 同步财务数据 */
export function syncFinanceData(code: string) {
  return request<boolean>({
    url: '/api/stock/sync/finance',
    method: 'post',
    params: { code }
  })
}

/** 同步概念板块 */
export function syncConcepts() {
  return request<boolean>({
    url: '/api/stock/sync/concepts',
    method: 'post'
  })
}

/** 同步单只股票(行情+财务) */
export function syncSingleStock(code: string) {
  return request<boolean>({
    url: '/api/stock/sync/single',
    method: 'post',
    params: { code }
  })
}
