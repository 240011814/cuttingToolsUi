<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroMoneySupplyMonth' })

const props = defineProps<{
  data: Api.Macro.MoneySupplyMonth[]
  loading: boolean
}>()

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['M2余额', 'M1余额', 'M0余额'], top: 0 },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: [
    { type: 'value', name: 'M1 / M2(亿元)', scale: true },
    { type: 'value', name: 'M0(亿元)', scale: true }
  ],
  series: [
    { name: 'M2余额', type: 'line', smooth: true, data: [] as number[] },
    { name: 'M1余额', type: 'line', smooth: true, data: [] as number[] },
    { name: 'M0余额', type: 'line', smooth: true, yAxisIndex: 1, data: [] as number[] }
  ]
}))

const tableData = computed(() => [...props.data].reverse())

function fmt(v: number | null) {
  return v === null || v === undefined ? '-' : v.toFixed(2)
}

function monthLabel(row: Api.Macro.MoneySupplyMonth) {
  return `${row.statYear}-${String(row.statMonth).padStart(2, '0')}`
}

watch(
  () => props.data,
  rows => {
    updateOptions(opts => {
      opts.xAxis.data = rows.map(monthLabel)
      opts.series[0].data = rows.map(r => r.m2Month ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.m1Month ?? Number.NaN)
      opts.series[2].data = rows.map(r => r.m0Month ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '年月', key: 'statMonth', width: 100, fixed: 'left' as const, render: (r: Api.Macro.MoneySupplyMonth) => monthLabel(r) },
  { title: 'M0(亿元)', key: 'm0Month', width: 120, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m0Month) },
  { title: 'M0同比(%)', key: 'm0Yoy', width: 110, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m0Yoy) },
  { title: 'M0环比(%)', key: 'm0ChainRelative', width: 110, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m0ChainRelative) },
  { title: 'M1(亿元)', key: 'm1Month', width: 130, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m1Month) },
  { title: 'M1同比(%)', key: 'm1Yoy', width: 110, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m1Yoy) },
  { title: 'M1环比(%)', key: 'm1ChainRelative', width: 110, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m1ChainRelative) },
  { title: 'M2(亿元)', key: 'm2Month', width: 130, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m2Month) },
  { title: 'M2同比(%)', key: 'm2Yoy', width: 110, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m2Yoy) },
  { title: 'M2环比(%)', key: 'm2ChainRelative', width: 110, render: (r: Api.Macro.MoneySupplyMonth) => fmt(r.m2ChainRelative) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="货币供应量走势(亿元)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="月度明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="1200"
        :pagination="{ pageSize: 24 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>