package model

import (
	"time"
)

// MacroReserveRatio 存款准备金率 (baostock query_required_reserve_ratio_data)
// 数值为百分比(如 9.5 表示 9.5%), 每条记录含调整前/调整后的大型与中小金融机构准备金率
type MacroReserveRatio struct {
	ID                           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PubDate                      time.Time `gorm:"not null;uniqueIndex:uk_rr_pub_eff" json:"pubDate"`
	EffectiveDate                time.Time `gorm:"not null;uniqueIndex:uk_rr_pub_eff" json:"effectiveDate"`
	BigInstitutionsRatioPre      *float64  `json:"bigInstitutionsRatioPre"`
	BigInstitutionsRatioAfter    *float64  `json:"bigInstitutionsRatioAfter"`
	MediumInstitutionsRatioPre   *float64  `json:"mediumInstitutionsRatioPre"`
	MediumInstitutionsRatioAfter *float64  `json:"mediumInstitutionsRatioAfter"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updatedAt"`
}

func (MacroReserveRatio) TableName() string {
	return "macro_reserve_ratio"
}

// MacroMoneySupplyMonth 货币供应量月度数据 (baostock query_money_supply_data_month)
// 余额单位为亿元, 同比/环比为百分比
type MacroMoneySupplyMonth struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StatYear        int       `gorm:"not null;uniqueIndex:uk_msm_year_month" json:"statYear"`
	StatMonth       int       `gorm:"not null;uniqueIndex:uk_msm_year_month" json:"statMonth"`
	M0Month         *float64  `json:"m0Month"`
	M0Yoy           *float64  `json:"m0Yoy"`
	M0ChainRelative *float64  `json:"m0ChainRelative"`
	M1Month         *float64  `json:"m1Month"`
	M1Yoy           *float64  `json:"m1Yoy"`
	M1ChainRelative *float64  `json:"m1ChainRelative"`
	M2Month         *float64  `json:"m2Month"`
	M2Yoy           *float64  `json:"m2Yoy"`
	M2ChainRelative *float64  `json:"m2ChainRelative"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (MacroMoneySupplyMonth) TableName() string {
	return "macro_money_supply_month"
}

// MacroMoneySupplyYear 货币供应量年度数据(年底余额) (baostock query_money_supply_data_year)
// 余额单位为亿元, 同比为百分比
type MacroMoneySupplyYear struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StatYear  int       `gorm:"not null;uniqueIndex" json:"statYear"`
	M0Year    *float64  `json:"m0Year"`
	M0YearYoy *float64  `json:"m0YearYoy"`
	M1Year    *float64  `json:"m1Year"`
	M1YearYoy *float64  `json:"m1YearYoy"`
	M2Year    *float64  `json:"m2Year"`
	M2YearYoy *float64  `json:"m2YearYoy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (MacroMoneySupplyYear) TableName() string {
	return "macro_money_supply_year"
}

// MacroDepositRate 存款利率 (baostock query_deposit_rate_data)
// 数值为百分比, pub_date 即调整生效的公告日期; 列名用显式 tag 避免 GORM 数字驼峰歧义
type MacroDepositRate struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PubDate       time.Time `gorm:"not null;uniqueIndex" json:"pubDate"`
	Demand        *float64  `gorm:"column:demand" json:"demand"`
	Fixed3Month   *float64  `gorm:"column:fixed_3m" json:"fixed3Month"`
	Fixed6Month   *float64  `gorm:"column:fixed_6m" json:"fixed6Month"`
	Fixed1Year    *float64  `gorm:"column:fixed_1y" json:"fixed1Year"`
	Fixed2Year    *float64  `gorm:"column:fixed_2y" json:"fixed2Year"`
	Fixed3Year    *float64  `gorm:"column:fixed_3y" json:"fixed3Year"`
	Fixed5Year    *float64  `gorm:"column:fixed_5y" json:"fixed5Year"`
	Installment1Y *float64  `gorm:"column:installment_1y" json:"installment1Year"`
	Installment3Y *float64  `gorm:"column:installment_3y" json:"installment3Year"`
	Installment5Y *float64  `gorm:"column:installment_5y" json:"installment5Year"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (MacroDepositRate) TableName() string {
	return "macro_deposit_rate"
}

// MacroLoanRate 贷款利率 (baostock query_loan_rate_data)
// 数值为百分比, 含各期限贷款基准利率与公积金贷款利率
type MacroLoanRate struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PubDate         time.Time `gorm:"not null;uniqueIndex" json:"pubDate"`
	Loan6Month      *float64  `gorm:"column:loan_6m" json:"loan6Month"`
	Loan6MonthTo1Y  *float64  `gorm:"column:loan_6m_1y" json:"loan6MonthTo1Year"`
	Loan1YTo3Y      *float64  `gorm:"column:loan_1y_3y" json:"loan1YearTo3Year"`
	Loan3YTo5Y      *float64  `gorm:"column:loan_3y_5y" json:"loan3YearTo5Year"`
	LoanAbove5Y     *float64  `gorm:"column:loan_above_5y" json:"loanAbove5Year"`
	MortgageBelow5Y *float64  `gorm:"column:mortgage_below_5y" json:"mortgageRateBelow5Year"`
	MortgageAbove5Y *float64  `gorm:"column:mortgage_above_5y" json:"mortgageRateAbove5Year"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (MacroLoanRate) TableName() string {
	return "macro_loan_rate"
}
