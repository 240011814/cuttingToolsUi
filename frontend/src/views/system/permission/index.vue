<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useMessage, useDialog } from 'naive-ui';
import {
  fetchGetPermissions,
  fetchGetRolePermissions,
  fetchGetRoles,
  fetchUpdateRolePermissions,
  fetchCreateRole,
  fetchDeleteRole,
  fetchCreatePermission,
  fetchUpdatePermission,
  fetchDeletePermission
} from '@/service/api';
import { useAuth } from '@/hooks/business/auth';

const message = useMessage();
const dialog = useDialog();
const { hasAuth } = useAuth();

const roles = ref<Api.Admin.Role[]>([]);
const permissions = ref<Api.Admin.Permission[]>([]);
const selectedRole = ref('');
const checkedPermissions = ref<string[]>([]);
const loading = ref(false);
const saving = ref(false);

const showAddRoleModal = ref(false);
const addRoleLoading = ref(false);
const newRole = ref({
  code: '',
  name: '',
  description: ''
});

const showPermissionModal = ref(false);
const permissionLoading = ref(false);
const isEditPermission = ref(false);
const permissionForm = ref({
  id: 0,
  code: '',
  name: '',
  groupName: ''
});

const groupedPermissions = computed(() => {
  const groups = new Map<string, Api.Admin.Permission[]>();
  permissions.value.forEach(permission => {
    const items = groups.get(permission.groupName) || [];
    items.push(permission);
    groups.set(permission.groupName, items);
  });
  return Array.from(groups.entries()).map(([name, items]) => ({ name, items }));
});

async function loadBaseData() {
  const [roleResult, permissionResult] = await Promise.all([fetchGetRoles(), fetchGetPermissions()]);
  if (!roleResult.error) {
    roles.value = roleResult.data;
    if (roles.value.length > 0 && !selectedRole.value) {
      selectedRole.value = roles.value[0].code;
    }
  }
  if (!permissionResult.error) {
    permissions.value = permissionResult.data;
  }
}

async function loadRolePermissions() {
  if (!selectedRole.value) return;
  loading.value = true;
  const { data, error } = await fetchGetRolePermissions(selectedRole.value);
  if (!error) {
    checkedPermissions.value = data;
  }
  loading.value = false;
}

async function handleSaveRolePermissions() {
  if (!selectedRole.value) return;
  saving.value = true;
  const { error } = await fetchUpdateRolePermissions(selectedRole.value, checkedPermissions.value);
  if (!error) {
    message.success('权限保存成功');
  }
  saving.value = false;
}

async function handleAddRole() {
  if (!newRole.value.code || !newRole.value.name) {
    message.warning('请填写完整的角色信息');
    return;
  }
  addRoleLoading.value = true;
  const { error } = await fetchCreateRole(newRole.value);
  if (!error) {
    message.success('角色创建成功');
    showAddRoleModal.value = false;
    newRole.value = { code: '', name: '', description: '' };
    await loadBaseData();
  }
  addRoleLoading.value = false;
}

function confirmDeleteRole(role: Api.Admin.Role) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除角色 "${role.name}" 吗？此操作不可撤销。`,
    positiveText: '确认',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await fetchDeleteRole(role.code);
      if (!error) {
        message.success('角色已删除');
        if (selectedRole.value === role.code) {
          selectedRole.value = '';
        }
        await loadBaseData();
      }
    }
  });
}

function handleAddPermission(groupName?: string) {
  isEditPermission.value = false;
  permissionForm.value = { id: 0, code: '', name: '', groupName: groupName || '' };
  showPermissionModal.value = true;
}

function handleEditPermission(item: Api.Admin.Permission) {
  isEditPermission.value = true;
  permissionForm.value = { ...item };
  showPermissionModal.value = true;
}

async function handleSubmitPermission() {
  permissionLoading.value = true;
  let error;
  if (isEditPermission.value) {
    ({ error } = await fetchUpdatePermission(permissionForm.value.id, permissionForm.value));
  } else {
    ({ error } = await fetchCreatePermission(permissionForm.value));
  }
  
  if (!error) {
    message.success(isEditPermission.value ? '权限点更新成功' : '权限点创建成功');
    showPermissionModal.value = false;
    await loadBaseData();
    if (selectedRole.value) {
      await loadRolePermissions();
    }
  }
  permissionLoading.value = false;
}

function confirmDeletePermission(item: Api.Admin.Permission) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除权限点 "${item.name}" (${item.code}) 吗？`,
    positiveText: '确认',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await fetchDeletePermission(item.id);
      if (!error) {
        message.success('权限点已删除');
        await loadBaseData();
        if (selectedRole.value) {
          await loadRolePermissions();
        }
      }
    }
  });
}

