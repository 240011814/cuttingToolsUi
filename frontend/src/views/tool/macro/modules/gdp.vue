<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroGDP' })

const props = defineProps<{
  data: Api.Macro.GDP[]
  loading: boolean
}>()

const legendNames = ['GDP同比增长(%)', 'GDP累计值(亿元)']

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: legendNames, top: 0 },
  grid: { left: '3%', right: '8%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: [
    { type: 'value', name: '增长率(%)', scale: true },
    { type: 'value', name: '累计值(亿元)', scale: true }
  ],
  series: [
    { name: legendNames[0], type: 'line', symbolSize: 6, data: [] as number[], yAxisIndex: 0 },
    { name: legendNames[1], type: 'bar', data: [] as number[], yAxisIndex: 1 }
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
      opts.xAxis.data = rows.map(r => r.quarter)
      opts.series[0].data = rows.map(r => r.gdpYoy ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.gdpCumulative ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '季度', key: 'quarter', width: 120, fixed: 'left' as const },
  { title: 'GDP同比增长(%)', key: 'gdpYoy', width: 140, render: (r: Api.Macro.GDP) => fmt(r.gdpYoy) },
  { title: 'GDP累计值(亿元)', key: 'gdpCumulative', width: 150, render: (r: Api.Macro.GDP) => r.gdpCumulative ? r.gdpCumulative.toFixed(0) : '-' }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="GDP走势">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="GDP数据明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="410"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>
