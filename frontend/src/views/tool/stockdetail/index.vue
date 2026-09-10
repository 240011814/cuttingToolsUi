<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { stockDetail, stockKline, stockFinanceHistory, addWatchlist, syncSingleStock, fetchSyncStatus } from '@/service/api'
import { useMessage, NButton, NDataTable, NTag, NSpin, NTabs, NTabPane } from 'naive-ui'
import * as echarts from 'echarts'

defineOptions({ name: 'ToolStockdetail' })

const route = useRoute()
const router = useRouter()
const message = useMessage()

const code = computed(() => {
  const c = route.query.code
  if (!c) return ''
  return Array.isArray(c) ? c[0] || '' : c
})
const detail = ref<Api.Stock.ScreenResult | null>(null)
const klineData = ref<Api.Stock.KlineData[]>([])
const financeHistory = ref<Api.Stock.FinanceHistory[]>([])
const loading = ref(false)
const activeTab = ref('local')
const financeViewMode = ref('chart')
const klineViewMode = ref('chart')

const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

const klineChartRef = ref<HTMLElement | null>(null)
let klineChartInstance: echarts.ECharts | null = null

const realtimeUrl = computed(() => {
  if (!code.value) return ''
  const market = code.value.startsWith('6') ? 'sh' : 'sz'
  return `https://quote.eastmoney.com/${market}${code.value}.html`
})

