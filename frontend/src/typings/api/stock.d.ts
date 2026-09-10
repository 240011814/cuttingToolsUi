declare namespace Api {
  namespace Stock {
    interface FilterCondition {
      field: string
      operator: 'gt' | 'lt' | 'eq' | 'gte' | 'lte' | 'between' | 'in' | 'not_in'
      value: number | string | (number | string)[]
    }

    interface ScreenRequest {
      conditions: FilterCondition[]
      sortBy?: string
      sortOrder?: 'asc' | 'desc'
      page?: number
      pageSize?: number
      conceptNames?: string[]
      industries?: string[]
      markets?: string[]
      excludeSt?: boolean
      keyword?: string
    }

    interface ScreenResult {
      code: string
      name: string
      market: string
      industry: string
      price: number | null
      changePct: number | null
      turnoverRate: number | null
      amount: number | null
      marketCap: number | null
      floatMarketCap: number | null
      peTtm: number | null
      pb: number | null
      roe: number | null
      revenueYoy: number | null
      netProfitYoy: number | null
      grossMargin: number | null
      netMargin: number | null
      debtRatio: number | null
      currentRatio: number | null
      quickRatio: number | null
      cashRatio: number | null
      nrTurnRatio: number | null
      invTurnRatio: number | null
      yoyEquity: number | null
      yoyAsset: number | null
      cfoToOr: number | null
      change5d: number | null
      change20d: number | null
      ma5: number | null
      ma10: number | null
      ma20: number | null
      ma60: number | null
      rsi6: number | null
      rsi12: number | null
      rsi24: number | null
      concepts: string[] | null
    }

    interface ScreenResponse {
      list: ScreenResult[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }

    interface KlineData {
      date: string
      open: number | null
      high: number | null
      low: number | null
      close: number | null
      volume: number | null
      amount: number | null
      changePct: number | null
      ma5: number | null
      ma10: number | null
      ma20: number | null
    }

    interface FinanceHistory {
      code: string
      reportDate: string
      roe: number | null
      grossMargin: number | null
      netMargin: number | null
      revenue: number | null
      revenueYoy: number | null
      netProfit: number | null
      netProfitYoy: number | null
      eps: number | null
      bps: number | null
      debtRatio: number | null
      currentRatio: number | null
      quickRatio: number | null
      cashRatio: number | null
      nrTurnRatio: number | null
      invTurnRatio: number | null
      caTurnRatio: number | null
      assetTurnRatio: number | null
      yoyEquity: number | null
      yoyAsset: number | null
      yoyEps: number | null
      cfoToOr: number | null
      cfoToNp: number | null
    }

    interface FilterConditionSave {
      id: number
      userId: number
      name: string
      description: string
      conditions: string
      resultCount: number
      isPinned: boolean
      createdAt: string
      updatedAt: string
    }

    interface WatchlistItem {
      id: number
      userId: number
      code: string
      name: string | null
      groupName: string
      note: string
      createdAt: string
    }

    interface SyncStatus {
      running: boolean
      task: string
      startedAt: string
      finishedAt: string
      lastError: string
      progress: number
      total: number
    }
  }
}
