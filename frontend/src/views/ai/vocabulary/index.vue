<script setup lang="ts">
import { h, onMounted, ref, resolveComponent, computed } from 'vue';
import { NButton, NPopconfirm, useMessage, useDialog } from 'naive-ui';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { useRouterPush } from '@/hooks/common/router';
import { fetchDeleteVocabulary, fetchGetVocabularyList, fetchUpdateVocabulary } from '@/service/api';
import { onKeyStroke } from '@vueuse/core';

const message = useMessage();
const dialog = useDialog();
const { routerPushByKey } = useRouterPush();
const loading = ref(false);
const data = ref<any[]>([]);
const keyword = ref('');
const checkedRowKeys = ref<DataTableRowKey[]>([]);
const isSelectionMode = ref(false);
const activeTab = ref<'new' | 'mastered'>('new');

const columns = computed<DataTableColumns<any>>(() => {
  const cols: DataTableColumns<any> = [
    {
      title: '单词',
      key: 'word',
      width: 150,
      render(row) {
        return h('span', { class: 'text-lg font-bold text-primary' }, row.word);
      }
    },
    {
      title: '发音',
      key: 'play',
      width: 80,
      render(row) {
        return h(
          NButton,
          {
            circle: true,
            size: 'small',
            quaternary: true,
            type: 'primary',
            onClick: () => handlePlay(row.word)
          },
          {
            icon: () => h(resolveComponent('SvgIcon'), { icon: 'mdi:volume-high', class: 'text-22px text-primary' })
          }
        );
      }
    },
    { title: '音标', key: 'phonetic', width: 100 },
    { title: '释义', key: 'definition', minWidth: 200 },
    { title: '例句', key: 'example', minWidth: 200 },
    { title: '易混淆', key: 'confusingWords', minWidth: 150 },
    {
      title: '添加时间',
      key: 'createdAt',
      width: 180,
      render(row) {
        return h('span', new Date(row.createdAt).toLocaleString());
      }
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      fixed: 'right',
      render(row) {
        const actions = [
          h(
            NButton,
            {
              size: 'small',
              type: activeTab.value === 'new' ? 'success' : 'warning',
              quaternary: true,
              onClick: () => handleToggleMastered(row)
            },
            { default: () => (activeTab.value === 'new' ? '掌握' : '移回') }
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDelete(row.id),
              trigger: 'click'
            },
            {
              trigger: () => h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => '删除' }),
              default: () => '确定删除此单词吗？'
            }
          )
        ];
        return h('div', { class: 'flex gap-1' }, actions);
      }
    }
  ];

  if (isSelectionMode.value) {
    cols.unshift({ type: 'selection' });
  }

  return cols;
});

const loadData = async () => {
  loading.value = true;
  try {
    const { data: res } = await fetchGetVocabularyList({
      keyword: keyword.value,
      isMastered: activeTab.value === 'mastered'
    });
    if (res) {
      data.value = res;
    }
  } catch (err: any) {
    message.error(`获取列表失败: ${err?.message || '未知错误'}`);
  } finally {
    loading.value = false;
  }
};

const handleDelete = async (id: number) => {
  try {
    await fetchDeleteVocabulary(id);
    message.success('删除成功');
    loadData();
  } catch (err: any) {
    message.error(`删除失败: ${err?.message || '未知错误'}`);
  }
};

const handleToggleMastered = async (row: any) => {
  try {
    const newStatus = !row.isMastered;
    await fetchUpdateVocabulary(row.id, { isMastered: newStatus });
    message.success(newStatus ? '已移至掌握列表' : '已移回生词本');
    loadData();
  } catch (err: any) {
    message.error(`操作失败: ${err?.message || '未知错误'}`);
  }
};

const handleBatchMastered = async () => {
  if (checkedRowKeys.value.length === 0) {
    message.warning('请先选择单词');
    return;
  }
  try {
    const isToMastered = activeTab.value === 'new';
    await Promise.all(checkedRowKeys.value.map(id => fetchUpdateVocabulary(id as number, { isMastered: isToMastered })));
    message.success(isToMastered ? '批量标记成功' : '批量还原成功');
    checkedRowKeys.value = [];
    loadData();
  } catch (err: any) {
    message.error(`批量操作失败: ${err?.message || '未知错误'}`);
  }
};

onKeyStroke(['m', 'M'], (e) => {
  if (e.ctrlKey) {
    e.preventDefault();
    handleBatchMastered();
  }
});

const handlePlay = (text: string) => {
  if (!window.speechSynthesis) {
    message.error('您的浏览器不支持语音播放');
    return;
  }
  // 停止之前的播放
  window.speechSynthesis.cancel();

  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = 'en-US';
  utterance.rate = 0.9; // 稍微放慢一点点，方便听清
  window.speechSynthesis.speak(utterance);
};

const handleStartExercise = () => {
  const selectedIds = checkedRowKeys.value;
  if (selectedIds.length === 0) {
    dialog.info({
      title: '训练全部',
      content: '您未勾选单词，是否训练生词本中的全部单词？',
      positiveText: '确认',
      negativeText: '取消',
      onPositiveClick: () => {
        routerPushByKey('ai_exercise', { query: { mode: 'all' } });
      }
    });
  } else {
    routerPushByKey('ai_exercise', { query: { ids: selectedIds.join(',') } });
  }
};


onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <NCard :bordered="false" shadow="sm" class="flex-1">
      <template #header>
        <div class="flex items-center gap-4">
          <span class="text-18px font-bold">词汇管理</span>
          <NTabs v-model:value="activeTab" type="segment" style="width: 280px" @update:value="loadData">
            <NTab name="new">
              <div class="flex items-center gap-2">
                <icon-mdi-book-open-variant class="text-lg" />
                <span>生词本</span>
              </div>
            </NTab>
            <NTab name="mastered">
              <div class="flex items-center gap-2">
                <icon-mdi-check-decagram class="text-lg text-success" />
                <span>掌握列表</span>
              </div>
            </NTab>
          </NTabs>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <div class="flex justify-between items-center">
          <div class="flex gap-4 items-center">
            <NInput
              v-model:value="keyword"
              placeholder="搜索单词..."
              clearable
              style="width: 260px"
              @keyup.enter="loadData"
            />
            <NButton type="primary" @click="loadData">
              <template #icon>
                <icon-mdi-magnify class="text-icon" />
              </template>
              搜索
            </NButton>
          </div>
          <div class="flex gap-2 items-center">
            <NButton :type="isSelectionMode ? 'primary' : 'default'" @click="isSelectionMode = !isSelectionMode">
              <template #icon>
                <icon-mdi-checkbox-multiple-marked-outline class="text-icon" />
              </template>
              {{ isSelectionMode ? '取消选择' : '开启选择' }}
            </NButton>
            <NButton type="info" @click="handleStartExercise">
              <template #icon>
                <icon-mdi-play-circle-outline class="text-icon" />
              </template>
              开始练习
            </NButton>
            <ButtonIcon
              icon="mdi:refresh"
              tooltip-content="刷新"
              @click="loadData"
            />
          </div>
        </div>

        <NDataTable
          v-model:checked-row-keys="checkedRowKeys"
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="{ pageSize: 10 }"
          :row-key="(row) => row.id"
          flex-height
          class="flex-1"
        />
      </div>
    </NCard>
  </div>
</template>

<style scoped></style>
