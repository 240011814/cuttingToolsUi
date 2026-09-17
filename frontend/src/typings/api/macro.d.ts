declare namespace Api {
  namespace Macro {
    /** 存款准备金率 (数值为百分比) */
    interface ReserveRatio {
      id: number
      pubDate: string
      effectiveDate: string
      bigInstitutionsRatioPre: number | null
      bigInstitutionsRatioAfter: number | null
      mediumInstitutionsRatioPre: number | null
      mediumInstitutionsRatioAfter: number | null
    }

    /** 货币供应量月度数据 (余额单位: 亿元, 同比/环比单位: %) */
    interface MoneySupplyMonth {
      id: number
      statYear: number
      statMonth: number
      m0Month: number | null
      m0Yoy: number | null
      m0ChainRelative: number | null
      m1Month: number | null
      m1Yoy: number | null
      m1ChainRelative: number | null
      m2Month: number | null
      m2Yoy: number | null
      m2ChainRelative: number | null
    }

    /** 货币供应量年度数据(年底余额, 单位: 亿元) */
    interface MoneySupplyYear {
      id: number
      statYear: number
      m0Year: number | null
      m0YearYoy: number | null
      m1Year: number | null
      m1YearYoy: number | null
      m2Year: number | null
      m2YearYoy: number | null
    }

    /** 存款利率 (数值为百分比) */
    interface DepositRate {
      id: number
      pubDate: string
      demand: number | null
      fixed3Month: number | null
      fixed6Month: number | null
      fixed1Year: number | null
      fixed2Year: number | null
      fixed3Year: number | null
      fixed5Year: number | null
      installment1Year: number | null
      installment3Year: number | null
      installment5Year: number | null
    }

    /** 贷款利率 (数值为百分比, 含公积金贷款利率) */
    interface LoanRate {
      id: number
      pubDate: string
      loan6Month: number | null
      loan6MonthTo1Year: number | null
      loan1YearTo3Year: number | null
      loan3YearTo5Year: number | null
      loanAbove5Year: number | null
      mortgageRateBelow5Year: number | null
      mortgageRateAbove5Year: number | null
    }

    /** 贷款市场报价利率 LPR (数值为百分比) */
    interface LPR {
      id: number
      tradeDate: string
      lpr1Year: number | null
      lpr5Year: number | null
      rate1: number | null
      rate2: number | null
    }

    /** GDP数据 */
    interface GDP {
      id: number
      quarter: string
      gdpYoy: number | null
      gdpCumulative: number | null
      gdpCumulativeYoy: number | null
    }

    /** CPI数据 */
    interface CPI {
      id: number
      month: string
      cpiYoy: number | null
      cpiMom: number | null
      cpiCumulativeYoy: number | null
    }

    /** PMI数据 */
    interface PMI {
      id: number
      month: string
      pmiManufacturing: number | null
      pmiManufacturingYoy: number | null
      pmiNonManufacturing: number | null
      pmiNonManufacturingYoy: number | null
    }

    /** PPI数据 */
    interface PPI {
      id: number
      month: string
      ppiYoy: number | null
      ppiMom: number | null
      ppiCumulativeYoy: number | null
    }
  }
}