<script setup lang="ts">
import { h, onMounted, ref, computed, resolveComponent } from "vue";
import { NButton, NPopconfirm, useMessage, useDialog } from "naive-ui";
import type { DataTableColumns, DataTableRowKey } from "naive-ui";
import { useRouterPush } from "@/hooks/common/router";
import {
  fetchGetErrorBookList,
  fetchUpdateErrorBook,
  fetchDeleteErrorBook,
  fetchGetErrorBookStats,
} from "@/service/api";
import { speak } from "@/utils/tts";
import { useAppStore } from "@/store/modules/app";

const appStore = useAppStore();

const message = useMessage();
const dialog = useDialog();
const { routerPushByKey } = useRouterPush();
const loading = ref(false);
const data = ref<any[]>([]);
const keyword = ref("");
const checkedRowKeys = ref<DataTableRowKey[]>([]);
const isSelectionMode = ref(false);
const activeTab = ref<"unmastered" | "mastered">("unmastered");
const activeType = ref<"" | "word" | "sentence">("");

const stats = ref({
  total: 0,
  mastered: 0,
  unmastered: 0,
  wordCount: 0,
  sentenceCount: 0,
});

const columns = computed<DataTableColumns<any>>(() => {
  const cols: DataTableColumns<any> = [
    {
      title: "内容",
      key: "content",
      minWidth: 200,
      render(row) {
        return h("span", { class: "text-lg font-bold text-primary" }, row.content);
      },
    },
    {
      title: "翻译",
      key: "translation",
      minWidth: 150,
    },
    {
      title: "来源",
      key: "sourceType",
      width: 100,
      render(row) {
        return h(
          "span",
          {
            class:
              row.sourceType === "vocabulary"
                ? "text-blue-500 font-medium"
                : "text-green-500 font-medium",
          },
          row.sourceType === "vocabulary" ? "生词本" : "课程包"
        );
      },
    },
    {
      title: "播放",
      key: "play",
      width: 80,
      render(row) {
        return h(
          NButton,
          {
            circle: true,
            size: "small",
            quaternary: true,
            type: "primary",
            onClick: () => handlePlay(row.content),
          },
          {
            icon: () =>
              h(resolveComponent("SvgIcon"), {
                icon: "mdi:volume-high",
                class: "text-22px text-primary",
              }),
          }
        );
      },
    },
    {
      title: "错误次数",
      key: "errorCount",
      width: 100,
      render(row) {
        return h(
          "span",
          { class: "text-red-500 font-bold" },
          `${row.errorCount} 次`
        );
      },
    },
    {
      title: "添加时间",
      key: "createdAt",
      width: 180,
      render(row) {
        return h("span", new Date(row.createdAt).toLocaleString());
      },
    },
    {
      title: "操作",
      key: "actions",
      width: 150,
      fixed: "right",
      render(row) {
        const actions = [
          h(
            NButton,
            {
              size: "small",
              type: activeTab.value === "unmastered" ? "success" : "warning",
              quaternary: true,
              onClick: () => handleToggleMastered(row),
            },
            {
              default: () =>
                activeTab.value === "unmastered" ? "已掌握" : "移回错题",
            }
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDelete(row.id),
              trigger: "click",
            },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: "small", type: "error", quaternary: true },
                  { default: () => "删除" }
                ),
              default: () => "确定删除这条错题记录吗？",
            }
          ),
        ];
        return h("div", { class: "flex gap-1" }, actions);
      },
    },
  ];

  if (isSelectionMode.value) {
    cols.unshift({ type: "selection" });
  }

  return cols;
});

const loadData = async () => {
  loading.value = true;
  try {
    const { data: res } = await fetchGetErrorBookList({
      sourceType: activeType.value || undefined,
      keyword: keyword.value || undefined,
      isMastered: activeTab.value === "mastered",
    });
    if (res) {
      data.value = res;
    }
  } catch (err: any) {
    message.error(`加载失败: ${err?.message || "未知错误"}`);
  } finally {
    loading.value = false;
  }
};

const loadStats = async () => {
  try {
    const { data: res } = await fetchGetErrorBookStats();
    if (res) {
      stats.value = res;
    }
  } catch (err: any) {
    console.error("加载统计失败", err);
  }
};

const handleDelete = async (id: number) => {
  try {
    await fetchDeleteErrorBook(id);
    message.success("删除成功");
    loadData();
    loadStats();
  } catch (err: any) {
    message.error(`删除失败: ${err?.message || "未知错误"}`);
  }
};

