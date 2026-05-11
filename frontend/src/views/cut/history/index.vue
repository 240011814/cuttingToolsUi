<script setup lang="tsx">
import { h, onMounted, ref, computed } from 'vue';
import { NButton, NCard, NDataTable, NPopconfirm, NTag, useMessage } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { cutList, deleteRecod } from '@/service/api';
import { useAppStore } from '@/store/modules/app';
import { useRouterPush } from '@/hooks/common/router';
import { $t } from '@/locales';
import RecordSearch from './modules/record-search.vue';

const appStore = useAppStore();
const message = useMessage();
const { routerPushByKey } = useRouterPush();

const loading = ref(false);
const data = ref<any[]>([]);
const total = ref(0);
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 15, 20],
  onChange: (page: number) => {
    pagination.value.page = page;
    getData();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.value.pageSize = pageSize;
    pagination.value.page = 1;
    getData();
  }
});

const searchParams = ref({
  name: null as string | null,
  type: null as string | null,
  startTime: null as string | null,
  endTime: null as string | null
});

const columns = computed<DataTableColumns<any>>(() => [
  { key: 'code', title: '编号', align: 'center', minWidth: 100 },
  { key: 'name', title: '名称', align: 'center', minWidth: 100 },
  {
    key: 'type',
    title: '类型',
    align: 'center',
    width: 120,
    render(row) {
      const label = row.type === '1' ? '一维' : '平面';
      return h(NTag, { bordered: false }, { default: () => label });
    }
  },
  { key: 'createTime', title: '创建时间', align: 'center', minWidth: 200 },
  {
    key: 'operate',
    title: $t('common.operate'),
    align: 'center',
    width: 180,
    render(row) {
      return h('div', { class: 'flex gap-2' }, [
        h(
          NButton,
          { size: 'small', type: 'primary', quaternary: true, onClick: () => edit(row) },
          { default: () => $t('common.view') }
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => deleteData(row.id) },
          {
            trigger: () =>
              h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => $t('common.delete') }),
            default: () => $t('common.confirmDelete')
          }
        )
      ]);
    }
  }
]);

async function getData() {
  loading.value = true;
  try {
    const { data: res, error } = await cutList({
      current: pagination.value.page,
      size: pagination.value.pageSize,
      ...searchParams.value
    });
    if (!error && res) {
      data.value = res.content || [];
      total.value = res.page?.totalElements || 0;
      pagination.value.itemCount = total.value;
    }
  } catch (err) {
    message.error('加载失败');
  } finally {
    loading.value = false;
  }
}

function getDataByPage(page: number = 1) {
  if (page !== pagination.value.page) {
    pagination.value.page = page;
  }
  getData();
}

function resetSearchParams() {
  searchParams.value = { name: null, type: null, startTime: null, endTime: null };
  getData();
}

async function deleteData(id: string) {
  const result = await deleteRecod(id);
  if (result) {
    message.success($t('common.deleteSuccess'));
    getData();
  } else {
    message.error('删除失败');
  }
}

function edit(row: Api.Cut.CutRecord) {
  if (row.type === '1') {
    routerPushByKey('cut_bar-detail', {
      params: { id: row.id },
      query: { request: row.request, response: row.response }
    });
  } else {
    routerPushByKey('cut_plane-detail', {
      params: { id: row.id },
      query: { request: row.request, response: row.response }
    });
  }
}

onMounted(() => {
  getData();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <NCard :bordered="false" shadow="sm" class="flex-1">
      <template #header>
        <div class="flex items-center gap-4">
          <span class="text-18px font-bold">历史记录</span>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <RecordSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />

        <NDataTable
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="pagination"
          remote
          :row-key="(row) => row.id"
          flex-height
          class="flex-1"
        />
      </div>
    </NCard>
  </div>
</template>

<style scoped></style>
