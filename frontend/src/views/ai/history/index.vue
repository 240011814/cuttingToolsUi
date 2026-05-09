<script setup lang="ts">
import { h, onMounted, ref, computed, resolveComponent } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage, NTag, NButton, NDrawer, NDrawerContent, NAvatar, NSelect, NPopconfirm } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchHistoryList, fetchUpdateFavorite, fetchDeleteHistory } from '@/service/api';
import { $t } from '@/locales';

import MarkdownIt from 'markdown-it';

const router = useRouter();
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
const favoriteFilter = ref(null);
const favoriteOptions = computed(() => [
  { label: $t('page.ai.history.all'), value: null },
  { label: $t('page.ai.history.favorited'), value: true },
  { label: $t('page.ai.history.unfavorited'), value: false }
]);
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
    message.error($t('page.ai.history.parseFailed'));
  }
};

const trainingTypeRouteMap: Record<string, string> = {
  ai_chat: '/ai/chat',
  ai_decision: '/ai/decision',
  ai_social: '/ai/social',
  ai_emergency: '/ai/emergency'
};

const handleContinue = (row: any) => {
  const route = trainingTypeRouteMap[row.training_type];
  if (!route) {
    message.error($t('page.ai.history.unknownType'));
    return;
  }
  
  router.push({
    path: route,
    query: {
      history_id: row.id
    }
  });
};

const handleToggleFavorite = async (row: any) => {
  try {
    await fetchUpdateFavorite(row.id, !row.is_favorite);
    row.is_favorite = !row.is_favorite;
    message.success(row.is_favorite ? $t('page.ai.history.favoritedSuccess') : $t('page.ai.history.unfavoritedSuccess'));
  } catch (err: any) {
    message.error(`${$t('page.ai.history.operationFailed')}: ${err?.message || $t('common.error')}`);
  }
};

const handleDelete = async (row: any) => {
  try {
    await fetchDeleteHistory(row.id);
    message.success($t('page.ai.history.deleteSuccess'));
    loadData();
  } catch (err: any) {
    message.error(`${$t('page.ai.history.deleteFailed')}: ${err?.message || $t('common.error')}`);
  }
};

const columns = computed<DataTableColumns<any>>(() => {
  return [
    {
      title: $t('page.ai.history.trainingProject'),
      key: 'training_type',
      width: 120,
      render(row) {
        const typeMap: Record<string, string> = {
          ai_chat: $t('route.ai_chat'),
          ai_decision: $t('route.ai_decision'),
          ai_social: $t('route.ai_social'),
          ai_emergency: $t('route.ai_emergency')
        };
        return h(NTag, { type: 'info', bordered: false }, { default: () => typeMap[row.training_type] || row.training_type });
      }
    },
    { 
      title: $t('page.ai.history.titleContent'), 
      key: 'title', 
      minWidth: 150
    },
    {
      title: $t('page.ai.history.recentChat'),
      key: 'last_message',
      minWidth: 200,
      render(row) {
        try {
          const msgs = JSON.parse(row.messages || '[]');
          const lastMsg = msgs[msgs.length - 1];
          if (!lastMsg) return h('span', { class: 'text-gray-400' }, $t('page.ai.history.noChatRecord'));
          
          let content = lastMsg.content;
          content = content.replace(/<[^>]*>?/gm, '').replace(/[#*`]/g, '').trim();
          if (content.length > 50) content = content.slice(0, 50) + '...';
          
          return h('div', { class: 'text-xs text-gray-500' }, [
            h(NTag, { size: 'small', quaternary: true, type: lastMsg.role === 'user' ? 'success' : 'info', class: 'mr-1' }, { default: () => (lastMsg.role === 'user' ? 'U' : 'AI') }),
            h('span', content)
          ]);
        } catch (e) {
          return h('span', { class: 'text-error' }, $t('page.ai.history.parseFailed'));
        }
      }
    },
    {
      title: $t('page.ai.history.favorite'),
      key: 'is_favorite',
      width: 80,
      align: 'center',
      render(row) {
        const SvgIcon = resolveComponent('SvgIcon');
        return h('div', { class: 'flex justify-center cursor-pointer', onClick: () => handleToggleFavorite(row) }, [
          h(SvgIcon, {
            icon: row.is_favorite ? 'mdi:star' : 'mdi:star-outline',
            class: row.is_favorite ? 'text-yellow-500 text-xl' : 'text-gray-400 text-xl'
          })
        ]);
      }
    },
    {
      title: $t('page.ai.history.trainingTime'),
      key: 'created_at',
      width: 160,
      render(row) {
        return h('span', { class: 'text-sm' }, new Date(row.created_at).toLocaleString());
      }
    },
    {
      title: $t('page.ai.history.actions'),
      key: 'actions',
      width: 200,
      fixed: 'right',
      render(row) {
        return h('div', { class: 'flex gap-2' }, [
          h(NButton, { size: 'small', type: 'primary', quaternary: true, onClick: () => handleView(row) }, { default: () => $t('page.ai.history.view') }),
          h(NButton, { size: 'small', type: 'success', quaternary: true, onClick: () => handleContinue(row) }, { default: () => $t('page.ai.history.continueTraining') }),
          h(NPopconfirm, { onPositiveClick: () => handleDelete(row) }, {
            trigger: () => h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => $t('common.delete') }),
            default: () => $t('page.ai.history.deleteConfirm')
          })
        ]);
      }
    }
  ];
});

const loadData = async () => {
  loading.value = true;
  try {
    const { data: res } = await fetchHistoryList({
      title: title.value,
      is_favorite: favoriteFilter.value !== null ? favoriteFilter.value : undefined,
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    });
    if (res) {
      data.value = res.items;
      total.value = res.total;
      pagination.value.itemCount = res.total;
    }
  } catch (err: any) {
    message.error(`${$t('page.ai.history.loadFailed')}: ${err?.message || $t('common.error')}`);
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
          <span class="text-18px font-bold">{{ $t('page.ai.history.title') }}</span>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <div class="flex justify-between items-center">
          <div class="flex gap-4 items-center">
            <NInput
              v-model:value="title"
              :placeholder="$t('page.ai.history.searchPlaceholder')"
              clearable
              style="width: 240px"
              @keyup.enter="loadData"
            />
            <NSelect
              v-model:value="favoriteFilter"
              :placeholder="$t('page.ai.history.favoriteStatus')"
              clearable
              :options="favoriteOptions"
              style="width: 120px"
              @update:value="loadData"
            />
            <NButton type="primary" @click="loadData">
              <template #icon>
                <icon-mdi-magnify class="text-icon" />
              </template>
              {{ $t('common.search') }}
            </NButton>
          </div>
          <div class="flex gap-2 items-center">
            <NButton quaternary @click="loadData">
              <template #icon>
                <SvgIcon icon="mdi:refresh" class="text-icon" />
              </template>
            </NButton>
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
      <NDrawerContent :title="$t('page.ai.history.chatDetail')">
        <div class="flex flex-col gap-4">
          <div
            v-for="(msg, index) in currentMessages"
            :key="index"
            class="flex flex-col gap-2"
          >
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
