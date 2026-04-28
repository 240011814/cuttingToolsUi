<script setup lang="ts">
import { h, onMounted, ref, resolveComponent } from 'vue';
import { NButton, NIcon, NPopconfirm, useMessage } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchDeleteVocabulary, fetchGetVocabularyList } from '@/service/api';

const message = useMessage();
const loading = ref(false);
const data = ref<any[]>([]);
const keyword = ref('');

const columns: DataTableColumns<any> = [
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
          secondary: true,
          type: 'primary',
          onClick: () => handlePlay(row.word)
        },
        {
          icon: () => h(NIcon, { size: 20 }, { default: () => h('div', { class: 'i-mdi:volume-high' }) })
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


onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <n-card title="生词本管理" :bordered="false" shadow="sm" class="flex-1">
      <div class="flex flex-col h-full gap-4">
        <div class="flex justify-between items-center">
          <div class="flex gap-4 items-center">
            <n-input
              v-model:value="keyword"
              placeholder="搜索单词..."
              clearable
              @keyup.enter="loadData"
              style="width: 260px"
            />
            <n-button type="primary" @click="loadData">
              <template #icon>
                <div class="i-mdi:search" />
              </template>
              搜索
            </n-button>
          </div>
          <n-button @click="loadData" circle quaternary>
            <template #icon>
              <div class="i-mdi:refresh" />
            </template>
          </n-button>
        </div>

        <n-data-table
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="{ pageSize: 10 }"
          flex-height
          class="flex-1"
        />
      </div>
    </n-card>
  </div>
</template>

<style scoped></style>
