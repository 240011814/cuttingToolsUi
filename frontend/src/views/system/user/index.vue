<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag, useMessage } from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import {
  fetchCreateUser,
  fetchDeleteUser,
  fetchGetRoles,
  fetchGetUsers,
  fetchUpdateUser
} from '@/service/api';
import { useAuth } from '@/hooks/business/auth';

const message = useMessage();
const { hasAuth } = useAuth();
const loading = ref(false);
const users = ref<Api.Admin.User[]>([]);
const roles = ref<Api.Admin.Role[]>([]);
const keyword = ref('');
const role = ref<string | null>(null);
const showModal = ref(false);
const editingUserId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);

const form = reactive({
  userName: '',
  password: '',
  nickname: '',
  role: 'R_USER'
});

const roleOptions = computed(() => roles.value.map(item => ({ label: `${item.name} (${item.code})`, value: item.code })));
const roleNameMap = computed(() => new Map(roles.value.map(item => [item.code, item.name])));

const rules: FormRules = {
  userName: [{ required: true, message: '请输入用户名', trigger: ['blur', 'input'] }],
  password: [
    {
      validator() {
        if (editingUserId.value === null && !form.password) {
          return new Error('请输入密码');
        }
        return true;
      },
      trigger: ['blur', 'input']
    }
  ],
  role: [{ required: true, message: '请选择角色', trigger: ['change', 'blur'] }]
};

const columns: DataTableColumns<Api.Admin.User> = [
  { title: '用户名', key: 'userName', minWidth: 140 },
  { title: '昵称', key: 'nickname', minWidth: 140 },
  {
    title: '角色',
    key: 'role',
    width: 180,
    render(row) {
      return h(NTag, { type: row.role === 'R_SUPER' ? 'error' : 'info', bordered: false }, {
        default: () => roleNameMap.value.get(row.role) || row.role
      });
    }
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 180,
    render(row) {
      return new Date(row.createdAt).toLocaleString();
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render(row) {
      const actions = [];
      if (hasAuth('system:user:update')) {
        actions.push(h(NButton, { size: 'small', type: 'primary', quaternary: true, onClick: () => openEdit(row) }, { default: () => '编辑' }));
      }
      if (hasAuth('system:user:delete')) {
        actions.push(
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row.userId) },
            {
              trigger: () => h(NButton, { size: 'small', type: 'error', quaternary: true }, { default: () => '删除' }),
              default: () => '确认删除该用户吗？'
            }
          )
        );
      }
      return h('div', { class: 'flex gap-1' }, actions);
    }
  }
];

async function loadRoles() {
  const { data, error } = await fetchGetRoles();
  if (!error) {
    roles.value = data;
  }
}

async function loadUsers() {
  loading.value = true;
  const { data, error } = await fetchGetUsers({ keyword: keyword.value, role: role.value || undefined });
  if (!error) {
    users.value = data;
  }
  loading.value = false;
}

function resetForm() {
  editingUserId.value = null;
  Object.assign(form, { userName: '', password: '', nickname: '', role: 'R_USER' });
}

function openCreate() {
  resetForm();
  showModal.value = true;
}

function openEdit(row: Api.Admin.User) {
  editingUserId.value = row.userId;
  Object.assign(form, { userName: row.userName, password: '', nickname: row.nickname, role: row.role });
  showModal.value = true;
}

async function handleSubmit() {
  await formRef.value?.validate();
  const payload = { password: form.password || undefined, nickname: form.nickname, role: form.role };
  const result =
    editingUserId.value === null
      ? await fetchCreateUser({ userName: form.userName, password: form.password, nickname: form.nickname, role: form.role })
      : await fetchUpdateUser(editingUserId.value, payload);
  if (!result.error) {
    message.success(editingUserId.value === null ? '创建成功' : '更新成功');
    showModal.value = false;
    loadUsers();
  }
}

async function handleDelete(userId: number) {
  const { error } = await fetchDeleteUser(userId);
  if (!error) {
    message.success('删除成功');
    loadUsers();
  }
}

onMounted(async () => {
  await loadRoles();
  loadUsers();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <NCard :bordered="false" shadow="sm" class="flex-1">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-18px font-bold">用户管理</span>
          <NButton v-if="hasAuth('system:user:create')" type="primary" @click="openCreate">
            <template #icon>
              <SvgIcon icon="mdi:account-plus-outline" />
            </template>
            新增用户
          </NButton>
        </div>
      </template>

      <div class="flex flex-col h-full gap-4">
        <div class="flex flex-wrap gap-3 items-center">
          <NInput v-model:value="keyword" clearable placeholder="搜索用户名或昵称" style="width: 260px" @keyup.enter="loadUsers" />
          <NSelect v-model:value="role" clearable :options="roleOptions" placeholder="角色" style="width: 220px" />
          <NButton type="primary" @click="loadUsers">
            <template #icon>
              <SvgIcon icon="mdi:magnify" />
            </template>
            搜索
          </NButton>
        </div>

        <NDataTable :columns="columns" :data="users" :loading="loading" :pagination="{ pageSize: 10 }" :row-key="row => row.userId" flex-height class="flex-1" />
      </div>
    </NCard>

    <NModal v-model:show="showModal" preset="card" :title="editingUserId === null ? '新增用户' : '编辑用户'" style="width: min(560px, 92vw)">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="86">
        <NFormItem label="用户名" path="userName">
          <NInput v-model:value="form.userName" :disabled="editingUserId !== null" placeholder="请输入用户名" />
        </NFormItem>
        <NFormItem label="密码" path="password">
          <NInput v-model:value="form.password" type="password" show-password-on="click" :placeholder="editingUserId === null ? '请输入密码' : '留空则不修改'" />
        </NFormItem>
        <NFormItem label="昵称" path="nickname">
          <NInput v-model:value="form.nickname" placeholder="请输入昵称" />
        </NFormItem>
        <NFormItem label="角色" path="role">
          <NSelect v-model:value="form.role" :options="roleOptions" />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" @click="handleSubmit">保存</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>
