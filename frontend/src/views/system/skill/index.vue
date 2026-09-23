<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import {
  NButton, NCard, NDataTable, NEmpty, NForm, NFormItem,
  NInput, NModal, NPopconfirm, NSelect, NSpace, NSwitch, NTag, useMessage,
} from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import { useAuth } from '@/hooks/business/auth';
import {
  fetchSkills, fetchCreateSkill, fetchUpdateSkill, fetchDeleteSkill,
} from '@/service/api';
import type { AISkillItem, AISkillCreateRequest } from '@/service/api';

const message = useMessage();
const appStore = useAppStore();
const { hasAuth } = useAuth();

const loading = ref(false);
const items = ref<AISkillItem[]>([]);
const searchText = ref('');

const showModal = ref(false);
const editingId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);
const form = reactive<AISkillCreateRequest>({
  name: '', description: '', context: 'inline', agent: '', model: '', content: '', enabled: true,
});

const contextOptions = [
  { label: 'inline（内联返回）', value: 'inline' },
  { label: 'fork（独立子 Agent）', value: 'fork' },
  { label: 'fork_with_context（携带上下文子 Agent）', value: 'fork_with_context' },
];

const contextTypeMap: Record<string, 'default' | 'info' | 'warning'> = {
  inline: 'default',
  fork: 'info',
  fork_with_context: 'warning',
};

const contextLabelMap: Record<string, string> = {
  inline: 'inline',
  fork: 'fork',
  fork_with_context: 'fork+ctx',
};

const rules: FormRules = {
  name: { required: true, message: '请输入 Skill 名称', trigger: 'blur' },
  content: { required: true, message: '请输入 Skill 内容', trigger: 'blur' },
};

const filtered = computed(() => {
  const q = searchText.value.trim().toLowerCase();
  if (!q) return items.value;
  return items.value.filter(i =>
    i.name.toLowerCase().includes(q) ||
    i.description.toLowerCase().includes(q) ||
    i.model.toLowerCase().includes(q)
  );
});

const stats = computed(() => ({
  total: items.value.length,
  enabled: items.value.filter(i => i.enabled).length,
}));

function formatTime(value: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('zh-CN', { hour12: false });
}

async function loadData() {
  loading.value = true;
  try {
    const { data } = await fetchSkills();
    if (data) items.value = data;
  } catch { /* ignore */ }
  loading.value = false;
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, {
    name: '', description: '', context: 'inline', agent: '', model: '', content: '', enabled: true,
  });
  showModal.value = true;
}

function openEdit(row: AISkillItem) {
  editingId.value = row.id;
  Object.assign(form, {
    name: row.name,
    description: row.description,
    context: row.context || 'inline',
    agent: row.agent,
    model: row.model,
    content: row.content,
    enabled: row.enabled,
  });
  showModal.value = true;
}

async function handleSubmit() {
  try { await formRef.value?.validate(); } catch { return; }
  const result = editingId.value === null
    ? await fetchCreateSkill(form)
    : await fetchUpdateSkill(editingId.value, form);
  if (!result.error) {
    message.success(editingId.value === null ? '创建成功' : '更新成功');
    showModal.value = false;
    loadData();
  } else {
    message.error(result.error.message || '操作失败');
  }
}

async function handleToggle(row: AISkillItem, enabled: boolean) {
  const { error } = await fetchUpdateSkill(row.id, { enabled });
  if (error) {
    message.error('更新失败');
    return;
  }
  row.enabled = enabled;
  message.success(enabled ? '已启用' : '已停用');
}

async function handleDelete(id: number) {
  const { error } = await fetchDeleteSkill(id);
  if (!error) {
    message.success('删除成功');
    loadData();
  } else {
    message.error('删除失败');
  }
}

