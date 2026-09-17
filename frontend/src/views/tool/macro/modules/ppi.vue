<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroPPI' })

const props = defineProps<{
  data: Api.Macro.PPI[]
  loading: boolean
}>()

const legendNames = ['PPI当月同比增长(%)', 'PPI累计指数']

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: legendNames, top: 0 },
  grid: { left: '3%', right: '8%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: [
    { type: 'value', name: '同比增长(%)', scale: true },
    { type: 'value', name: '累计指数', scale: true }
  ],
  series: [
    { name: legendNames[0], type: 'line', symbolSize: 6, data: [] as number[], yAxisIndex: 0 },
    { name: legendNames[1], type: 'line', symbolSize: 6, data: [] as number[], yAxisIndex: 1, lineStyle: { type: 'dashed' } }
  ]
}))

const tableData = computed(() => [...props.data].reverse())

function fmt(v: number | null) {
  return v === null || v === undefined ? '-' : v.toFixed(2)
}

watch(
  () => props.data,
  rows => {
    updateOptions(opts => {
      opts.xAxis.data = rows.map(r => r.month)
      opts.series[0].data = rows.map(r => r.ppiYoy ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.ppiCumulativeYoy ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '月份', key: 'month', width: 120, fixed: 'left' as const },
  { title: 'PPI当月同比增长(%)', key: 'ppiYoy', width: 160, render: (r: Api.Macro.PPI) => fmt(r.ppiYoy) },
  { title: 'PPI累计指数', key: 'ppiCumulativeYoy', width: 140, render: (r: Api.Macro.PPI) => fmt(r.ppiCumulativeYoy) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="PPI走势">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="PPI数据明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="420"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>