async function loadDetail() {
  if (!code.value) {
    message.error('股票代码不能为空')
    return
  }
  loading.value = true
  try {
    const [detailRes, klineRes, financeRes] = await Promise.all([
      stockDetail(code.value),
      stockKline(code.value, { period: 'daily', count: 120 }),
      stockFinanceHistory(code.value, { limit: 8 })
    ])
    if (detailRes.data) {
      detail.value = detailRes.data
    }
    if (klineRes.data) {
      klineData.value = klineRes.data
    }
    if (financeRes.data) {
      financeHistory.value = financeRes.data
    }
  } catch (e: any) {
    message.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleAddWatchlist() {
  try {
    await addWatchlist({ code: code.value })
    message.success('已添加到自选股')
  } catch (e: any) {
    message.error(e.message || '添加失败')
  }
}

const syncLoading = ref(false)
const syncRunning = ref(false)
let syncPollTimer: ReturnType<typeof setInterval> | null = null

function stopSyncPolling() {
  if (syncPollTimer) {
    clearInterval(syncPollTimer)
    syncPollTimer = null
  }
  syncRunning.value = false
}

function waitForSyncDone(onDone: () => void) {
  stopSyncPolling()
  syncRunning.value = true
  let runningSeen = false
  let ticks = 0
  syncPollTimer = setInterval(async () => {
    ticks++
    if (ticks > 100) {
      stopSyncPolling()
      onDone()
      return
    }
    try {
      const { data } = await fetchSyncStatus()
      if (!data) return
      if (data.running) {
        runningSeen = true
        return
      }
      if (!runningSeen) {
        runningSeen = true
        return
      }
      stopSyncPolling()
      if (data.lastError) {
        message.error(`同步失败: ${data.lastError}`)
      }
      onDone()
    } catch {
      // ignore
    }
  }, 3000)
}

async function handleSync() {
  if (syncLoading.value || syncRunning.value) return
  syncLoading.value = true
  const { error } = await syncSingleStock(code.value, detail.value?.market || '')
  syncLoading.value = false
  if (error) {
    message.warning(error.message || '同步启动失败，可能有其他同步任务在运行')
    waitForSyncDone(() => loadDetail())
    return
  }
  message.loading('正在同步最新数据，请稍候...')
  waitForSyncDone(() => {
    message.success('同步完成')
    loadDetail()
  })
}

function goBack() {
  router.push({ name: 'tool_stockscreen' })
}

const infoItems = computed(() => {
  if (!detail.value) return []
  const d = detail.value
  return [
    { label: '代码', value: d.code, tip: '' },
    { label: '名称', value: d.name, tip: '' },
    { label: '市场', value: d.market, tip: '' },
    { label: '行业', value: d.industry || '-', tip: '' },
    { label: '现价', value: d.price?.toFixed(2) || '-', tip: '' },
    { label: '涨跌幅', value: d.changePct !== null ? `${d.changePct >= 0 ? '+' : ''}${d.changePct.toFixed(2)}%` : '-', tip: '公式: (当前价-昨收价)/昨收价×100%' },
    { label: '换手率', value: d.turnoverRate?.toFixed(2) ? `${d.turnoverRate.toFixed(2)}%` : '-', tip: '公式: 成交量/流通股本×100%。<3%冷门，3-7%正常，>10%非常活跃' },
    { label: '成交额', value: d.amount ? `${(d.amount / 10000).toFixed(2)}亿` : '-', tip: '' },
    { label: '总市值', value: d.marketCap ? `${d.marketCap.toFixed(2)}亿` : '-', tip: '公式: 股价×总股本。<50亿小盘，50-200亿中盘，>1000亿超大盘' },
    { label: '流通市值', value: d.floatMarketCap ? `${d.floatMarketCap.toFixed(2)}亿` : '-', tip: '' }
  ]
})

const financeItems = computed(() => {
  if (!detail.value) return []
  const d = detail.value
  return [
    { label: 'PE(TTM)', value: d.peTtm?.toFixed(2) || '-', tip: '公式: 股价/最近四个季度每股收益之和。<20低估，20-30合理，>30高估' },
    { label: 'PB', value: d.pb?.toFixed(2) || '-', tip: '公式: 股价/每股净资产。<1破净，1-2低估，>3高估' },
    { label: 'ROE', value: d.roe?.toFixed(2) ? `${d.roe.toFixed(2)}%` : '-', tip: '公式: 净利润/净资产×100%。>15%优秀，10-15%良好，<10%一般' },
    { label: '营收增长', value: d.revenueYoy?.toFixed(2) ? `${d.revenueYoy.toFixed(2)}%` : '-', tip: '公式: (本期营收-去年同期营收)/去年同期营收×100%。>20%高增长' },
    { label: '净利润增长', value: d.netProfitYoy?.toFixed(2) ? `${d.netProfitYoy.toFixed(2)}%` : '-', tip: '公式: (本期净利润-去年同期净利润)/去年同期净利润×100%' },
    { label: '毛利率', value: d.grossMargin?.toFixed(2) ? `${d.grossMargin.toFixed(2)}%` : '-', tip: '公式: (营业收入-营业成本)/营业收入×100%' },
    { label: '净利率', value: d.netMargin?.toFixed(2) ? `${d.netMargin.toFixed(2)}%` : '-', tip: '公式: 净利润/营业收入×100%' },
    { label: '资产负债率', value: d.debtRatio?.toFixed(2) ? `${d.debtRatio.toFixed(2)}%` : '-', tip: '公式: 总负债/总资产×100%。<50%低风险，50-70%正常，>70%高风险' },
    { label: '流动比率', value: d.currentRatio?.toFixed(2) || '-', tip: '公式: 流动资产/流动负债。>2优秀，1-2正常，<1有风险' },
    { label: '速动比率', value: d.quickRatio?.toFixed(2) || '-', tip: '公式: (流动资产-存货)/流动负债。>1优秀，0.5-1正常' }
  ]
})

function getChartOption() {
  const data = [...financeHistory.value].reverse()
  const dates = data.map(d => d.reportDate?.slice(0, 10) || '')
  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' }
    },
    legend: {
      data: ['ROE%', '毛利率%', '净利率%', '资产负债率%'],
      top: 0,
      textStyle: { fontSize: 12 }
    },
    grid: { left: 50, right: 20, top: 40, bottom: 30 },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: { fontSize: 11 }
    },
    yAxis: [
      {
        type: 'value',
        name: '%',
        axisLabel: { fontSize: 11 }
      }
    ],
    series: [
      {
        name: 'ROE%',
        type: 'line',
        data: data.map(d => d.roe),
        smooth: true,
        itemStyle: { color: '#1890ff' }
      },
      {
        name: '毛利率%',
        type: 'line',
        data: data.map(d => d.grossMargin),
        smooth: true,
        itemStyle: { color: '#52c41a' }
      },
      {
        name: '净利率%',
        type: 'line',
        data: data.map(d => d.netMargin),
        smooth: true,
        itemStyle: { color: '#faad14' }
      },
      {
        name: '资产负债率%',
        type: 'line',
        data: data.map(d => d.debtRatio),
        smooth: true,
        itemStyle: { color: '#f5222d' }
      }
    ]
  }
}