const handleToggleMastered = async (row: any) => {
  try {
    const newStatus = !row.isMastered;
    await fetchUpdateErrorBook(row.id, { isMastered: newStatus });
    message.success(newStatus ? "已标记为掌握" : "已移回错题本");
    loadData();
    loadStats();
  } catch (err: any) {
    message.error(`操作失败: ${err?.message || "未知错误"}`);
  }
};

const handleBatchMastered = async () => {
  if (checkedRowKeys.value.length === 0) {
    message.warning("请先选择要操作的错题");
    return;
  }
  try {
    const isToMastered = activeTab.value === "unmastered";
    await Promise.all(
      checkedRowKeys.value.map((id) =>
        fetchUpdateErrorBook(id as number, { isMastered: isToMastered })
      )
    );
    message.success(isToMastered ? "批量标记为掌握成功" : "批量移回错题成功");
    checkedRowKeys.value = [];
    loadData();
    loadStats();
  } catch (err: any) {
    message.error(`批量操作失败: ${err?.message || "未知错误"}`);
  }
};

const handlePlay = (text: string) => {
  speak(text, {
    lang: "en-US",
    rate: 0.9,
  }).catch((err) => {
    message.error("语音播放失败");
    console.error(err);
  });
};

const handleStartPractice = () => {
  dialog.info({
    title: "错题练习",
    content: "将开始练习所有未掌握的错题，确定开始吗？",
    positiveText: "确定",
    negativeText: "取消",
    onPositiveClick: () => {
      routerPushByKey("ai_exercise", {
        query: { mode: "error-book", ...(activeType.value ? { type: activeType.value } : {}) },
      });
    },
  });
};

