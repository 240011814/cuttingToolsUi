<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NButton } from 'naive-ui'
import {
  stockScreen,
  stockIndustries,
  stockConcepts,
  saveFilterCondition,
  getFilterConditions,
  deleteFilterCondition,
  addWatchlist,
  syncStockList,
  syncDailyQuotes,
  syncConcepts
} from '@/service/api'

defineOptions({ name: 'ToolStockscreen' })

const router = useRouter()
const message = useMessage()

const loading = ref(false)
const results = ref<Api.Stock.ScreenResult[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const pagination = reactive({
  page: currentPage,
  pageSize: pageSize,
  itemCount: total,
  pageSizes: [20, 50, 100],
  showSizePicker: true,
  onChange: (page: number) => {
    currentPage.value = page
    doScreen()
  },
  onUpdatePageSize: (size: number) => {
    pageSize.value = size
    currentPage.value = 1
    doScreen()
  }
})

const industryOptions = ref<string[]>([])
const conceptOptions = ref<{ label: string; value: string }[]>([])
const savedFilters = ref<Api.Stock.FilterConditionSave[]>([])

const filterForm = reactive({
  conditions: [] as Api.Stock.FilterCondition[],
  conceptNames: [] as string[],
  industries: [] as string[],
  markets: [] as string[],
  excludeSt: true,
  sortBy: 'code',
  sortOrder: 'asc' as 'asc' | 'desc',
  keyword: ''
})

// 筛选条件模板
const filterTemplates = [
  { label: 'PE(TTM)', field: 'peTtm', type: 'number' },
  { label: 'PB', field: 'pb', type: 'number' },
  { label: 'ROE(%)', field: 'roe', type: 'number' },
  { label: '营收增长率(%)', field: 'revenueYoy', type: 'number' },
  { label: '净利润增长率(%)', field: 'netProfitYoy', type: 'number' },
  { label: '涨跌幅(%)', field: 'changePct', type: 'number' },
  { label: '换手率(%)', field: 'turnoverRate', type: 'number' },
  { label: '总市值(亿)', field: 'marketCap', type: 'number' }
]

const operatorOptions = [
  { label: '>', value: 'gt' },
  { label: '>=', value: 'gte' },
  { label: '<', value: 'lt' },
  { label: '<=', value: 'lte' },
  { label: '=', value: 'eq' },
  { label: '区间', value: 'between' }
]

const marketOptions = [
  { label: '沪市', value: 'SH' },
  { label: '深市', value: 'SZ' },
  { label: '北交所', value: 'BJ' }
]

const columns = [
  { title: '代码', key: 'code', width: 80, fixed: 'left' as const,
    render: (row: Api.Stock.ScreenResult) => h('a', {
      class: 'text-blue-500 cursor-pointer hover:underline',
      onClick: () => goToDetail(row.code)
    }, row.code)
  },
  { title: '名称', key: 'name', width: 100, fixed: 'left' as const },
  { title: '现价', key: 'price', width: 80, render: (row: Api.Stock.ScreenResult) => row.price?.toFixed(2) || '-' },
  { title: '涨跌%', key: 'changePct', width: 80, sorter: true,
    render: (row: Api.Stock.ScreenResult) => {
      if (row.changePct === null) return '-'
      const color = row.changePct >= 0 ? 'text-red-500' : 'text-green-500'
      return h('span', { class: color }, `${row.changePct >= 0 ? '+' : ''}${row.changePct.toFixed(2)}%`)
    }
  },
  { title: 'PE(TTM)', key: 'peTtm', width: 80, sorter: true,
    render: (row: Api.Stock.ScreenResult) => row.peTtm?.toFixed(2) || '-'
  },
  { title: 'PB', key: 'pb', width: 60, sorter: true,
    render: (row: Api.Stock.ScreenResult) => row.pb?.toFixed(2) || '-'
  },
  { title: 'ROE%', key: 'roe', width: 70, sorter: true,
    render: (row: Api.Stock.ScreenResult) => row.roe?.toFixed(2) || '-'
  },
  { title: '换手率%', key: 'turnoverRate', width: 80, sorter: true,
    render: (row: Api.Stock.ScreenResult) => row.turnoverRate?.toFixed(2) || '-'
  },
  { title: '市值(亿)', key: 'marketCap', width: 90, sorter: true,
    render: (row: Api.Stock.ScreenResult) => row.marketCap?.toFixed(2) || '-'
  },
  { title: '营收增长%', key: 'revenueYoy', width: 90, sorter: true,
    render: (row: Api.Stock.ScreenResult) => row.revenueYoy?.toFixed(2) || '-'
  },
  { title: '行业', key: 'industry', width: 80 },
  { title: '操作', key: 'actions', width: 100, fixed: 'right' as const,
    render: (row: Api.Stock.ScreenResult) => h('div', { class: 'flex gap-1' }, [
      h(NButton, { size: 'tiny', text: true, onClick: () => goToDetail(row.code) }, { default: () => '详情' }),
      h(NButton, { size: 'tiny', text: true, onClick: () => handleAddWatchlist(row.code) }, { default: () => '加自选' })
    ])
  }
]

// 动态筛选条件
const dynamicFilters = ref<Array<{
  field: string
  label: string
  operator: string
  value: number | null
  value2: number | null
}>>([])

function addFilter() {
  dynamicFilters.value.push({
    field: 'peTtm',
    label: 'PE(TTM)',
    operator: 'lt',
    value: null,
    value2: null
  })
}

function removeFilter(index: number) {
  dynamicFilters.value.splice(index, 1)
}

function buildConditions(): Api.Stock.FilterCondition[] {
  const conditions: Api.Stock.FilterCondition[] = []
  for (const f of dynamicFilters.value) {
    if (f.operator === 'between') {
      if (f.value !== null && f.value2 !== null) {
        conditions.push({
          field: f.field,
          operator: 'between',
          value: [f.value, f.value2]
        })
      }
    } else if (f.value !== null) {
      conditions.push({
        field: f.field,
        operator: f.operator as any,
        value: f.value
      })
    }
  }
  return conditions
}

async function doScreen() {
  loading.value = true
  try {
    const conditions = buildConditions()
    const { data: res } = await stockScreen({
      conditions,
      conceptNames: filterForm.conceptNames,
      industries: filterForm.industries,
      markets: filterForm.markets,
      excludeSt: filterForm.excludeSt,
      sortBy: filterForm.sortBy,
      sortOrder: filterForm.sortOrder,
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: filterForm.keyword || undefined
    })
    if (res) {
      results.value = res.list
      total.value = res.total
    }
  } catch (e: any) {
    message.error(e.message || '筛选失败')
  } finally {
    loading.value = false
  }
}

function handleSorter(options: any) {
  if (options.columnKey) {
    filterForm.sortBy = options.columnKey
    filterForm.sortOrder = options.order === 'ascend' ? 'asc' : 'desc'
    doScreen()
  }
}

function goToDetail(code: string) {
  router.push({ name: 'tool_stockdetail', query: { code } })
}

async function handleAddWatchlist(code: string) {
  try {
    await addWatchlist({ code })
    message.success('已添加到自选股')
  } catch (e: any) {
    message.error(e.message || '添加失败')
  }
}

async function handleSaveFilter() {
  const conditions = buildConditions()
  if (conditions.length === 0) {
    message.warning('请先添加筛选条件')
    return
  }
  try {
    await saveFilterCondition({
      name: `筛选条件 ${new Date().toLocaleString()}`,
      conditions: JSON.stringify(conditions)
    })
    message.success('保存成功')
    loadSavedFilters()
  } catch (e: any) {
    message.error(e.message || '保存失败')
  }
}

async function loadSavedFilters() {
  try {
    const { data } = await getFilterConditions()
    if (data) {
      savedFilters.value = data
    }
  } catch {
    // ignore
  }
}

function loadFilter(filter: Api.Stock.FilterConditionSave) {
  try {
    const conditions = JSON.parse(filter.conditions) as Api.Stock.FilterCondition[]
    dynamicFilters.value = conditions.map(c => ({
      field: c.field,
      label: filterTemplates.find(t => t.field === c.field)?.label || c.field,
      operator: c.operator,
      value: Array.isArray(c.value) ? c.value[0] as number : c.value as number,
      value2: Array.isArray(c.value) ? c.value[1] as number : null
    }))
    doScreen()
  } catch {
    message.error('加载筛选条件失败')
  }
}

async function handleDeleteFilter(id: number) {
  try {
    await deleteFilterCondition(id)
    message.success('已删除')
    loadSavedFilters()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  }
}

const syncLoading = ref(false)
async function handleSyncAll() {
  syncLoading.value = true
  try {
    await syncStockList()
    message.success('股票列表同步已启动，请稍候刷新数据')
    // 延迟后自动刷新
    setTimeout(() => doScreen(), 3000)
  } catch (e: any) {
    message.error(e.message || '同步失败')
  } finally {
    syncLoading.value = false
  }
}

async function handleSyncQuotes() {
  syncLoading.value = true
  try {
    await syncDailyQuotes()
    message.success('行情数据同步已启动')
    setTimeout(() => doScreen(), 5000)
  } catch (e: any) {
    message.error(e.message || '同步失败')
  } finally {
    syncLoading.value = false
  }
}

async function handleSyncConcepts() {
  syncLoading.value = true
  try {
    await syncConcepts()
    message.success('概念板块同步已启动')
  } catch (e: any) {
    message.error(e.message || '同步失败')
  } finally {
    syncLoading.value = false
  }
}

onMounted(async () => {
  try {
    const [industriesRes, conceptsRes] = await Promise.all([stockIndustries(), stockConcepts()])
    if (industriesRes.data) {
      industryOptions.value = industriesRes.data
    }
    if (conceptsRes.data) {
      conceptOptions.value = conceptsRes.data.map((c: { name: string; code: string }) => ({ label: c.name, value: c.name }))
    }
  } catch {
    // ignore
  }
  loadSavedFilters()
  doScreen()
})
</script>

<template>
  <div class="flex h-full">
    <!-- 左侧筛选面板 -->
    <div class="w-72 border-r border-gray-200 p-4 overflow-y-auto flex-shrink-0">
      <div class="mb-4">
        <h3 class="text-lg font-bold mb-2">筛选条件</h3>
        <div class="space-y-3">
          <!-- 名称/代码搜索 -->
          <div>
            <NInput
              v-model:value="filterForm.keyword"
              placeholder="输入股票代码或名称"
              clearable
              @keyup.enter="doScreen"
            />
          </div>

          <!-- 行业筛选 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">行业</label>
            <NSelect
              v-model:value="filterForm.industries"
              :options="industryOptions.map(i => ({ label: i, value: i }))"
              multiple
              placeholder="选择行业"
              clearable
            />
          </div>

          <!-- 概念筛选 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">概念板块</label>
            <NSelect
              v-model:value="filterForm.conceptNames"
              :options="conceptOptions"
              multiple
              placeholder="选择概念"
              clearable
              filterable
            />
          </div>

          <!-- 市场筛选 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">上市板块</label>
            <NSelect
              v-model:value="filterForm.markets"
              :options="marketOptions"
              multiple
              placeholder="选择板块"
              clearable
            />
          </div>

          <!-- 排除ST -->
          <div class="flex items-center">
            <NCheckbox v-model:checked="filterForm.excludeSt">排除ST</NCheckbox>
          </div>
        </div>
      </div>

      <!-- 动态筛选条件 -->
      <div class="mb-4">
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-sm font-bold">指标筛选</h3>
          <NButton size="small" @click="addFilter">+ 添加</NButton>
        </div>
        <div class="space-y-2">
          <div v-for="(filter, index) in dynamicFilters" :key="index" class="flex items-center gap-1">
            <NSelect
              v-model:value="filter.field"
              :options="filterTemplates.map(t => ({ label: t.label, value: t.field }))"
              size="small"
              class="w-24"
              @update:value="(val: string) => {
                filter.label = filterTemplates.find(t => t.field === val)?.label || val
              }"
            />
            <NSelect
              v-model:value="filter.operator"
              :options="operatorOptions"
              size="small"
              class="w-16"
            />
            <NInputNumber
              v-model:value="filter.value"
              size="small"
              class="w-20"
              placeholder="值"
            />
            <NInputNumber
              v-if="filter.operator === 'between'"
              v-model:value="filter.value2"
              size="small"
              class="w-20"
              placeholder="至"
            />
            <NButton size="small" text @click="removeFilter(index)">
              <template #icon><span class="i-material-icons-close text-red-500" /></template>
            </NButton>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="space-y-2">
        <NButton type="primary" block :loading="loading" @click="doScreen">开始筛选</NButton>
        <NButton block @click="handleSaveFilter">保存筛选条件</NButton>
      </div>

      <!-- 数据同步 -->
      <div class="mt-4 pt-4 border-t border-gray-200">
        <h3 class="text-sm font-bold mb-2">数据同步</h3>
        <div class="space-y-2">
          <NButton size="small" block :loading="syncLoading" @click="handleSyncAll">同步股票列表</NButton>
          <NButton size="small" block :loading="syncLoading" @click="handleSyncQuotes">同步行情数据</NButton>
          <NButton size="small" block :loading="syncLoading" @click="handleSyncConcepts">同步概念板块</NButton>
        </div>
      </div>

      <!-- 已保存的筛选条件 -->
      <div v-if="savedFilters.length > 0" class="mt-4">
        <h3 class="text-sm font-bold mb-2">已保存的条件</h3>
        <div class="space-y-1">
          <div
            v-for="filter in savedFilters"
            :key="filter.id"
            class="flex items-center justify-between p-2 bg-gray-50 rounded cursor-pointer hover:bg-gray-100"
            @click="loadFilter(filter)"
          >
            <span class="text-sm truncate">{{ filter.name }}</span>
            <NButton size="tiny" text @click.stop="handleDeleteFilter(filter.id)">
              <template #icon><span class="i-material-icons-delete text-gray-400" /></template>
            </NButton>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧结果区域 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 结果统计 -->
      <div class="p-3 border-b border-gray-200 flex items-center justify-between">
        <span class="text-sm text-gray-600">
          共筛选出 <strong>{{ total }}</strong> 只股票
        </span>
      </div>

      <!-- 结果表格 -->
      <div class="flex-1 overflow-auto">
        <NDataTable
          :columns="columns"
          :data="results"
          :loading="loading"
          :row-key="(row: Api.Stock.ScreenResult) => row.code"
          :pagination="pagination"
          :scroll-x="1200"
          size="small"
          striped
          remote
          @update:sorter="handleSorter"
        />
      </div>
    </div>
  </div>
</template>
