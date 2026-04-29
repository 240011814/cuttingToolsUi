<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useMessage } from 'naive-ui';
import {
  fetchGetPermissions,
  fetchGetRolePermissions,
  fetchGetRoles,
  fetchUpdateRolePermissions
} from '@/service/api';
import { useAuth } from '@/hooks/business/auth';

const message = useMessage();
const { hasAuth } = useAuth();
const roles = ref<Api.Admin.Role[]>([]);
const permissions = ref<Api.Admin.Permission[]>([]);
const selectedRole = ref('R_ADMIN');
const checkedPermissions = ref<string[]>([]);
const loading = ref(false);
const saving = ref(false);

const roleOptions = computed(() =>
  roles.value.filter(role => role.code !== 'R_SUPER').map(role => ({ label: `${role.name} (${role.code})`, value: role.code }))
);

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
    selectedRole.value = roleOptions.value[0]?.value || 'R_ADMIN';
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

async function handleSave() {
  saving.value = true;
  const { error } = await fetchUpdateRolePermissions(selectedRole.value, checkedPermissions.value);
  if (!error) {
    message.success('权限保存成功');
  }
  saving.value = false;
}

watch(selectedRole, loadRolePermissions);

onMounted(async () => {
  await loadBaseData();
  loadRolePermissions();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <NCard :bordered="false" shadow="sm" class="flex-1">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-18px font-bold">权限管理</span>
          <NButton type="primary" :loading="saving" :disabled="!hasAuth('system:permission:update')" @click="handleSave">
            <template #icon>
              <SvgIcon icon="mdi:content-save-outline" />
            </template>
            保存配置
          </NButton>
        </div>
      </template>

      <div class="flex flex-col gap-4">
        <div class="flex items-center gap-3">
          <span class="w-72px text-right text-gray-500">角色</span>
          <NSelect v-model:value="selectedRole" :options="roleOptions" :loading="loading" style="width: 260px" />
        </div>

        <NSpin :show="loading">
          <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <NCard v-for="group in groupedPermissions" :key="group.name" size="small" :title="group.name" :bordered="true">
              <NCheckboxGroup v-model:value="checkedPermissions">
                <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <NCheckbox v-for="item in group.items" :key="item.code" :value="item.code">
                    <div class="flex flex-col">
                      <span>{{ item.name }}</span>
                      <span class="text-12px text-gray-400">{{ item.code }}</span>
                    </div>
                  </NCheckbox>
                </div>
              </NCheckboxGroup>
            </NCard>
          </div>
        </NSpin>
      </div>
    </NCard>
  </div>
</template>