function getKlineChartOption() {
  const data = [...klineData.value].slice(-10)
  const dates = data.map(d => d.date || '')
  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['收盘价', '涨跌幅%'],
      top: 0,
      textStyle: { fontSize: 12 }
    },
    grid: { left: 50, right: 50, top: 40, bottom: 30 },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: { fontSize: 11, rotate: 30 }
    },
    yAxis: [
      {
        type: 'value',
        name: '价格',
        position: 'left',
        axisLabel: { fontSize: 11 }
      },
      {
        type: 'value',
        name: '涨跌%',
        position: 'right',
        axisLabel: { fontSize: 11 }
      }
    ],
    series: [
      {
        name: '收盘价',
        type: 'line',
        data: data.map(d => d.close),
        smooth: true,
        itemStyle: { color: '#1890ff' },
        areaStyle: { color: 'rgba(24,144,255,0.1)' }
      },
      {
        name: '涨跌幅%',
        type: 'bar',
        yAxisIndex: 1,
        data: data.map(d => d.changePct),
        itemStyle: {
          color: (params: any) => (params.value >= 0 ? '#f5222d' : '#52c41a')
        }
      }
    ]
  }
}

function renderChart() {
  if (!chartRef.value || financeHistory.value.length === 0) return
  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }
  chartInstance.setOption(getChartOption())
}

function renderKlineChart() {
  if (!klineChartRef.value || klineData.value.length === 0) return
  if (!klineChartInstance) {
    klineChartInstance = echarts.init(klineChartRef.value)
  }
  klineChartInstance.setOption(getKlineChartOption())
}

watch(financeViewMode, (val) => {
  if (val === 'chart') {
    nextTick(() => renderChart())
  }
})

watch(klineViewMode, (val) => {
  if (val === 'chart') {
    nextTick(() => renderKlineChart())
  }
})

watch(financeHistory, () => {
  if (financeViewMode.value === 'chart') {
    nextTick(() => renderChart())
  }
})

watch(klineData, () => {
  if (klineViewMode.value === 'chart') {
    nextTick(() => renderKlineChart())
  }
})

onMounted(() => {
  loadDetail()

  fetchSyncStatus()
    .then(({ data }) => {
      if (data?.running) {
        message.info('检测到同步任务正在运行，完成后将自动刷新')
        waitForSyncDone(() => loadDetail())
      }
    })
    .catch(() => {
      // ignore
    })
})

onUnmounted(() => {
  stopSyncPolling()
})
</script>

