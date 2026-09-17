import { request } from '../request'

/** 存款利率历史 */
export function getDepositRates() {
  return request<Api.Macro.DepositRate[]>({
    url: '/api/stock/macro/deposit-rate',
    method: 'get'
  })
}

/** 贷款利率历史 */
export function getLoanRates() {
  return request<Api.Macro.LoanRate[]>({
    url: '/api/stock/macro/loan-rate',
    method: 'get'
  })
}

/** 存款准备金率历史 */
export function getReserveRatios() {
  return request<Api.Macro.ReserveRatio[]>({
    url: '/api/stock/macro/reserve-ratio',
    method: 'get'
  })
}

/** 货币供应量月度历史 */
export function getMoneySupplyMonth() {
  return request<Api.Macro.MoneySupplyMonth[]>({
    url: '/api/stock/macro/money-supply-month',
    method: 'get'
  })
}

/** 货币供应量年度历史(年底余额) */
export function getMoneySupplyYear() {
  return request<Api.Macro.MoneySupplyYear[]>({
    url: '/api/stock/macro/money-supply-year',
    method: 'get'
  })
}

/** 贷款市场报价利率 LPR 历史 */
export function getLPR() {
  return request<Api.Macro.LPR[]>({
    url: '/api/stock/macro/lpr',
    method: 'get'
  })
}

/** 同步宏观经济数据 */
export function syncMacroData() {
  return request<boolean>({
    url: '/api/stock/macro/sync',
    method: 'post'
  })
}