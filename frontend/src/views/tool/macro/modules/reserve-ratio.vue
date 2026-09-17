<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroReserveRatio' })

const props = defineProps<{
  data: Api.Macro.ReserveRatio[]
  loading: boolean
}>()

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['大型金融机构', '中小金融机构'], top: 0 },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
  xAxis: { type: 'category', boundaryGap: false, data: [] as string[] },
  yAxis: { type: 'value', name: '准备金率(%)', scale: true },
  series: [
    { name: '大型金融机构', type: 'line', step: 'end', symbolSize: 6, data: [] as number[] },
    { name: '中小金融机构', type: 'line', step: 'end', symbolSize: 6, data: [] as number[] }
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
      opts.xAxis.data = rows.map(r => r.effectiveDate.slice(0, 10))
      opts.series[0].data = rows.map(r => r.bigInstitutionsRatioAfter ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.mediumInstitutionsRatioAfter ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '公告日期', key: 'pubDate', width: 120, render: (row: Api.Macro.ReserveRatio) => row.pubDate.slice(0, 10) },
  {
    title: '生效日期',
    key: 'effectiveDate',
    width: 120,
    render: (row: Api.Macro.ReserveRatio) => row.effectiveDate.slice(0, 10)
  },
  { title: '大型-调整前(%)', key: 'bigInstitutionsRatioPre', width: 140, render: (r: Api.Macro.ReserveRatio) => fmt(r.bigInstitutionsRatioPre) },
  { title: '大型-调整后(%)', key: 'bigInstitutionsRatioAfter', width: 140, render: (r: Api.Macro.ReserveRatio) => fmt(r.bigInstitutionsRatioAfter) },
  { title: '中小-调整前(%)', key: 'mediumInstitutionsRatioPre', width: 140, render: (r: Api.Macro.ReserveRatio) => fmt(r.mediumInstitutionsRatioPre) },
  { title: '中小-调整后(%)', key: 'mediumInstitutionsRatioAfter', width: 140, render: (r: Api.Macro.ReserveRatio) => fmt(r.mediumInstitutionsRatioAfter) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="存款准备金率走势(%)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="历次调整明细">
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