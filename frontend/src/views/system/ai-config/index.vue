<script setup lang="ts">
import { computed, h, onMounted, ref } from "vue";
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NPopconfirm,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import {
  fetchCreateAIModel,
  fetchCreateAIProvider,
  fetchDeleteAIModel,
  fetchDeleteAIProvider,
  fetchGetAIProviders,
  fetchUpdateAIModel,
  fetchUpdateAIProvider,
} from "@/service/api/admin";
import { $t } from "@/locales";

const message = useMessage();
const loading = ref(false);
const providers = ref<Api.Admin.AIProvider[]>([]);

// Provider Modal
const showProviderModal = ref(false);
const providerModalTitle = ref("");
const providerForm = ref<Partial<Api.Admin.AIProvider>>({
  name: "",
  api_key: "",
  base_url: "",
  is_active: false,
});

// Model Modal
const showModelModal = ref(false);
const modelModalTitle = ref("");
const modelModalMode = ref<"list" | "form">("list");
const modelDialogTitle = computed(() =>
  modelModalMode.value === "list"
    ? $t("page.system.aiConfig.manageModel")
    : modelModalTitle.value
);
const currentProviderId = ref<number | null>(null);
const modelForm = ref<Partial<Api.Admin.AIModel>>({
  model_code: "",
  display_name: "",
  is_default: false,
  config_json: "{}",
});

const columns = computed<DataTableColumns<Api.Admin.AIProvider>>(() => [
  { title: $t("page.system.aiConfig.providerName"), key: "name", width: 120 },
  {
    title: "API Key",
    key: "api_key",
    render(row) {
      if (!row.api_key) return $t("page.system.aiConfig.notConfigured");
      return row.api_key.length > 10
        ? `${row.api_key.slice(0, 6)}***${row.api_key.slice(-4)}`
        : "******";
    },
  },
  { title: "Base URL", key: "base_url" },
  {
    title: $t("page.system.aiConfig.status"),
    key: "is_active",
    render(row) {
      return h(NSwitch, {
        value: row.is_active,
        onUpdateValue: (val: boolean) => handleToggleProviderStatus(row, val),
      });
    },
  },
  {
    title: $t("page.system.aiConfig.actions"),
    key: "actions",
    render(row) {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              { size: "small", onClick: () => handleEditProvider(row) },
              { default: () => $t("common.edit") }
            ),
            h(
              NButton,
              {
                size: "small",
                type: "primary",
                ghost: true,
                onClick: () => handleManageModels(row),
              },
              { default: () => $t("page.system.aiConfig.manageModel") }
            ),
            h(
              NPopconfirm,
              { onPositiveClick: () => handleDeleteProvider(row.id) },
              {
                default: () => $t("page.system.aiConfig.deleteProviderConfirm"),
                trigger: () =>
                  h(
                    NButton,
                    { size: "small", type: "error", ghost: true },
                    { default: () => $t("common.delete") }
                  ),
              }
            ),
          ],
        }
      );
    },
  },
]);

async function getProviders() {
  loading.value = true;
  const { data } = await fetchGetAIProviders();
  if (data) {
    providers.value = data;
  }
  loading.value = false;
}

function handleAddProvider() {
  providerModalTitle.value = $t("page.system.aiConfig.addProvider");
  providerForm.value = { name: "", api_key: "", base_url: "", is_active: true };
  showProviderModal.value = true;
}

function handleEditProvider(row: Api.Admin.AIProvider) {
  providerModalTitle.value = $t("page.system.aiConfig.editProvider");
  providerForm.value = { ...row };
  showProviderModal.value = true;
}

async function handleSaveProvider() {
  if (providerForm.value.id) {
    await fetchUpdateAIProvider(providerForm.value.id, providerForm.value);
    message.success($t("page.system.aiConfig.updateSuccess"));
  } else {
    await fetchCreateAIProvider(providerForm.value);
    message.success($t("page.system.aiConfig.createSuccess"));
  }
  showProviderModal.value = false;
  getProviders();
}

async function handleToggleProviderStatus(row: Api.Admin.AIProvider, val: boolean) {
  await fetchUpdateAIProvider(row.id, { ...row, is_active: val });
  message.success(
    val
      ? $t("page.system.aiConfig.enabledStatus")
      : $t("page.system.aiConfig.disabledStatus")
  );
  getProviders();
}

async function handleDeleteProvider(id: number) {
  await fetchDeleteAIProvider(id);
  message.success($t("page.system.aiConfig.deleteSuccess"));
  getProviders();
}

// Model Management
const activeModels = ref<Api.Admin.AIModel[]>([]);
const modelColumns = computed<DataTableColumns<Api.Admin.AIModel>>(() => [
  { title: $t("page.system.aiConfig.modelCode"), key: "model_code" },
  { title: $t("page.system.aiConfig.displayName"), key: "display_name" },
  {
    title: $t("page.system.aiConfig.default"),
    key: "is_default",
    render(row) {
      return row.is_default
        ? h(
            NTag,
            { type: "success" },
            { default: () => $t("page.system.aiConfig.default") }
          )
        : null;
    },
  },
  {
    title: $t("page.system.aiConfig.actions"),
    key: "actions",
    render(row) {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              { size: "small", onClick: () => handleEditModel(row) },
              { default: () => $t("common.edit") }
            ),
            h(
              NPopconfirm,
              { onPositiveClick: () => handleDeleteModel(row) },
              {
                default: () => $t("page.system.aiConfig.deleteModelConfirm"),
                trigger: () =>
                  h(
                    NButton,
                    { size: "small", type: "error", ghost: true },
                    { default: () => $t("common.delete") }
                  ),
              }
            ),
          ],
        }
      );
    },
  },
]);