<template>
  <div class="p-4">
    <!-- 顶部导航 -->
    <div class="flex items-center justify-between mb-4">
      <NButton @click="goBack">
        <template #icon><span class="i-mdi:arrow-left" /></template>
        返回筛选
      </NButton>
      <div class="flex gap-2">
        <NButton type="primary" :loading="syncLoading || syncRunning" :disabled="syncRunning" @click="handleSync">
          <template #icon><span class="i-mdi:refresh" /></template>
          {{ syncRunning ? '同步中...' : '同步最新' }}
        </NButton>
        <NButton type="primary" @click="handleAddWatchlist">加自选</NButton>
      </div>
    </div>

    <NSpin :show="loading">
      <template v-if="detail">
        <!-- 股票头部信息 -->
        <div class="mb-4 p-4 bg-white rounded-lg shadow">
          <div class="flex items-center gap-4">
            <div>
              <h1 class="text-2xl font-bold">{{ detail.name }}</h1>
              <p class="text-gray-500">{{ detail.code }} | {{ detail.market }}</p>
            </div>
            <div v-if="detail.price" class="ml-auto text-right">
              <p class="text-3xl font-bold" :class="detail.changePct && detail.changePct >= 0 ? 'text-red-500' : 'text-green-500'">
                {{ detail.price.toFixed(2) }}
              </p>
              <p v-if="detail.changePct !== null" class="text-lg" :class="detail.changePct >= 0 ? 'text-red-500' : 'text-green-500'">
                {{ detail.changePct >= 0 ? '+' : '' }}{{ detail.changePct.toFixed(2) }}%
              </p>
            </div>
          </div>
        </div>

        <NTabs v-model:value="activeTab" type="line" animated>
          <!-- 本地数据Tab -->
          <NTabPane name="local" tab="本地数据">
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <!-- 基本信息 -->
              <div class="p-4 bg-white rounded-lg shadow">
                <h2 class="text-lg font-bold mb-3">基本信息</h2>
                <div class="grid grid-cols-2 gap-2">
                  <div v-for="item in infoItems" :key="item.label" class="flex justify-between items-center">
                    <span class="text-gray-500">{{ item.label }}</span>
                    <span class="font-medium flex items-center gap-1">
                      {{ item.value }}
                      <span v-if="item.tip" class="text-gray-400 cursor-help" :title="item.tip">?</span>
                    </span>
                  </div>
                </div>
              </div>

              <!-- 财务指标 -->
              <div class="p-4 bg-white rounded-lg shadow">
                <h2 class="text-lg font-bold mb-3">财务指标</h2>
                <div class="grid grid-cols-2 gap-2">
                  <div v-for="item in financeItems" :key="item.label" class="flex justify-between items-center">
                    <span class="text-gray-500">{{ item.label }}</span>
                    <span class="font-medium flex items-center gap-1">
                      {{ item.value }}
                      <span v-if="item.tip" class="text-gray-400 cursor-help" :title="item.tip">?</span>
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 概念板块 -->
            <div v-if="detail.concepts && detail.concepts.length > 0" class="mt-4 p-4 bg-white rounded-lg shadow">
              <h2 class="text-lg font-bold mb-3">概念板块</h2>
              <div class="flex flex-wrap gap-2">
                <NTag v-for="concept in detail.concepts" :key="concept" type="info" size="small">
                  {{ concept }}
                </NTag>
              </div>
            </div>

            <!-- K线数据预览 -->
            <div v-if="klineData.length > 0" class="mt-4 p-4 bg-white rounded-lg shadow">
              <div class="flex items-center justify-between mb-3">
                <h2 class="text-lg font-bold">近期行情 (最近10日)</h2>
                <NTabs v-model:value="klineViewMode" type="segment" size="small" style="width: 160px">
                  <NTabPane name="chart" tab="图表" />
                  <NTabPane name="table" tab="表格" />
                </NTabs>
              </div>

              <!-- 图表模式 -->
              <div v-show="klineViewMode === 'chart'" ref="klineChartRef" style="width: 100%; height: 320px" />

              <!-- 表格模式 -->
              <NDataTable
                v-show="klineViewMode === 'table'"
                :columns="[
                  { title: '日期', key: 'date', width: 100 },
                  { title: '开盘', key: 'open', width: 80, render: (row: Api.Stock.KlineData) => row.open?.toFixed(2) || '-' },
                  { title: '最高', key: 'high', width: 80, render: (row: Api.Stock.KlineData) => row.high?.toFixed(2) || '-' },
                  { title: '最低', key: 'low', width: 80, render: (row: Api.Stock.KlineData) => row.low?.toFixed(2) || '-' },
                  { title: '收盘', key: 'close', width: 80, render: (row: Api.Stock.KlineData) => row.close?.toFixed(2) || '-' },
                  { title: '涨跌%', key: 'changePct', width: 80, render: (row: Api.Stock.KlineData) => row.changePct !== null ? `${row.changePct >= 0 ? '+' : ''}${row.changePct.toFixed(2)}%` : '-' },
                  { title: '成交量', key: 'volume', width: 100, render: (row: Api.Stock.KlineData) => row.volume ? `${(row.volume / 10000).toFixed(2)}万手` : '-' }
                ]"
                :data="klineData.slice(-10)"
                :bordered="false"
                size="small"
                striped
              />
            </div>

            <!-- 历史财务数据 -->
            <div v-if="financeHistory.length > 0" class="mt-4 p-4 bg-white rounded-lg shadow">
              <div class="flex items-center justify-between mb-3">
                <h2 class="text-lg font-bold">历史财务数据</h2>
                <NTabs v-model:value="financeViewMode" type="segment" size="small" style="width: 160px">
                  <NTabPane name="chart" tab="图表" />
                  <NTabPane name="table" tab="表格" />
                </NTabs>
              </div>

              <!-- 图表模式 -->
              <div v-show="financeViewMode === 'chart'" ref="chartRef" style="width: 100%; height: 320px" />

              <!-- 表格模式 -->
              <NDataTable
                v-show="financeViewMode === 'table'"
                :columns="[
                  { title: '报告期', key: 'reportDate', width: 100, render: (row: Api.Stock.FinanceHistory) => row.reportDate?.slice(0, 10) || '-' },
                  { title: 'ROE%', key: 'roe', width: 70, render: (row: Api.Stock.FinanceHistory) => row.roe?.toFixed(2) || '-' },
                  { title: '毛利率%', key: 'grossMargin', width: 70, render: (row: Api.Stock.FinanceHistory) => row.grossMargin?.toFixed(2) || '-' },
                  { title: '净利率%', key: 'netMargin', width: 70, render: (row: Api.Stock.FinanceHistory) => row.netMargin?.toFixed(2) || '-' },
                  { title: '营收(万)', key: 'revenue', width: 80, render: (row: Api.Stock.FinanceHistory) => row.revenue?.toFixed(0) || '-' },
                  { title: '净利润(万)', key: 'netProfit', width: 80, render: (row: Api.Stock.FinanceHistory) => row.netProfit?.toFixed(0) || '-' },
                  { title: 'EPS', key: 'eps', width: 60, render: (row: Api.Stock.FinanceHistory) => row.eps?.toFixed(3) || '-' },
                  { title: '资产负债率%', key: 'debtRatio', width: 80, render: (row: Api.Stock.FinanceHistory) => row.debtRatio?.toFixed(2) || '-' },
                  { title: '流动比率', key: 'currentRatio', width: 70, render: (row: Api.Stock.FinanceHistory) => row.currentRatio?.toFixed(2) || '-' },
                  { title: '速动比率', key: 'quickRatio', width: 70, render: (row: Api.Stock.FinanceHistory) => row.quickRatio?.toFixed(2) || '-' }
                ]"
                :data="financeHistory"
                :bordered="false"
                size="small"
                striped
                :scroll-x="900"
              />
            </div>
          </NTabPane>

          <!-- 实时行情Tab -->
          <NTabPane name="realtime" tab="实时行情">
            <div class="bg-white rounded-lg shadow overflow-hidden" style="height: calc(100vh - 280px)">
              <iframe
                v-if="realtimeUrl"
                :src="realtimeUrl"
                class="w-full h-full border-0"
                loading="lazy"
              />
            </div>
          </NTabPane>
        </NTabs>
      </template>
    </NSpin>
  </div>
</template>
