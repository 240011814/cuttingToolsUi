<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { stockDetail, stockKline, addWatchlist, syncSingleStock } from '@/service/api'
import { useMessage, NButton, NDataTable, NTag, NSpin, NTabs, NTabPane } from 'naive-ui'

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
const loading = ref(false)
const activeTab = ref('local')

// 东方财富行情iframe地址
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
    const [detailRes, klineRes] = await Promise.all([
      stockDetail(code.value),
      stockKline(code.value, { period: 'daily', count: 120 })
    ])
    if (detailRes.data) {
      detail.value = detailRes.data
    }
    if (klineRes.data) {
      klineData.value = klineRes.data
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
async function handleSync() {
  syncLoading.value = true
  try {
    await syncSingleStock(code.value)
    message.success('同步成功')
    loadDetail()
  } catch (e: any) {
    message.error(e.message || '同步失败')
  } finally {
    syncLoading.value = false
  }
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

onMounted(() => {
  loadDetail()
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
        <NButton :loading="syncLoading" @click="handleSync">
          <template #icon><span class="i-mdi:refresh" /></template>
          同步最新
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
            <div class="mt-4 p-4 bg-white rounded-lg shadow">
              <h2 class="text-lg font-bold mb-3">近期行情 (最近10日)</h2>
              <NDataTable
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