onMounted(() => {
  loadData();
  loadStats();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4" :class="appStore.isMobile ? 'p-2' : 'p-4'">
    <!-- Stats Cards -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
      <NCard size="small" :bordered="false" class="shadow-sm">
        <div class="text-center">
          <div class="text-2xl font-bold text-primary">{{ stats.total }}</div>
          <div class="text-xs text-gray-400">总错题</div>
        </div>
      </NCard>
      <NCard size="small" :bordered="false" class="shadow-sm">
        <div class="text-center">
          <div class="text-2xl font-bold text-red-500">{{ stats.unmastered }}</div>
          <div class="text-xs text-gray-400">待掌握</div>
        </div>
      </NCard>
      <NCard size="small" :bordered="false" class="shadow-sm">
        <div class="text-center">
          <div class="text-2xl font-bold text-green-500">{{ stats.mastered }}</div>
          <div class="text-xs text-gray-400">已掌握</div>
        </div>
      </NCard>
      <NCard size="small" :bordered="false" class="shadow-sm">
        <div class="text-center">
          <div class="text-2xl font-bold text-blue-500">{{ stats.wordCount }}</div>
          <div class="text-xs text-gray-400">生词本</div>
        </div>
      </NCard>
      <NCard size="small" :bordered="false" class="shadow-sm">
        <div class="text-center">
          <div class="text-2xl font-bold text-purple-500">{{ stats.sentenceCount }}</div>
          <div class="text-xs text-gray-400">课程包</div>
        </div>
      </NCard>
    </div>

    <NCard :bordered="false" :shadow="appStore.isMobile ? false : 'sm'" class="flex-1">
      <template #header>
        <div class="flex items-center gap-2" :class="appStore.isMobile ? 'flex-col items-start' : 'gap-4'">
          <span class="font-bold" :class="appStore.isMobile ? 'text-16px' : 'text-18px'">错题本</span>
          <NTabs
            v-model:value="activeTab"
            type="segment"
            :style="{ width: appStore.isMobile ? '100%' : '280px' }"
            @update:value="loadData"
          >
            <NTab name="unmastered">
              <div class="flex items-center gap-2">
                <IconMdiAlertCircle class="text-lg text-red-500" />
                <span>待掌握</span>
              </div>
            </NTab>
            <NTab name="mastered">
              <div class="flex items-center gap-2">
                <IconMdiCheckDecagram class="text-lg text-green-500" />
                <span>已掌握</span>
              </div>
            </NTab>
          </NTabs>
          <NSelect
            v-model:value="activeType"
            :options="[
              { label: '全部来源', value: '' },
              { label: '生词本', value: 'vocabulary' },
              { label: '课程包', value: 'course' },
            ]"
            placeholder="筛选来源"
            clearable
            style="width: 120px"
            @update:value="loadData"
          />
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <!-- Search and Actions -->
        <template v-if="appStore.isMobile">
          <div class="flex gap-2">
            <NInput
              v-model:value="keyword"
              placeholder="搜索错题内容..."
              clearable
              class="flex-1"
              @keyup.enter="loadData"
            />
            <NButton type="primary" @click="loadData">
              <template #icon>
                <IconMdiMagnify class="text-icon" />
              </template>
            </NButton>
          </div>
          <div class="flex gap-2 justify-between">
            <div class="flex gap-2">
              <NButton
                size="small"
                :type="isSelectionMode ? 'primary' : 'default'"
                @click="isSelectionMode = !isSelectionMode"
              >
                <template #icon>
                  <IconMdiCheckboxMultipleMarkedOutline class="text-icon" />
                </template>
              </NButton>
              <NButton
                type="error"
                size="small"
                @click="handleStartPractice"
              >
                <template #icon>
                  <IconMdiPlayCircleOutline class="text-icon" />
                </template>
                错题练习
              </NButton>
            </div>
            <ButtonIcon
              icon="mdi:refresh"
              tooltip-content="刷新"
              @click="loadData(); loadStats()"
            />
          </div>
        </template>
        <template v-else>
          <div class="flex justify-between items-center">
            <div class="flex gap-4 items-center">
              <NInput
                v-model:value="keyword"
                placeholder="搜索错题内容..."
                clearable
                style="width: 260px"
                @keyup.enter="loadData"
              />
              <NButton type="primary" @click="loadData">
                <template #icon>
                  <IconMdiMagnify class="text-icon" />
                </template>
                搜索
              </NButton>
            </div>
            <div class="flex gap-2 items-center">
              <NButton
                :type="isSelectionMode ? 'primary' : 'default'"
                @click="isSelectionMode = !isSelectionMode"
              >
                <template #icon>
                  <IconMdiCheckboxMultipleMarkedOutline class="text-icon" />
                </template>
                {{ isSelectionMode ? "选择模式" : "批量选择" }}
              </NButton>
              <NButton
                v-if="isSelectionMode"
                type="success"
                @click="handleBatchMastered"
              >
                <template #icon>
                  <IconMdiCheckAll class="text-icon" />
                </template>
                批量标记掌握
              </NButton>
              <NButton type="error" @click="handleStartPractice">
                <template #icon>
                  <IconMdiPlayCircleOutline class="text-icon" />
                </template>
                错题练习
              </NButton>
              <ButtonIcon
                icon="mdi:refresh"
                tooltip-content="刷新"
                @click="loadData(); loadStats()"
              />
            </div>
          </div>
        </template>

        <!-- PC: DataTable -->
        <NDataTable
          v-if="!appStore.isMobile"
          v-model:checked-row-keys="checkedRowKeys"
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="{ pageSize: 15 }"
          :row-key="(row) => row.id"
          flex-height
          class="flex-1"
        />

        <!-- Mobile: Card List -->
        <NSpin v-if="appStore.isMobile && loading" class="flex justify-center py-8" />
        <div v-else-if="appStore.isMobile" class="flex flex-col gap-3">
          <NCard
            v-for="row in data"
            :key="row.id"
            size="small"
            :bordered="true"
          >
            <div class="flex flex-col gap-2">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <NTag
                    :type="row.contentType === 'word' ? 'info' : 'success'"
                    size="small"
                  >
                    {{ row.contentType === "word" ? "单词" : "句子" }}
                  </NTag>
                  <span class="text-lg font-bold text-primary">{{ row.content }}</span>
                  <NButton
                    circle
                    size="tiny"
                    quaternary
                    type="primary"
                    @click="handlePlay(row.content)"
                  >
                    <template #icon>
                      <SvgIcon icon="mdi:volume-high" class="text-18px" />
                    </template>
                  </NButton>
                </div>
                <NTag type="error" size="small">{{ row.errorCount }} 次</NTag>
              </div>
              <div v-if="row.translation" class="text-sm text-gray-600 dark:text-gray-400">
                {{ row.translation }}
              </div>
              <div class="flex items-center justify-between mt-1">
                <span class="text-xs text-gray-400">
                  {{ new Date(row.createdAt).toLocaleDateString() }}
                </span>
                <div class="flex gap-1">
                  <NButton
                    size="tiny"
                    :type="activeTab === 'unmastered' ? 'success' : 'warning'"
                    quaternary
                    @click="handleToggleMastered(row)"
                  >
                    {{ activeTab === "unmastered" ? "已掌握" : "移回错题" }}
                  </NButton>
                  <NPopconfirm @positive-click="handleDelete(row.id)">
                    <template #trigger>
                      <NButton size="tiny" type="error" quaternary>
                        删除
                      </NButton>
                    </template>
                    确定删除这条错题记录吗？
                  </NPopconfirm>
                </div>
              </div>
            </div>
          </NCard>
          <NEmpty v-if="data.length === 0" class="py-8" description="暂无错题记录" />
        </div>
      </div>
    </NCard>
  </div>
</template>

<style scoped></style>
