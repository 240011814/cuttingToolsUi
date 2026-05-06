<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import { NButton, NCard, NDataTable, NForm, NFormItem, NInput, NModal, NPopconfirm, NSpace, NSwitch, NTag, useMessage } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { fetchCreateAIModel, fetchCreateAIProvider, fetchDeleteAIModel, fetchDeleteAIProvider, fetchGetAIProviders, fetchUpdateAIModel, fetchUpdateAIProvider } from '@/service/api/admin';

const message = useMessage();
const loading = ref(false);
const providers = ref<Api.Admin.AIProvider[]>([]);

// Provider Modal
const showProviderModal = ref(false);
const providerModalTitle = ref('');
const providerForm = ref<Partial<Api.Admin.AIProvider>>({
  name: '',
  api_key: '',
  base_url: '',
  is_active: false
});

// Model Modal
const showModelModal = ref(false);
const modelModalTitle = ref('');
const currentProviderId = ref<number | null>(null);
const modelForm = ref<Partial<Api.Admin.AIModel>>({
  model_code: '',
  display_name: '',
  is_default: false,
  config_json: '{}'
});

const columns: DataTableColumns<Api.Admin.AIProvider> = [
  { title: '提供商', key: 'name', width: 120 },
  { 
    title: 'API Key', 
    key: 'api_key',
    render(row) {
      if (!row.api_key) return '未配置';
      return row.api_key.length > 10 ? `${row.api_key.slice(0, 6)}***${row.api_key.slice(-4)}` : '******';
    }
  },
  { title: 'Base URL', key: 'base_url' },
  {
    title: '状态',
    key: 'is_active',
    render(row) {
      return h(NSwitch, {
        value: row.is_active,
        onUpdateValue: (val: boolean) => handleToggleProviderStatus(row, val)
      });
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(NSpace, {}, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => handleEditProvider(row) }, { default: () => '编辑' }),
          h(NButton, { size: 'small', type: 'primary', ghost: true, onClick: () => handleManageModels(row) }, { default: () => '管理模型' }),
          h(NPopconfirm, { onPositiveClick: () => handleDeleteProvider(row.id) }, {
            default: () => '确认删除该提供商及其所有模型吗？',
            trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => '删除' })
          })
        ]
      });
    }
  }
];

async function getProviders() {
  loading.value = true;
  const { data } = await fetchGetAIProviders();
  if (data) {
    providers.value = data;
  }
  loading.value = false;
}

function handleAddProvider() {
  providerModalTitle.value = '新增提供商';
  providerForm.value = { name: '', api_key: '', base_url: '', is_active: true };
  showProviderModal.value = true;
}

function handleEditProvider(row: Api.Admin.AIProvider) {
  providerModalTitle.value = '编辑提供商';
  providerForm.value = { ...row };
  showProviderModal.value = true;
}

async function handleSaveProvider() {
  if (providerForm.value.id) {
    await fetchUpdateAIProvider(providerForm.value.id, providerForm.value);
    message.success('更新成功');
  } else {
    await fetchCreateAIProvider(providerForm.value);
    message.success('创建成功');
  }
  showProviderModal.value = false;
  getProviders();
}

async function handleToggleProviderStatus(row: Api.Admin.AIProvider, val: boolean) {
  await fetchUpdateAIProvider(row.id, { ...row, is_active: val });
  message.success(val ? '已启用' : '已禁用');
  getProviders();
}

async function handleDeleteProvider(id: number) {
  await fetchDeleteAIProvider(id);
  message.success('删除成功');
  getProviders();
}

// Model Management
const activeModels = ref<Api.Admin.AIModel[]>([]);
const modelColumns: DataTableColumns<Api.Admin.AIModel> = [
  { title: '模型代码', key: 'model_code' },
  { title: '显示名称', key: 'display_name' },
  {
    title: '默认',
    key: 'is_default',
    render(row) {
      return row.is_default ? h(NTag, { type: 'success' }, { default: () => '默认' }) : null;
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(NSpace, {}, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => handleEditModel(row) }, { default: () => '编辑' }),
          h(NPopconfirm, { onPositiveClick: () => handleDeleteModel(row) }, {
            default: () => '确认删除该模型吗？',
            trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => '删除' })
          })
        ]
      });
    }
  }
];

function handleManageModels(row: Api.Admin.AIProvider) {
  currentProviderId.value = row.id;
  activeModels.value = row.models || [];
  showModelModal.value = true;
}

