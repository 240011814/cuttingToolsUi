<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useRouter } from "vue-router";
import { useMessage } from "naive-ui";
import { useAuth } from "@/hooks/business/auth";
import {
  fetchCustomTrainingList,
  fetchCustomTrainingDetail,
  fetchCreateCustomTraining,
  fetchUpdateCustomTraining,
  fetchDeleteCustomTraining,
} from "@/service/api";
import type { CustomTraining } from "@/service/api";
import { $t } from "@/locales";

const router = useRouter();
const message = useMessage();
const { hasAuth } = useAuth();

interface TrainingModule {
  key: string;
  title: string;
  description: string;
  icon: string;
  color: string;
  route: string;
  permission: string;
}

const modules = computed<TrainingModule[]>(() => [
  {
    key: "chat",
    title: $t("route.ai_chat"),
    description: "通过模拟真实生活场景练习地道英语口语表达",
    icon: "mdi:translate-variant",
    color: "#2080f0",
    route: "/ai/chat",
    permission: "ai:chat:view",
  },
  {
    key: "decision",
    title: $t("route.ai_decision"),
    description: "学习60+决策模型，提升在工作生活中的决策能力",
    icon: "mdi:scale-balance",
    color: "#8a6d3b",
    route: "/ai/decision",
    permission: "ai:decision:view",
  },
  {
    key: "social",
    title: $t("route.ai_social"),
    description: "练习聊天破冰、安慰、拒绝等40+沟通场景",
    icon: "mdi:account-group-outline",
    color: "#7c3aed",
    route: "/ai/social",
    permission: "ai:social:view",
  },
  {
    key: "emergency",
    title: $t("route.ai_emergency"),
    description: "突发应变与反应力训练，掌握应变策略",
    icon: "mdi:incognito",
    color: "#d9534f",
    route: "/ai/emergency",
    permission: "ai:emergency:view",
  },
]);

const customTrainings = ref<CustomTraining[]>([]);
const loading = ref(false);

const showModal = ref(false);
const isEdit = ref(false);
const editId = ref(0);
const submitting = ref(false);

const form = ref({
  title: "",
  description: "",
  system_prompt: "",
  icon: "mdi:robot-outline",
  color: "#2080f0",
  initial_message: "",
  input_placeholder: "",
  speech_lang: "zh-CN",
  speech_rate: 0.95,
});

const iconOptions = computed(() => [
  { label: $t("page.ai.training.icons.robot"), value: "mdi:robot-outline" },
  { label: $t("page.ai.training.icons.chat"), value: "mdi:chat-outline" },
  { label: $t("page.ai.training.icons.brain"), value: "mdi:brain" },
  { label: $t("page.ai.training.icons.lightbulb"), value: "mdi:lightbulb-outline" },
  { label: $t("page.ai.training.icons.book"), value: "mdi:book-open-outline" },
  { label: $t("page.ai.training.icons.pencil"), value: "mdi:pencil-outline" },
  { label: $t("page.ai.training.icons.star"), value: "mdi:star-outline" },
  { label: $t("page.ai.training.icons.heart"), value: "mdi:heart-outline" },
  { label: $t("page.ai.training.icons.rocket"), value: "mdi:rocket-launch-outline" },
  { label: $t("page.ai.training.icons.cog"), value: "mdi:cog-outline" },
]);

const colorOptions = computed(() => [
  { label: $t("page.ai.training.colors.blue"), value: "#2080f0" },
  { label: $t("page.ai.training.colors.green"), value: "#18a058" },
  { label: $t("page.ai.training.colors.red"), value: "#d9534f" },
  { label: $t("page.ai.training.colors.orange"), value: "#f0a020" },
  { label: $t("page.ai.training.colors.purple"), value: "#7c3aed" },
  { label: $t("page.ai.training.colors.brown"), value: "#8a6d3b" },
  { label: $t("page.ai.training.colors.gray"), value: "#666666" },
  { label: $t("page.ai.training.colors.cyan"), value: "#20c997" },
]);

const resetForm = () => {
  form.value = {
    title: "",
    description: "",
    system_prompt: "",
    icon: "mdi:robot-outline",
    color: "#2080f0",
    initial_message: "",
    input_placeholder: "",
    speech_lang: "zh-CN",
    speech_rate: 0.95,
  };
};

