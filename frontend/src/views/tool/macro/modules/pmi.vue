<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroPMI' })

const props = defineProps<{
  data: Api.Macro.PMI[]
  loading: boolean
}>()

const legendNames = ['制造业PMI', '非制造业PMI']

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: legendNames, top: 0 },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: { type: 'value', name: 'PMI指数', scale: true },
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
      opts.series[0].data = rows.map(r => r.pmiManufacturing ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.pmiNonManufacturing ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '月份', key: 'month', width: 120, fixed: 'left' as const },
  { title: '制造业PMI', key: 'pmiManufacturing', width: 120, render: (r: Api.Macro.PMI) => fmt(r.pmiManufacturing) },
  { title: '制造业PMI同比', key: 'pmiManufacturingYoy', width: 130, render: (r: Api.Macro.PMI) => fmt(r.pmiManufacturingYoy) },
  { title: '非制造业PMI', key: 'pmiNonManufacturing', width: 130, render: (r: Api.Macro.PMI) => fmt(r.pmiNonManufacturing) },
  { title: '非制造业PMI同比', key: 'pmiNonManufacturingYoy', width: 140, render: (r: Api.Macro.PMI) => fmt(r.pmiNonManufacturingYoy) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="PMI指数走势">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="PMI数据明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="640"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>
