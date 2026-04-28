<script setup lang="ts">
import { h, onMounted, ref, resolveComponent, computed } from 'vue';
import { NButton, NPopconfirm, useMessage, useDialog } from 'naive-ui';
import type { DataTableColumns, DataTableRowKey } from 'naive-ui';
import { useRouterPush } from '@/hooks/common/router';
import { fetchDeleteVocabulary, fetchGetVocabularyList } from '@/service/api';

const message = useMessage();
const dialog = useDialog();
const { routerPushByKey } = useRouterPush();
const loading = ref(false);
const data = ref<any[]>([]);
const keyword = ref('');
const checkedRowKeys = ref<DataTableRowKey[]>([]);
const isSelectionMode = ref(false);

const columns = computed<DataTableColumns<any>>(() => {
  const cols: DataTableColumns<any> = [
    { title: '单词', key: 'word', width: 120 },
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
      width: 100,
      fixed: 'right',
      render(row) {
        return h(
          NPopconfirm,
          {
            onPositiveClick: () => handleDelete(row.id),
            trigger: 'click'
          },
          {
            trigger: () => h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => '删除' }),
            default: () => '确定删除此单词吗？'
          }
        );
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
    const { data: res } = await fetchGetVocabularyList({ keyword: keyword.value });
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
    <NCard title="生词本管理" :bordered="false" shadow="sm" class="flex-1">
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