const loadCustomTrainings = async () => {
  loading.value = true;
  try {
    const { data } = await fetchCustomTrainingList();
    if (data) {
      customTrainings.value = data;
    }
  } catch (err: any) {
    console.error("loadCustomTrainings error:", err);
  } finally {
    loading.value = false;
  }
};

function goToModule(route: string) {
  router.push(route);
}

function goToCustomTraining(id: number) {
  router.push(`/ai/custom-training/${id}`);
}

function handleCreate() {
  isEdit.value = false;
  editId.value = 0;
  resetForm();
  showModal.value = true;
}

async function handleEdit(id: number) {
  isEdit.value = true;
  editId.value = id;
  try {
    const { data } = await fetchCustomTrainingDetail(id);
    if (data) {
      form.value = {
        title: data.title,
        description: data.description,
        system_prompt: data.system_prompt,
        icon: data.icon || "mdi:robot-outline",
        color: data.color || "#2080f0",
        initial_message: data.initial_message,
        input_placeholder: data.input_placeholder || "",
        speech_lang: data.speech_lang || "zh-CN",
        speech_rate: data.speech_rate || 0.95,
      };
    }
  } catch (err: any) {
    message.error(
      `${$t("page.ai.training.loadFailed")}: ${err?.message || $t("common.error")}`
    );
    return;
  }
  showModal.value = true;
}

async function handleSubmit() {
  if (!form.value.title.trim()) {
    message.warning($t("page.ai.training.titleRequired"));
    return;
  }
  if (!form.value.system_prompt.trim()) {
    message.warning($t("page.ai.training.promptRequired"));
    return;
  }

  submitting.value = true;
  try {
    if (isEdit.value) {
      await fetchUpdateCustomTraining(editId.value, form.value);
      message.success($t("page.ai.training.updateSuccess"));
    } else {
      await fetchCreateCustomTraining(form.value);
      message.success($t("page.ai.training.createSuccess"));
    }
    showModal.value = false;
    loadCustomTrainings();
  } catch (err: any) {
    message.error(
      `${$t("page.ai.training.operationFailed")}: ${err?.message || $t("common.error")}`
    );
  } finally {
    submitting.value = false;
  }
}

async function handleDeleteCustomTraining(id: number) {
  try {
    await fetchDeleteCustomTraining(id);
    message.success($t("page.ai.training.deleteSuccess"));
    loadCustomTrainings();
  } catch (err: any) {
    message.error(
      `${$t("page.ai.training.deleteFailed")}: ${err?.message || $t("common.error")}`
    );
  }
}

onMounted(() => {
  loadCustomTrainings();
});
</script>

