<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroLPR' })

const props = defineProps<{
  data: Api.Macro.LPR[]
  loading: boolean
}>()

const legendNames = ['LPR1Y', 'LPR5Y', '基准利率(6M-1Y)', '基准利率(5Y+)']

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
      opts.xAxis.data = rows.map(r => r.tradeDate.slice(0, 10))
      opts.series[0].data = rows.map(r => r.lpr1Year ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.lpr5Year ?? Number.NaN)
      opts.series[2].data = rows.map(r => r.rate1 ?? Number.NaN)
      opts.series[3].data = rows.map(r => r.rate2 ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '日期', key: 'tradeDate', width: 120, fixed: 'left' as const, render: (r: Api.Macro.LPR) => r.tradeDate.slice(0, 10) },
  { title: 'LPR1Y(%)', key: 'lpr1Year', width: 110, render: (r: Api.Macro.LPR) => fmt(r.lpr1Year) },
  { title: 'LPR5Y(%)', key: 'lpr5Year', width: 110, render: (r: Api.Macro.LPR) => fmt(r.lpr5Year) },
  { title: '基准利率6M-1Y(%)', key: 'rate1', width: 170, render: (r: Api.Macro.LPR) => fmt(r.rate1) },
  { title: '基准利率5Y+(%)', key: 'rate2', width: 160, render: (r: Api.Macro.LPR) => fmt(r.rate2) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="LPR与基准贷款利率走势(%)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="历次调整明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="670"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>