function handleManageModels(row: Api.Admin.AIProvider) {
  currentProviderId.value = row.id;
  activeModels.value = row.models || [];
  modelModalMode.value = "list";
  showModelModal.value = true;
}

function handleAddModel() {
  modelModalTitle.value = $t("page.system.aiConfig.addModel");
  modelForm.value = {
    provider_id: currentProviderId.value!,
    model_code: "",
    display_name: "",
    is_default: activeModels.value.length === 0,
    config_json: "{}",
  };
  modelModalMode.value = "form";
  showModelModal.value = true;
}

function handleEditModel(row: Api.Admin.AIModel) {
  modelModalTitle.value = $t("page.system.aiConfig.editModel");
  modelForm.value = { ...row };
  modelModalMode.value = "form";
  showModelModal.value = true;
}

function closeModelForm() {
  modelModalMode.value = "list";
}

async function handleSaveModel() {
  if (modelForm.value.id) {
    await fetchUpdateAIModel(modelForm.value.id, modelForm.value);
    message.success($t("page.system.aiConfig.updateSuccess"));
  } else {
    await fetchCreateAIModel(modelForm.value as Api.Admin.AIModel);
    message.success($t("page.system.aiConfig.createSuccess"));
  }
  // Refresh data
  const { data } = await fetchGetAIProviders();
  if (data) {
    providers.value = data;
    const current = data.find((p) => p.id === currentProviderId.value);
    if (current) {
      activeModels.value = current.models || [];
    }
  }
  modelModalMode.value = "list";
}

async function handleDeleteModel(row: Api.Admin.AIModel) {
  await fetchDeleteAIModel(row.id);
  message.success($t("page.system.aiConfig.deleteSuccess"));
  // Refresh data
  const { data } = await fetchGetAIProviders();
  if (data) {
    providers.value = data;
    const current = data.find((p) => p.id === currentProviderId.value);
    if (current) {
      activeModels.value = current.models || [];
    }
  }
}

onMounted(() => {
  getProviders();
});
</script>

<template>
  <div class="h-full p-4">
    <NCard
      :title="$t('page.system.aiConfig.title')"
      :bordered="false"
      class="h-full rounded-16px shadow-sm"
    >
      <template #header-extra>
        <NButton type="primary" @click="handleAddProvider">
          {{ $t("page.system.aiConfig.addProvider") }}
        </NButton>
      </template>
      <NDataTable
        :columns="columns"
        :data="providers"
        :loading="loading"
        :pagination="false"
      />
    </NCard>

    <!-- Provider Modal -->
    <NModal
      v-model:show="showProviderModal"
      :title="providerModalTitle"
      preset="card"
      class="w-500px"
    >
      <NForm :model="providerForm" label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.system.aiConfig.providerName')" path="name">
          <NInput
            v-model:value="providerForm.name"
            :placeholder="$t('page.system.aiConfig.providerNamePlaceholder')"
          />
        </NFormItem>
        <NFormItem label="API Key" path="api_key">
          <NInput
            v-model:value="providerForm.api_key"
            type="password"
            show-password-on="click"
            :placeholder="$t('page.system.aiConfig.apiKeyPlaceholder')"
          />
        </NFormItem>
        <NFormItem label="Base URL" path="base_url">
          <NInput
            v-model:value="providerForm.base_url"
            :placeholder="$t('page.system.aiConfig.baseUrlPlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.system.aiConfig.enabled')" path="is_active">
          <NSwitch v-model:value="providerForm.is_active" />
        </NFormItem>
        <div class="flex justify-end gap-2">
          <NButton @click="showProviderModal = false">{{ $t("common.cancel") }}</NButton>
          <NButton type="primary" @click="handleSaveProvider">
            {{ $t("common.confirm") }}
          </NButton>
        </div>
      </NForm>
    </NModal>

    <!-- Models Modal -->
    <NModal
      v-model:show="showModelModal"
      :title="modelDialogTitle"
      preset="card"
      :class="modelModalMode === 'list' ? 'w-800px' : 'w-500px'"
    >
      <div v-if="modelModalMode === 'list'">
        <div class="mb-4 flex justify-end">
          <NButton type="primary" size="small" @click="handleAddModel">
            {{ $t("page.system.aiConfig.addModel") }}
          </NButton>
        </div>
        <NDataTable :columns="modelColumns" :data="activeModels" />
      </div>

      <NForm v-else :model="modelForm" label-placement="left" :label-width="100">
        <NFormItem :label="$t('page.system.aiConfig.modelCode')" path="model_code">
          <NInput
            v-model:value="modelForm.model_code"
            :placeholder="$t('page.system.aiConfig.modelCodePlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.system.aiConfig.displayName')" path="display_name">
          <NInput
            v-model:value="modelForm.display_name"
            :placeholder="$t('page.system.aiConfig.displayNamePlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.system.aiConfig.setDefault')" path="is_default">
          <NSwitch v-model:value="modelForm.is_default" />
        </NFormItem>
        <NFormItem :label="$t('page.system.aiConfig.runParams')" path="config_json">
          <NInput
            v-model:value="modelForm.config_json"
            type="textarea"
            placeholder="{&quot;temperature&quot;: 0.7}"
          />
        </NFormItem>
        <div class="flex justify-end gap-2">
          <NButton @click="closeModelForm">{{ $t("common.cancel") }}</NButton>
          <NButton type="primary" @click="handleSaveModel">
            {{ $t("common.confirm") }}
          </NButton>
        </div>
      </NForm>
    </NModal>
  </div>
</template>

<style scoped></style>