<template>
  <div class="h-full overflow-auto p-6">
    <div class="mx-auto max-w-4xl">
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-800 dark:text-gray-200">
          {{ $t("page.ai.training.title") }}
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          {{ $t("page.ai.training.subtitle") }}
        </p>
      </div>

      <!-- 内置训练模块 -->
      <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mb-8">
        <template v-for="item in modules" :key="item.key">
          <div
            v-if="hasAuth(item.permission)"
            class="group cursor-pointer rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
            @click="goToModule(item.route)"
          >
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg"
                :style="{ backgroundColor: item.color + '15', color: item.color }"
              >
                <SvgIcon :icon="item.icon" class="text-2xl" />
              </div>
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">
                  {{ item.title }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ item.description }}
                </p>
              </div>
              <SvgIcon
                icon="mdi:chevron-right"
                class="text-xl text-gray-400 transition-transform group-hover:translate-x-1"
              />
            </div>
          </div>
        </template>
      </div>

      <!-- 自定义训练 -->
      <div v-if="hasAuth('ai:custom-training:view')">
        <div class="mb-6 flex items-center justify-between">
          <h2 class="text-xl font-bold text-gray-800 dark:text-gray-200">
            {{ $t("page.ai.training.customTraining") }}
          </h2>
          <NButton
            v-if="hasAuth('ai:custom-training:create')"
            type="primary"
            @click="handleCreate"
          >
            <template #icon>
              <SvgIcon icon="mdi:plus" />
            </template>
            {{ $t("page.ai.training.addTraining") }}
          </NButton>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <NSpin size="large" />
        </div>

        <div
          v-else-if="customTrainings.length === 0"
          class="rounded-xl border border-dashed border-gray-300 p-8 text-center dark:border-gray-600"
        >
          <SvgIcon icon="mdi:robot-outline" class="text-4xl text-gray-400 mb-2" />
          <p class="text-gray-500 dark:text-gray-400">
            {{ $t("page.ai.training.noCustomTraining") }}
          </p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">
            {{ $t("page.ai.training.noCustomTrainingTip") }}
          </p>
        </div>

        <div v-else class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div
            v-for="item in customTrainings"
            :key="item.id"
            class="group relative cursor-pointer rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
            @click="goToCustomTraining(item.id)"
          >
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg"
                :style="{
                  backgroundColor: (item.color || '#2080f0') + '15',
                  color: item.color || '#2080f0',
                }"
              >
                <SvgIcon :icon="item.icon || 'mdi:robot-outline'" class="text-2xl" />
              </div>
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">
                  {{ item.title }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                  {{ item.description || $t("page.ai.training.noDescription") }}
                </p>
              </div>
              <SvgIcon
                icon="mdi:chevron-right"
                class="text-xl text-gray-400 transition-transform group-hover:translate-x-1"
              />
            </div>
            <div
              v-if="
                hasAuth('ai:custom-training:edit') || hasAuth('ai:custom-training:delete')
              "
              class="absolute bottom-3 right-3 flex gap-1 opacity-0 transition-opacity group-hover:opacity-100"
              @click.stop
            >
              <NButton
                v-if="hasAuth('ai:custom-training:edit')"
                size="small"
                quaternary
                @click="handleEdit(item.id)"
              >
                <template #icon>
                  <SvgIcon icon="mdi:pencil-outline" />
                </template>
              </NButton>
              <NPopconfirm
                v-if="hasAuth('ai:custom-training:delete')"
                @positive-click="handleDeleteCustomTraining(item.id)"
              >
                <template #trigger>
                  <NButton size="small" quaternary type="error">
                    <template #icon>
                      <SvgIcon icon="mdi:delete-outline" />
                    </template>
                  </NButton>
                </template>
                {{ $t("page.ai.training.deleteConfirm") }}
              </NPopconfirm>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="
        isEdit
          ? $t('page.ai.training.editTraining')
          : $t('page.ai.training.createTraining')
      "
      style="width: 600px; max-height: 80vh"
      :segmented="{ content: 'soft', footer: 'soft' }"
    >
      <NForm :model="form" label-placement="left" label-width="80">
        <NFormItem :label="$t('page.ai.training.titleLabel')" path="title">
          <NInput
            v-model:value="form.title"
            :placeholder="$t('page.ai.training.titlePlaceholder')"
            maxlength="50"
            show-count
          />
        </NFormItem>
        <NFormItem :label="$t('page.ai.training.descLabel')" path="description">
          <NInput
            v-model:value="form.description"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 3 }"
            :placeholder="$t('page.ai.training.descPlaceholder')"
            maxlength="200"
            show-count
          />
        </NFormItem>
        <NFormItem :label="$t('page.ai.training.promptLabel')" path="system_prompt">
          <NInput
            v-model:value="form.system_prompt"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 8 }"
            :placeholder="$t('page.ai.training.promptPlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.ai.training.welcomeLabel')" path="initial_message">
          <NInput
            v-model:value="form.initial_message"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 3 }"
            :placeholder="$t('page.ai.training.welcomePlaceholder')"
          />
        </NFormItem>
        <div class="grid grid-cols-2 gap-4">
          <NFormItem :label="$t('page.ai.training.iconLabel')" path="icon">
            <NSelect v-model:value="form.icon" :options="iconOptions" />
          </NFormItem>
          <NFormItem :label="$t('page.ai.training.colorLabel')" path="color">
            <NSelect v-model:value="form.color" :options="colorOptions" />
          </NFormItem>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <NFormItem :label="$t('page.ai.training.langLabel')" path="speech_lang">
            <NSelect
              v-model:value="form.speech_lang"
              :options="[
                { label: $t('page.ai.training.chinese'), value: 'zh-CN' },
                { label: $t('page.ai.training.english'), value: 'en-US' },
                { label: $t('page.ai.training.japanese'), value: 'ja-JP' },
              ]"
            />
          </NFormItem>
          <NFormItem :label="$t('page.ai.training.speedLabel')" path="speech_rate">
            <NSlider
              v-model:value="form.speech_rate"
              :min="0.5"
              :max="1.5"
              :step="0.05"
            />
          </NFormItem>
        </div>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showModal = false">{{ $t("common.cancel") }}</NButton>
          <NButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? $t("common.save") : $t("common.confirm") }}
          </NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>
