<script setup lang="ts">
import { h, onMounted, ref, resolveComponent, computed } from "vue";
import { NButton, NPopconfirm, useMessage, useDialog } from "naive-ui";
import type { DataTableColumns, DataTableRowKey } from "naive-ui";
import { useRouterPush } from "@/hooks/common/router";
import {
  fetchDeleteVocabulary,
  fetchGetVocabularyList,
  fetchUpdateVocabulary,
} from "@/service/api";
import { onKeyStroke } from "@vueuse/core";
import { speak } from "@/utils/tts";
import { $t } from "@/locales";
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
const activeTab = ref<"new" | "mastered">("new");

const columns = computed<DataTableColumns<any>>(() => {
  const cols: DataTableColumns<any> = [
    {
      title: $t("page.ai.vocabulary.word"),
      key: "word",
      width: 150,
      render(row) {
        return h("span", { class: "text-lg font-bold text-primary" }, row.word);
      },
    },
    {
      title: $t("page.ai.vocabulary.phonetic"),
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
            // eslint-disable-next-line @typescript-eslint/no-use-before-define
            onClick: () => handlePlay(row.word),
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
    { title: $t("page.ai.vocabulary.phoneticSymbol"), key: "phonetic", width: 100 },
    { title: $t("page.ai.vocabulary.definition"), key: "definition", minWidth: 200 },
    { title: $t("page.ai.vocabulary.example"), key: "example", minWidth: 200 },
    { title: $t("page.ai.vocabulary.confusing"), key: "confusingWords", minWidth: 150 },
    {
      title: $t("page.ai.vocabulary.addedAt"),
      key: "createdAt",
      width: 180,
      render(row) {
        return h("span", new Date(row.createdAt).toLocaleString());
      },
    },
    {
      title: $t("page.ai.vocabulary.actions"),
      key: "actions",
      width: 150,
      fixed: "right",
      render(row) {
        const actions = [
          h(
            NButton,
            {
              size: "small",
              type: activeTab.value === "new" ? "success" : "warning",
              quaternary: true,
              // eslint-disable-next-line @typescript-eslint/no-use-before-define
              onClick: () => handleToggleMastered(row),
            },
            {
              default: () =>
                activeTab.value === "new"
                  ? $t("page.ai.vocabulary.masteredBtn")
                  : $t("page.ai.vocabulary.moveBackBtn"),
            }
          ),
          h(
            NPopconfirm,
            {
              // eslint-disable-next-line @typescript-eslint/no-use-before-define
              onPositiveClick: () => handleDelete(row.id),
              trigger: "click",
            },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: "small", type: "error", quaternary: true },
                  { default: () => $t("common.delete") }
                ),
              default: () => $t("page.ai.vocabulary.deleteConfirm"),
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
    const { data: res } = await fetchGetVocabularyList({
      keyword: keyword.value,
      isMastered: activeTab.value === "mastered",
    });
    if (res) {
      data.value = res;
    }
  } catch (err: any) {
    message.error(
      `${$t("page.ai.vocabulary.loadFailed")}: ${err?.message || $t("common.error")}`
    );
  } finally {
    loading.value = false;
  }
};

const handleDelete = async (id: number) => {
  try {
    await fetchDeleteVocabulary(id);
    message.success($t("page.ai.vocabulary.deleteSuccess"));
    loadData();
  } catch (err: any) {
    message.error(
      `${$t("page.ai.vocabulary.deleteFailed")}: ${err?.message || $t("common.error")}`
    );
  }
};

const handleToggleMastered = async (row: any) => {
  try {
    const newStatus = !row.isMastered;
    await fetchUpdateVocabulary(row.id, { isMastered: newStatus });
    message.success(
      newStatus
        ? $t("page.ai.vocabulary.movedToMastered")
        : $t("page.ai.vocabulary.movedBack")
    );
    loadData();
  } catch (err: any) {
    message.error(
      `${$t("page.ai.vocabulary.operationFailed")}: ${err?.message || $t("common.error")}`
    );
  }
};

const handleBatchMastered = async () => {
  if (checkedRowKeys.value.length === 0) {
    message.warning($t("page.ai.vocabulary.selectWordFirst"));
    return;
  }
  try {
    const isToMastered = activeTab.value === "new";
    await Promise.all(
      checkedRowKeys.value.map((id) =>
        fetchUpdateVocabulary(id as number, { isMastered: isToMastered })
      )
    );
    message.success(
      isToMastered
        ? $t("page.ai.vocabulary.batchMasteredSuccess")
        : $t("page.ai.vocabulary.batchRestoreSuccess")
    );
    checkedRowKeys.value = [];
    loadData();
  } catch (err: any) {
    message.error(
      `${$t("page.ai.vocabulary.batchFailed")}: ${err?.message || $t("common.error")}`
    );
  }
};

onKeyStroke(["m", "M"], (e) => {
  if (e.ctrlKey) {
    e.preventDefault();
    handleBatchMastered();
  }
});

const handlePlay = (text: string) => {
  speak(text, {
    lang: "en-US",
    rate: 0.9,
  }).catch((err) => {
    message.error($t("page.ai.vocabulary.speechFailed"));
    console.error(err);
  });
};

const handleStartExercise = () => {
  const selectedIds = checkedRowKeys.value;
  if (selectedIds.length === 0) {
    dialog.info({
      title: $t("page.ai.vocabulary.trainAll"),
      content: $t("page.ai.vocabulary.trainAllConfirm"),
      positiveText: $t("common.confirm"),
      negativeText: $t("common.cancel"),
      onPositiveClick: () => {
        routerPushByKey("ai_exercise", { query: { mode: "all" } });
      },
    });
  } else {
    routerPushByKey("ai_exercise", { query: { ids: selectedIds.join(",") } });
  }
};

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4" :class="appStore.isMobile ? 'p-2' : 'p-4'">
    <NCard :bordered="false" :shadow="appStore.isMobile ? false : 'sm'" class="flex-1">
      <template #header>
        <div class="flex items-center gap-2" :class="appStore.isMobile ? 'flex-col items-start' : 'gap-4'">
          <span class="font-bold" :class="appStore.isMobile ? 'text-16px' : 'text-18px'">{{ $t("page.ai.vocabulary.title") }}</span>
          <NTabs
            v-model:value="activeTab"
            type="segment"
            :style="{ width: appStore.isMobile ? '100%' : '280px' }"
            @update:value="loadData"
          >
            <NTab name="new">
              <div class="flex items-center gap-2">
                <IconMdiBookOpenVariant class="text-lg" />
                <span>{{ $t("page.ai.vocabulary.wordBook") }}</span>
              </div>
            </NTab>
            <NTab name="mastered">
              <div class="flex items-center gap-2">
                <IconMdiCheckDecagram class="text-lg text-success" />
                <span>{{ $t("page.ai.vocabulary.mastered") }}</span>
              </div>
            </NTab>
          </NTabs>
        </div>
      </template>
      <div class="flex flex-col h-full gap-4">
        <!-- Search and Actions -->
        <template v-if="appStore.isMobile">
          <div class="flex gap-2">
            <NInput
              v-model:value="keyword"
              :placeholder="$t('page.ai.vocabulary.searchPlaceholder')"
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
                type="info"
                size="small"
                @click="handleStartExercise"
              >
                <template #icon>
                  <IconMdiPlayCircleOutline class="text-icon" />
                </template>
              </NButton>
            </div>
            <ButtonIcon
              icon="mdi:refresh"
              :tooltip-content="$t('common.refresh')"
              @click="loadData"
            />
          </div>
        </template>
        <template v-else>
          <div class="flex justify-between items-center">
            <div class="flex gap-4 items-center">
              <NInput
                v-model:value="keyword"
                :placeholder="$t('page.ai.vocabulary.searchPlaceholder')"
                clearable
                style="width: 260px"
                @keyup.enter="loadData"
              />
              <NButton type="primary" @click="loadData">
                <template #icon>
                  <IconMdiMagnify class="text-icon" />
                </template>
                {{ $t("common.search") }}
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
                {{
                  isSelectionMode
                    ? $t("page.ai.vocabulary.selectMode")
                    : $t("page.ai.vocabulary.selectModeOn")
                }}
              </NButton>
              <NButton type="info" @click="handleStartExercise">
                <template #icon>
                  <IconMdiPlayCircleOutline class="text-icon" />
                </template>
                {{ $t("page.ai.vocabulary.startPractice") }}
              </NButton>
              <ButtonIcon
                icon="mdi:refresh"
                :tooltip-content="$t('common.refresh')"
                @click="loadData"
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
          :pagination="{ pageSize: 10 }"
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
                  <span class="text-lg font-bold text-primary">{{ row.word }}</span>
                  <NButton
                    circle
                    size="tiny"
                    quaternary
                    type="primary"
                    @click="handlePlay(row.word)"
                  >
                    <template #icon>
                      <SvgIcon icon="mdi:volume-high" class="text-18px" />
                    </template>
                  </NButton>
                </div>
                <span class="text-xs text-gray-400">{{ row.phonetic }}</span>
              </div>
              <div class="text-sm text-gray-600 dark:text-gray-400">{{ row.definition }}</div>
              <div v-if="row.example" class="text-xs text-gray-500 italic">{{ row.example }}</div>
              <div v-if="row.confusingWords" class="text-xs text-gray-400">
                {{ $t("page.ai.vocabulary.confusing") }}: {{ row.confusingWords }}
              </div>
              <div class="flex items-center justify-between mt-1">
                <span class="text-xs text-gray-400">{{ new Date(row.createdAt).toLocaleDateString() }}</span>
                <div class="flex gap-1">
                  <NButton
                    size="tiny"
                    :type="activeTab === 'new' ? 'success' : 'warning'"
                    quaternary
                    @click="handleToggleMastered(row)"
                  >
                    {{ activeTab === 'new' ? $t("page.ai.vocabulary.masteredBtn") : $t("page.ai.vocabulary.moveBackBtn") }}
                  </NButton>
                  <NPopconfirm @positive-click="handleDelete(row.id)">
                    <template #trigger>
                      <NButton size="tiny" type="error" quaternary>
                        {{ $t("common.delete") }}
                      </NButton>
                    </template>
                    {{ $t("page.ai.vocabulary.deleteConfirm") }}
                  </NPopconfirm>
                </div>
              </div>
            </div>
          </NCard>
          <NEmpty v-if="data.length === 0" class="py-8" />
        </div>
      </div>
    </NCard>
  </div>
</template>

<style scoped></style>
