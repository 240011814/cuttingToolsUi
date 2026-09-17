<script setup lang="ts">
import { computed, watch } from 'vue'
import { NCard, NDataTable } from 'naive-ui'
import { useEcharts } from '@/hooks/common/echarts'

defineOptions({ name: 'MacroDepositRate' })

const props = defineProps<{
  data: Api.Macro.DepositRate[]
  loading: boolean
}>()

// 图表只画关键期限, 全部期限见明细表
const legendNames = ['活期', '定期1年', '定期3年', '定期5年']

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
      opts.series[0].data = rows.map(r => r.demand ?? Number.NaN)
      opts.series[1].data = rows.map(r => r.fixed1Year ?? Number.NaN)
      opts.series[2].data = rows.map(r => r.fixed3Year ?? Number.NaN)
      opts.series[3].data = rows.map(r => r.fixed5Year ?? Number.NaN)
      return opts
    })
  },
  { immediate: true }
)

const columns = [
  { title: '公告日期', key: 'pubDate', width: 120, fixed: 'left' as const, render: (r: Api.Macro.DepositRate) => r.pubDate.slice(0, 10) },
  { title: '活期(%)', key: 'demand', width: 90, render: (r: Api.Macro.DepositRate) => fmt(r.demand) },
  { title: '整存整取3个月(%)', key: 'fixed3Month', width: 160, render: (r: Api.Macro.DepositRate) => fmt(r.fixed3Month) },
  { title: '整存整取6个月(%)', key: 'fixed6Month', width: 160, render: (r: Api.Macro.DepositRate) => fmt(r.fixed6Month) },
  { title: '整存整取1年(%)', key: 'fixed1Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.fixed1Year) },
  { title: '整存整取2年(%)', key: 'fixed2Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.fixed2Year) },
  { title: '整存整取3年(%)', key: 'fixed3Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.fixed3Year) },
  { title: '整存整取5年(%)', key: 'fixed5Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.fixed5Year) },
  { title: '零存整取1年(%)', key: 'installment1Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.installment1Year) },
  { title: '零存整取3年(%)', key: 'installment3Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.installment3Year) },
  { title: '零存整取5年(%)', key: 'installment5Year', width: 150, render: (r: Api.Macro.DepositRate) => fmt(r.installment5Year) }
]
</script>

<template>
  <div class="flex flex-col gap-4">
    <NCard :bordered="false" size="small" title="存款利率走势(%)">
      <div ref="domRef" class="h-360px overflow-hidden" />
    </NCard>
    <NCard :bordered="false" size="small" title="历次调整明细">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :scroll-x="1520"
        :pagination="{ pageSize: 20 }"
        size="small"
        striped
      />
    </NCard>
  </div>
</template>