const columns: DataTableColumns<AISkillItem> = [
  { title: '名称', key: 'name', width: 160, ellipsis: { tooltip: true } },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
  {
    title: '模式', key: 'context', width: 110,
    render: row => h(NTag, { type: contextTypeMap[row.context] || 'default', size: 'small' },
      { default: () => contextLabelMap[row.context] || row.context || 'inline' }),
  },
  {
    title: '模型', key: 'model', width: 120,
    render: row => row.model || '-',
  },
  {
    title: '状态', key: 'enabled', width: 90,
    render: row => h(NSwitch, {
      value: row.enabled,
      size: 'small',
      disabled: !hasAuth('system:skill:update'),
      'onUpdate:value': (val: boolean) => handleToggle(row, val),
    }),
  },
  {
    title: '更新时间', key: 'updated_at', width: 170,
    render: row => formatTime(row.updated_at),
  },
  {
    title: '操作', key: 'actions', width: 140,
    render: row => h(NSpace, { size: 'small' }, {
      default: () => [
        hasAuth('system:skill:update')
          ? h(NButton, { size: 'small', quaternary: true, type: 'primary', onClick: () => openEdit(row) }, { default: () => '编辑' })
          : null,
        hasAuth('system:skill:delete')
          ? h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
              trigger: () => h(NButton, { size: 'small', quaternary: true, type: 'error' }, { default: () => '删除' }),
              default: () => `确认删除 Skill「${row.name}」？`,
            })
          : null,
      ],
    }),
  },
];

onMounted(loadData);
</script>

<template>
  <div class="h-full flex-col flex gap-4" :class="appStore.isMobile ? 'p-2' : 'p-4'">
    <!-- ====== 统计 ====== -->
    <div class="grid grid-cols-2 gap-3">
      <NCard :bordered="false" size="small" class="text-center">
        <div class="text-2xl font-bold text-blue-500">{{ stats.total }}</div>
        <div class="text-xs text-gray-500 mt-1">Skill 总数</div>
      </NCard>
      <NCard :bordered="false" size="small" class="text-center">
        <div class="text-2xl font-bold text-green-500">{{ stats.enabled }}</div>
        <div class="text-xs text-gray-500 mt-1">已启用</div>
      </NCard>
    </div>

    <!-- ====== 主卡片 ====== -->
    <NCard :bordered="false" :shadow="appStore.isMobile ? false : 'sm'" class="flex-1">
      <div class="mb-4 flex flex-wrap items-center gap-3" :class="appStore.isMobile ? 'flex-col' : 'justify-between'">
        <NInput
          v-model:value="searchText"
          placeholder="搜索名称/描述/模型..."
          clearable
          :style="{ width: appStore.isMobile ? '100%' : '240px' }"
        >
          <template #prefix><SvgIcon icon="mdi:magnify" /></template>
        </NInput>
        <NButton v-if="hasAuth('system:skill:create')" type="primary" @click="openCreate">
          <template #icon><SvgIcon icon="mdi:plus" /></template>
          新增 Skill
        </NButton>
      </div>

      <NDataTable
        v-if="filtered.length > 0"
        :columns="columns"
        :data="filtered"
        :loading="loading"
        :scroll-x="1000"
        :row-key="row => row.id"
        size="small"
      />
      <NEmpty v-else description="暂无 Skill, 点击右上角新增" class="py-12" />
    </NCard>

    <!-- ====== 编辑弹窗 ====== -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="editingId === null ? '新增 Skill' : '编辑 Skill'"
      :style="{ width: appStore.isMobile ? '95vw' : '720px' }"
      :segmented="{ content: true, footer: true }"
    >
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="90">
        <NFormItem label="名称" path="name">
          <NInput v-model:value="form.name" placeholder="唯一标识, 如 pdf / xlsx" />
        </NFormItem>
        <NFormItem label="描述">
          <NInput v-model:value="form.description" type="textarea" :rows="2" placeholder="描述该 Skill 的能力与使用场景" />
        </NFormItem>
        <NFormItem label="执行模式">
          <NSelect v-model:value="form.context" :options="contextOptions" />
        </NFormItem>
        <NFormItem v-if="form.context !== 'inline'" label="子 Agent">
          <NInput v-model:value="form.agent" placeholder="fork 使用的子 Agent 名称(留空用默认)" />
        </NFormItem>
        <NFormItem label="模型">
          <NInput v-model:value="form.model" placeholder="留空使用默认模型" />
        </NFormItem>
        <NFormItem label="内容" path="content">
          <NInput
            v-model:value="form.content"
            type="textarea"
            :rows="10"
            placeholder="SKILL.md 正文 (frontmatter 之后的 markdown 指令)"
          />
        </NFormItem>
        <NFormItem label="启用">
          <NSwitch v-model:value="form.enabled" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" @click="handleSubmit">保存</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>