function handleAddModel() {
  modelModalTitle.value = '新增模型';
  modelForm.value = { 
    provider_id: currentProviderId.value!,
    model_code: '', 
    display_name: '', 
    is_default: activeModels.value.length === 0,
    config_json: '{}' 
  };
  showInnerModelModal.value = true;
}

function handleEditModel(row: Api.Admin.AIModel) {
  modelModalTitle.value = '编辑模型';
  modelForm.value = { ...row };
  showInnerModelModal.value = true;
}

const showInnerModelModal = ref(false);

async function handleSaveModel() {
  if (modelForm.value.id) {
    await fetchUpdateAIModel(modelForm.value.id, modelForm.value);
    message.success('更新成功');
  } else {
    await fetchCreateAIModel(modelForm.value as Api.Admin.AIModel);
    message.success('创建成功');
  }
  showInnerModelModal.value = false;
  // Refresh data
  const { data } = await fetchGetAIProviders();
  if (data) {
    providers.value = data;
    const current = data.find(p => p.id === currentProviderId.value);
    if (current) {
      activeModels.value = current.models || [];
    }
  }
}

async function handleDeleteModel(row: Api.Admin.AIModel) {
  await fetchDeleteAIModel(row.id);
  message.success('删除成功');
  // Refresh data
  const { data } = await fetchGetAIProviders();
  if (data) {
    providers.value = data;
    const current = data.find(p => p.id === currentProviderId.value);
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
    <NCard title="AI 配置管理" :bordered="false" class="h-full rounded-16px shadow-sm">
      <template #header-extra>
        <NButton type="primary" @click="handleAddProvider">新增提供商</NButton>
      </template>
      <NDataTable :columns="columns" :data="providers" :loading="loading" :pagination="false" />
    </NCard>

    <!-- Provider Modal -->
    <NModal v-model:show="showProviderModal" :title="providerModalTitle" preset="card" class="w-500px">
      <NForm :model="providerForm" label-placement="left" :label-width="100">
        <NFormItem label="名称" path="name">
          <NInput v-model:value="providerForm.name" placeholder="如 DeepSeek, Gemini" />
        </NFormItem>
        <NFormItem label="API Key" path="api_key">
          <NInput v-model:value="providerForm.api_key" type="password" show-password-on="click" placeholder="请输入 API Key" />
        </NFormItem>
        <NFormItem label="Base URL" path="base_url">
          <NInput v-model:value="providerForm.base_url" placeholder="可选，如 https://api.deepseek.com/v1" />
        </NFormItem>
        <NFormItem label="是否启用" path="is_active">
          <NSwitch v-model:value="providerForm.is_active" />
        </NFormItem>
        <div class="flex justify-end gap-2">
          <NButton @click="showProviderModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveProvider">确定</NButton>
        </div>
      </NForm>
    </NModal>

    <!-- Models List Modal -->
    <NModal v-model:show="showModelModal" title="模型管理" preset="card" class="w-800px">
      <div class="mb-4 flex justify-end">
        <NButton type="primary" size="small" @click="handleAddModel">新增模型</NButton>
      </div>
      <NDataTable :columns="modelColumns" :data="activeModels" />
    </NModal>

    <!-- Model Edit Modal -->
    <NModal v-model:show="showInnerModelModal" :title="modelModalTitle" preset="card" class="w-500px" :z-index="2001">
      <NForm :model="modelForm" label-placement="left" :label-width="100">
        <NFormItem label="模型代码" path="model_code">
          <NInput v-model:value="modelForm.model_code" placeholder="如 deepseek-chat" />
        </NFormItem>
        <NFormItem label="显示名称" path="display_name">
          <NInput v-model:value="modelForm.display_name" placeholder="如 DeepSeek Chat" />
        </NFormItem>
        <NFormItem label="设为默认" path="is_default">
          <NSwitch v-model:value="modelForm.is_default" />
        </NFormItem>
        <NFormItem label="运行参数 (JSON)" path="config_json">
          <NInput v-model:value="modelForm.config_json" type="textarea" placeholder='{"temperature": 0.7}' />
        </NFormItem>
        <div class="flex justify-end gap-2">
          <NButton @click="showInnerModelModal = false">取消</NButton>
          <NButton type="primary" @click="handleSaveModel">确定</NButton>
        </div>
      </NForm>
    </NModal>
  </div>
</template>

<style scoped></style>