const currentRoleName = computed(() => {
  return roles.value.find(r => r.code === selectedRole.value)?.name || '';
});

watch(selectedRole, loadRolePermissions);

onMounted(async () => {
  await loadBaseData();
  if (selectedRole.value) {
    loadRolePermissions();
  }
});
</script>

<template>
  <div class="h-full flex gap-4 p-4">
    <!-- Role List Sidebar -->
    <NCard :bordered="false" shadow="sm" class="w-280px flex-shrink-0 flex flex-col" content-style="padding: 0; display: flex; flex-direction: column;">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-16px font-bold">角色列表</span>
          <NButton size="small" type="primary" quaternary circle @click="showAddRoleModal = true">
            <template #icon>
              <SvgIcon icon="mdi:plus" />
            </template>
          </NButton>
        </div>
      </template>

      <NScrollbar class="flex-1 px-2 py-2">
        <div class="flex flex-col gap-1">
          <div
            v-for="role in roles"
            :key="role.code"
            class="group relative flex cursor-pointer items-center justify-between rounded-8px px-4 py-3 transition-all hover:bg-primary/5"
            :class="selectedRole === role.code ? 'bg-primary/10 text-primary shadow-sm' : 'text-gray-600'"
            @click="selectedRole = role.code"
          >
            <div class="flex flex-col overflow-hidden">
              <span class="truncate font-medium">{{ role.name }}</span>
              <span class="text-12px opacity-60">{{ role.code }}</span>
            </div>
            
            <NButton
              v-if="role.code !== 'R_SUPER' && role.code !== 'R_ADMIN' && role.code !== 'R_USER'"
              size="tiny"
              type="error"
              quaternary
              circle
              class="opacity-0 group-hover:opacity-100"
              @click.stop="confirmDeleteRole(role)"
            >
              <template #icon>
                <SvgIcon icon="mdi:trash-can-outline" />
              </template>
            </NButton>
          </div>
        </div>
      </NScrollbar>
    </NCard>

    <!-- Permission Area -->
    <NCard :bordered="false" shadow="sm" class="flex-1" content-style="display: flex; flex-direction: column;">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="text-18px font-bold">权限配置</span>
            <NTag v-if="selectedRole" type="primary" round size="small">{{ currentRoleName }}</NTag>
          </div>
          <div class="flex gap-2">
            <NButton type="info" secondary @click="handleAddPermission()">
              <template #icon>
                <SvgIcon icon="mdi:plus-circle-outline" />
              </template>
              新增权限点
            </NButton>
            <NButton
              type="primary"
              :loading="saving"
              :disabled="!selectedRole || selectedRole === 'R_SUPER'"
              @click="handleSaveRolePermissions"
            >
              <template #icon>
                <SvgIcon icon="mdi:content-save-outline" />
              </template>
              保存配置
            </NButton>
          </div>
        </div>
      </template>

      <div v-if="!selectedRole" class="flex flex-1 items-center justify-center text-gray-400">
        请选择一个角色以配置权限
      </div>

      <div v-else-if="selectedRole === 'R_SUPER'" class="flex flex-1 flex-col items-center justify-center text-gray-400 gap-2">
        <SvgIcon icon="mdi:shield-check" class="text-48px text-primary/40" />
        <span>超级管理员默认拥有所有权限，无需配置</span>
      </div>

      <NSpin v-else :show="loading" class="flex-1">
        <NScrollbar class="pr-2">
          <div class="grid grid-cols-1 gap-4 lg:grid-cols-2 pb-4">
            <NCard
              v-for="group in groupedPermissions"
              :key="group.name"
              size="small"
              :bordered="true"
              class="rounded-12px border-gray-100 transition-shadow hover:shadow-md"
            >
              <template #header>
                <div class="flex items-center justify-between w-full">
                  <span class="font-bold">{{ group.name }}</span>
                  <NButton size="tiny" quaternary type="primary" @click="handleAddPermission(group.name)">
                    <template #icon><SvgIcon icon="mdi:plus" /></template>
                  </NButton>
                </div>
              </template>
              
              <NCheckboxGroup v-model:value="checkedPermissions">
                <div class="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-1">
                  <div v-for="item in group.items" :key="item.code" class="group/item flex items-center justify-between py-1 px-2 rounded hover:bg-gray-50">
                    <NCheckbox :value="item.code" class="flex-1">
                      <div class="flex flex-col">
                        <span class="group-hover/item:text-primary transition-colors">{{ item.name }}</span>
                        <span class="text-12px text-gray-400 font-mono">{{ item.code }}</span>
                      </div>
                    </NCheckbox>
                    <div class="flex opacity-0 group-hover/item:opacity-100 transition-opacity">
                      <NButton size="tiny" quaternary type="info" @click.stop="handleEditPermission(item)">
                        <template #icon><SvgIcon icon="mdi:pencil-outline" /></template>
                      </NButton>
                      <NButton size="tiny" quaternary type="error" @click.stop="confirmDeletePermission(item)">
                        <template #icon><SvgIcon icon="mdi:trash-can-outline" /></template>
                      </NButton>
                    </div>
                  </div>
                </div>
              </NCheckboxGroup>
            </NCard>
          </div>
        </NScrollbar>
      </NSpin>
    </NCard>

    <!-- Add Role Modal -->
    <NModal v-model:show="showAddRoleModal" preset="card" title="新增角色" class="w-450px" :segmented="{ content: true, footer: true }">
      <NForm>
        <NFormItem label="角色名称" path="name">
          <NInput v-model:value="newRole.name" placeholder="请输入角色名称，如：高级用户" />
        </NFormItem>
        <NFormItem label="角色编码" path="code">
          <NInput v-model:value="newRole.code" placeholder="请输入唯一编码，如：R_SENIOR" />
        </NFormItem>
        <NFormItem label="描述" path="description">
          <NInput v-model:value="newRole.description" type="textarea" placeholder="请输入角色描述" />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showAddRoleModal = false">取消</NButton>
          <NButton type="primary" :loading="addRoleLoading" @click="handleAddRole">确定</NButton>
        </div>
      </template>
    </NModal>

    <!-- Add/Edit Permission Modal -->
    <NModal v-model:show="showPermissionModal" preset="card" :title="isEditPermission ? '编辑权限点' : '新增权限点'" class="w-450px" :segmented="{ content: true, footer: true }">
      <NForm>
        <NFormItem label="权限名称" path="name">
          <NInput v-model:value="permissionForm.name" placeholder="请输入权限显示名称" />
        </NFormItem>
        <NFormItem label="权限编码" path="code">
          <NInput v-model:value="permissionForm.code" :disabled="isEditPermission" placeholder="请输入唯一编码，如：system:user:list" />
        </NFormItem>
        <NFormItem label="所属分组" path="groupName">
          <NInput v-model:value="permissionForm.groupName" placeholder="请输入分组名称，如：用户管理" />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showPermissionModal = false">取消</NButton>
          <NButton type="primary" :loading="permissionLoading" @click="handleSubmitPermission">确定</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
:deep(.n-card-header__main) {
  width: 100%;
}
</style>
