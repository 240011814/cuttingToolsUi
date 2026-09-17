<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroMoneySupplyYear' })

const props = defineProps<{
  data: Api.Macro.MoneySupplyYear[]
  loading: boolean
}>()

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['M2年末余额', 'M1年末余额', 'M0年末余额'], top: 0 },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: [
    { type: 'value', name: 'M1 / M2(亿元)', scale: true },
    { type: 'value', name: 'M0(亿元)', scale: true }
  ],
  series: [
    { name: 'M2年末余额', type: 'line', smooth: true, data: [] as number[] },
    { name: 'M1年末余额', type: 'line', smooth: true, data: [] as number[] },
    { name: 'M0年末余额', type: 'line', smooth: true, yAxisIndex: 1, data: [] as number[] }
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
      opts.xAxis.data = rows.map(r => String(r.statYear))
      opts.series[0].data = rows.map(r => r.m2Year ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.m1Year ?? Number.NaN)
      opts.series[2].data = rows.map(r => r.m0Year ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '年度', key: 'statYear', width: 90, fixed: 'left' as const },
  { title: 'M0年末(亿元)', key: 'm0Year', width: 130, render: (r: Api.Macro.MoneySupplyYear) => fmt(r.m0Year) },
  { title: 'M0同比(%)', key: 'm0YearYoy', width: 110, render: (r: Api.Macro.MoneySupplyYear) => fmt(r.m0YearYoy) },
  { title: 'M1年末(亿元)', key: 'm1Year', width: 140, render: (r: Api.Macro.MoneySupplyYear) => fmt(r.m1Year) },
  { title: 'M1同比(%)', key: 'm1YearYoy', width: 110, render: (r: Api.Macro.MoneySupplyYear) => fmt(r.m1YearYoy) },
  { title: 'M2年末(亿元)', key: 'm2Year', width: 140, render: (r: Api.Macro.MoneySupplyYear) => fmt(r.m2Year) },
  { title: 'M2同比(%)', key: 'm2YearYoy', width: 110, render: (r: Api.Macro.MoneySupplyYear) => fmt(r.m2YearYoy) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="货币供应量年底余额走势(亿元)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="年度明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="860"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>