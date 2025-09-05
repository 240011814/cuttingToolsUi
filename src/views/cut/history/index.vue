<script setup lang="tsx">
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { cutList, deleteRecod } from '@/service/api';
import { useAppStore } from '@/store/modules/app';
import { useTable, useTableOperate } from '@/hooks/common/table';
import { $t } from '@/locales';
import RecordSearch from './modules/record-search.vue';

const appStore = useAppStore();

const {
  columns,
  columnChecks,
  data,
  getData,
  getDataByPage,
  loading,
  mobilePagination,
  searchParams,
  resetSearchParams
} = useTable({
  apiFn: cutList,
  showTotal: true,
  immediate: true,
  apiParams: {
    current: 1,
    size: 10,
    name: null,
    type: null,
    startTime: null,
    endTime: null
  },
  columns: () => [
    {
      key: 'code',
      title: '编号',
      align: 'center',
      minWidth: 100
    },
    {
      key: 'name',
      title: '名称',
      align: 'center',
      minWidth: 100
    },

    {
      key: 'type',
      title: '类型',
      align: 'center',
      width: 120,
      render: row => {
        const label = row.type === '1' ? '一维' : '平面';

        return (
          <NTag>
            {{
              default: () => label
            }}
          </NTag>
        );
      }
    },
    {
      key: 'createTime',
      title: '创建时间',
      align: 'center',
      minWidth: 200
    },

    {
      key: 'operate',
      title: $t('common.operate'),
      align: 'center',
      width: 180,
      render: row => (
        <div class="flex-center gap-8px">
          <NButton
            type="primary"
            ghost
            size="small"
            onClick={() => {
              edit(row.id);
            }}
          >
            {$t('common.view')}
          </NButton>
          <NPopconfirm onPositiveClick={() => deleteData(row.id)}>
            {{
              default: () => $t('common.confirmDelete'),
              trigger: () => (
                <NButton type="error" ghost size="small">
                  {$t('common.delete')}
                </NButton>
              )
            }}
          </NPopconfirm>
        </div>
      )
    }
  ]
});

const {
  handleEdit,
  checkedRowKeys,
  onDeleted
  // closeDrawer
} = useTableOperate(data, getData);

async function deleteData(id: string) {
  const result = await deleteRecod(id);
  if (result) {
    window.$message?.success('删除成功');
    onDeleted();
  } else {
    window.$message?.error('删除失败');
  }
}

function edit(id: string) {
  handleEdit(id, item => {
    if (item?.enabled) {
      item.status = '1';
    } else {
      item.status = '2';
    }
  });
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <RecordSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <NCard title="历史记录" :bordered="false" size="small" class="card-wrapper sm:flex-1-hidden">
      <template #header-extra>
        <TableHeaderOperation
          v-model:columns="columnChecks"
          :disabled-delete="checkedRowKeys.length === 0"
          :loading="loading"
          :enable-add="false"
          :enable-delete="false"
          @refresh="getData"
        />
      </template>
      <NDataTable
        v-model:checked-row-keys="checkedRowKeys"
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="962"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>
  </div>
</template>

<style scoped></style>
