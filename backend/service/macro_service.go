package service

import "backend/model"

// GetReserveRatios 查询存款准备金率历史 (按生效日期升序)
func (s *StockService) GetReserveRatios() ([]model.MacroReserveRatio, error) {
	var list []model.MacroReserveRatio
	err := DB.Order("effective_date ASC").Find(&list).Error
	return list, err
}

// GetMoneySupplyMonth 查询货币供应量月度历史 (按年月升序)
func (s *StockService) GetMoneySupplyMonth() ([]model.MacroMoneySupplyMonth, error) {
	var list []model.MacroMoneySupplyMonth
	err := DB.Order("stat_year ASC, stat_month ASC").Find(&list).Error
	return list, err
}

// GetMoneySupplyYear 查询货币供应量年度历史 (按年份升序)
func (s *StockService) GetMoneySupplyYear() ([]model.MacroMoneySupplyYear, error) {
	var list []model.MacroMoneySupplyYear
	err := DB.Order("stat_year ASC").Find(&list).Error
	return list, err
}

// GetDepositRates 查询存款利率历史 (按公告日期升序)
func (s *StockService) GetDepositRates() ([]model.MacroDepositRate, error) {
	var list []model.MacroDepositRate
	err := DB.Order("pub_date ASC").Find(&list).Error
	return list, err
}

// GetLoanRates 查询贷款利率历史 (按公告日期升序)
func (s *StockService) GetLoanRates() ([]model.MacroLoanRate, error) {
	var list []model.MacroLoanRate
	err := DB.Order("pub_date ASC").Find(&list).Error
	return list, err
}

// GetLPR 查询贷款市场报价利率历史 (按交易日期升序)
func (s *StockService) GetLPR() ([]model.MacroLPR, error) {
	var list []model.MacroLPR
	err := DB.Order("trade_date ASC").Find(&list).Error
	return list, err
}
