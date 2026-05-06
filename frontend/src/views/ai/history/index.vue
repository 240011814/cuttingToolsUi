<script setup lang="ts">
import { h, onMounted, ref, computed } from 'vue';
import { useMessage, NTag, NButton, NDrawer, NDrawerContent, NAvatar, NSelect } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchHistoryList } from '@/service/api';

import MarkdownIt from 'markdown-it';

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
});

const renderMarkdown = (content: string) => {
  return md.render(content || '');
};

const message = useMessage();
const loading = ref(false);
const data = ref<any[]>([]);
const title = ref('');
const recordType = ref(null);
const recordTypeOptions = [
  { label: '全部方式', value: null },
  { label: '自动', value: 'auto' },
  { label: '手动', value: 'manual' }
];
const total = ref(0);
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  onChange: (page: number) => {
    pagination.value.page = page;
    loadData();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.value.pageSize = pageSize;
    pagination.value.page = 1;
    loadData();
  }
});

const showDrawer = ref(false);
const currentMessages = ref<any[]>([]);

const handleView = (row: any) => {
  try {
    currentMessages.value = JSON.parse(row.messages || '[]');
    showDrawer.value = true;
  } catch (e) {
    message.error('无法解析该记录的对话数据');
  }
};

const columns = computed<DataTableColumns<any>>(() => {
  return [
    {
      title: '训练项目',
      key: 'training_type',
      width: 150,
      render(row) {
        return h(NTag, { type: 'info', bordered: false }, { default: () => row.training_type });
      }
    },
    { 
      title: '标题/内容', 
      key: 'title', 
      minWidth: 150
    },
    {
      title: '最近对话',
      key: 'last_message',
      minWidth: 250,
      render(row) {
        try {
          const msgs = JSON.parse(row.messages || '[]');
          const lastMsg = msgs[msgs.length - 1];
          if (!lastMsg) return h('span', { class: 'text-gray-400' }, '无对话记录');
          
          let content = lastMsg.content;
          // 简单过滤 Markdown 标签
          content = content.replace(/<[^>]*>?/gm, '').replace(/[#*`]/g, '').trim();
          if (content.length > 50) content = content.slice(0, 50) + '...';
          
          return h('div', { class: 'text-xs text-gray-500' }, [
            h(NTag, { size: 'small', quaternary: true, type: lastMsg.role === 'user' ? 'success' : 'info', class: 'mr-1' }, { default: () => (lastMsg.role === 'user' ? 'U' : 'AI') }),
            h('span', content)
          ]);
        } catch (e) {
          return h('span', { class: 'text-error' }, '解析失败');
        }
      }
    },
    {
      title: '记录方式',
      key: 'record_type',
      width: 120,
      render(row) {
        const isAuto = row.record_type === 'auto';
        return h(NTag, { type: isAuto ? 'success' : 'warning', bordered: false }, { default: () => (isAuto ? '自动' : '手动') });
      }
    },
    {
      title: '训练时间',
      key: 'created_at',
      width: 180,
      render(row) {
        return h('span', new Date(row.created_at).toLocaleString());
      }
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render(row) {
        return h(NButton, { size: 'small', type: 'primary', quaternary: true, onClick: () => handleView(row) }, { default: () => '查看对话' });
      }
    }
  ];
});

const loadData = async () => {
  loading.value = true;
  try {
    const { data: res } = await fetchHistoryList({
      title: title.value,
      record_type: recordType.value || undefined,
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    });
    if (res) {
      data.value = res.items;
      total.value = res.total;
      pagination.value.itemCount = res.total;
    }
  } catch (err: any) {
    message.error(`获取列表失败: ${err?.message || '未知错误'}`);
  } finally {
    loading.value = false;
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
          <span class="text-18px font-bold">历史记录</span>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <div class="flex justify-between items-center">
          <div class="flex gap-4 items-center">
            <NInput
              v-model:value="title"
              placeholder="搜索对话标题/内容..."
              clearable
              style="width: 240px"
              @keyup.enter="loadData"
            />
            <NSelect
              v-model:value="recordType"
              placeholder="记录方式"
              clearable
              :options="recordTypeOptions"
              style="width: 150px"
              @update:value="loadData"
            />
            <NButton type="primary" @click="loadData">
              <template #icon>
                <icon-mdi-magnify class="text-icon" />
              </template>
              搜索
            </NButton>
          </div>
          <div class="flex gap-2 items-center">
            <ButtonIcon
              icon="mdi:refresh"
              tooltip-content="刷新"
              @click="loadData"
            />
          </div>
        </div>

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

    <NDrawer v-model:show="showDrawer" width="600" placement="right">
      <NDrawerContent title="对话详情">
        <div class="flex flex-col gap-4">
          <div
            v-for="(msg, index) in currentMessages"
            :key="index"
            class="flex flex-col gap-2"
          >
            <!-- 过滤掉系统提示词，通常是第一条或 role 为 system 的 -->
            <template v-if="msg.role !== 'system'">
              <div
                class="flex items-start gap-3"
                :class="msg.role === 'user' ? 'flex-row-reverse' : 'flex-row'"
              >
                <NAvatar
                  :color="msg.role === 'user' ? '#18a058' : '#2080f0'"
                  round
                  size="small"
                >
                  {{ msg.role === 'user' ? 'U' : 'AI' }}
                </NAvatar>
                <div class="flex flex-col gap-1 max-w-[80%]">
                  <div
                    class="p-3 rounded-2xl whitespace-pre-wrap leading-relaxed shadow-sm text-sm relative"
                    :class="
                      msg.role === 'user'
                        ? 'bg-[#18a058] text-white rounded-tr-none'
                        : 'bg-gray-100 text-gray-800 rounded-tl-none dark:bg-gray-800 dark:text-gray-200'
                    "
                  >
                    <div v-html="renderMarkdown(msg.content)"></div>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped></style>
