<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroLoanRate' })

const props = defineProps<{
  data: Api.Macro.LoanRate[]
  loading: boolean
}>()

const legendNames = ['贷款6个月至1年', '贷款1至3年', '贷款3至5年', '贷款5年以上', '公积金5年以下', '公积金5年以上']

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: legendNames, top: 0 },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: { type: 'value', name: '利率(%)', scale: true },
  series: legendNames.map(name => ({ name, type: 'line', step: 'end', symbolSize: 6, data: [] as number[] }))
}))

const tableData = computed(() => [...props.data].reverse())

function fmt(v: number | null) {
  return v === null || v === undefined ? '-' : v.toFixed(2)
}

watch(
  () => props.data,
  rows => {
    updateOptions(opts => {
      opts.xAxis.data = rows.map(r => r.pubDate.slice(0, 10))
      opts.series[0].data = rows.map(r => r.loan6MonthTo1Year ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.loan1YearTo3Year ?? Number.NaN)
      opts.series[2].data = rows.map(r => r.loan3YearTo5Year ?? Number.NaN)
      opts.series[3].data = rows.map(r => r.loanAbove5Year ?? Number.NaN)
      opts.series[4].data = rows.map(r => r.mortgageRateBelow5Year ?? Number.NaN)
      opts.series[5].data = rows.map(r => r.mortgageRateAbove5Year ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '公告日期', key: 'pubDate', width: 120, fixed: 'left' as const, render: (r: Api.Macro.LoanRate) => r.pubDate.slice(0, 10) },
  { title: '贷款6个月(%)', key: 'loan6Month', width: 130, render: (r: Api.Macro.LoanRate) => fmt(r.loan6Month) },
  { title: '贷款6个月至1年(%)', key: 'loan6MonthTo1Year', width: 170, render: (r: Api.Macro.LoanRate) => fmt(r.loan6MonthTo1Year) },
  { title: '贷款1至3年(%)', key: 'loan1YearTo3Year', width: 150, render: (r: Api.Macro.LoanRate) => fmt(r.loan1YearTo3Year) },
  { title: '贷款3至5年(%)', key: 'loan3YearTo5Year', width: 150, render: (r: Api.Macro.LoanRate) => fmt(r.loan3YearTo5Year) },
  { title: '贷款5年以上(%)', key: 'loanAbove5Year', width: 150, render: (r: Api.Macro.LoanRate) => fmt(r.loanAbove5Year) },
  { title: '公积金5年以下(%)', key: 'mortgageRateBelow5Year', width: 160, render: (r: Api.Macro.LoanRate) => fmt(r.mortgageRateBelow5Year) },
  { title: '公积金5年以上(%)', key: 'mortgageRateAbove5Year', width: 160, render: (r: Api.Macro.LoanRate) => fmt(r.mortgageRateAbove5Year) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="贷款利率走势(%)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="历次调整明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="1200"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>