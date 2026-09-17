<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroCPI' })

const props = defineProps<{
  data: Api.Macro.CPI[]
  loading: boolean
}>()

const legendNames = ['CPI同比增长(%)', 'CPI环比增长(%)']

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: legendNames, top: 0 },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: { type: 'value', name: '增长率(%)', scale: true },
  series: legendNames.map(name => ({ name, type: 'line', symbolSize: 6, data: [] as number[] }))
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
      opts.series[0].data = rows.map(r => r.cpiYoy ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.cpiMom ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '月份', key: 'month', width: 120, fixed: 'left' as const },
  { title: 'CPI同比增长(%)', key: 'cpiYoy', width: 140, render: (r: Api.Macro.CPI) => fmt(r.cpiYoy) },
  { title: 'CPI环比增长(%)', key: 'cpiMom', width: 140, render: (r: Api.Macro.CPI) => fmt(r.cpiMom) },
  { title: 'CPI累计同比增长(%)', key: 'cpiCumulativeYoy', width: 170, render: (r: Api.Macro.CPI) => fmt(r.cpiCumulativeYoy) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="CPI增长率走势(%)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="CPI数据明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="570"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>
