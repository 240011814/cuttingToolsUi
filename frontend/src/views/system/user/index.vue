<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from "vue";
import { NButton, NPopconfirm, NTag, useMessage } from "naive-ui";
import type { DataTableColumns, FormInst, FormRules } from "naive-ui";
import {
  fetchCreateUser,
  fetchDeleteUser,
  fetchGetRoles,
  fetchGetUsers,
  fetchUpdateUser,
} from "@/service/api";
import { useAuth } from "@/hooks/business/auth";
import { $t } from "@/locales";

const message = useMessage();
const { hasAuth } = useAuth();
const loading = ref(false);
const users = ref<Api.Admin.User[]>([]);
const roles = ref<Api.Admin.Role[]>([]);
const keyword = ref("");
const role = ref<string | null>(null);
const showModal = ref(false);
const editingUserId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);

const form = reactive({
  userName: "",
  password: "",
  nickname: "",
  role: "R_USER",
});

const roleOptions = computed(() =>
  roles.value.map((item) => ({ label: `${item.name} (${item.code})`, value: item.code }))
);
const roleNameMap = computed(
  () => new Map(roles.value.map((item) => [item.code, item.name]))
);

const rules: FormRules = {
  userName: [
    {
      required: true,
      message: $t("page.system.user.userNameRequired"),
      trigger: ["blur", "input"],
    },
  ],
  password: [
    {
      validator() {
        if (editingUserId.value === null && !form.password) {
          return new Error($t("page.system.user.passwordRequired"));
        }
        return true;
      },
      trigger: ["blur", "input"],
    },
  ],
  role: [
    {
      required: true,
      message: $t("page.system.user.roleRequired"),
      trigger: ["change", "blur"],
    },
  ],
};

const hasRolePermission = computed(() => hasAuth("system:role:list"));

const columns = computed<DataTableColumns<Api.Admin.User>>(() => [
  { title: $t("page.system.user.userName"), key: "userName", minWidth: 140 },
  { title: $t("page.system.user.nickname"), key: "nickname", minWidth: 140 },
  ...(hasRolePermission.value
    ? [
        {
          title: $t("page.system.user.role"),
          key: "role",
          width: 180,
          render(row: Api.Admin.User) {
            return h(
              NTag,
              { type: row.role === "R_SUPER" ? "error" : "info", bordered: false },
              {
                default: () => roleNameMap.value.get(row.role) || row.role,
              }
            );
          },
        },
      ]
    : []),
  {
    title: $t("page.system.user.createdAt"),
    key: "createdAt",
    width: 180,
    render(row) {
      return new Date(row.createdAt).toLocaleString();
    },
  },
  {
    title: $t("page.system.user.actions"),
    key: "actions",
    width: 180,
    fixed: "right",
    render(row) {
      const actions = [];
      if (hasAuth("system:user:update")) {
        actions.push(
          h(
            NButton,
            {
              size: "small",
              type: "primary",
              quaternary: true,
              onClick: () => openEdit(row),
            },
            { default: () => $t("common.edit") }
          )
        );
      }
      if (hasAuth("system:user:delete")) {
        actions.push(
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row.userId) },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: "small", type: "error", quaternary: true },
                  { default: () => $t("common.delete") }
                ),
              default: () => $t("page.system.user.deleteUserConfirm"),
            }
          )
        );
      }
      return h("div", { class: "flex gap-1" }, actions);
    },
  },
]);

async function loadRoles() {
  const { data, error } = await fetchGetRoles();
  if (!error) {
    roles.value = data;
  }
}

async function loadUsers() {
  loading.value = true;
  const { data, error } = await fetchGetUsers({
    keyword: keyword.value,
    role: role.value || undefined,
  });
  if (!error) {
    users.value = data;
  }
  loading.value = false;
}

function resetForm() {
  editingUserId.value = null;
  Object.assign(form, { userName: "", password: "", nickname: "", role: "R_USER" });
}

function openCreate() {
  resetForm();
  showModal.value = true;
}

function openEdit(row: Api.Admin.User) {
  editingUserId.value = row.userId;
  Object.assign(form, {
    userName: row.userName,
    password: "",
    nickname: row.nickname,
    role: row.role,
  });
  showModal.value = true;
}

async function handleSubmit() {
  await formRef.value?.validate();
  const payload = {
    password: form.password || undefined,
    nickname: form.nickname,
    role: form.role,
  };
  const result =
    editingUserId.value === null
      ? await fetchCreateUser({
          userName: form.userName,
          password: form.password,
          nickname: form.nickname,
          role: form.role,
        })
      : await fetchUpdateUser(editingUserId.value, payload);
  if (!result.error) {
    message.success(
      editingUserId.value === null
        ? $t("page.system.user.createSuccess")
        : $t("page.system.user.updateSuccess")
    );
    showModal.value = false;
    loadUsers();
  }
}

async function handleDelete(userId: number) {
  const { error } = await fetchDeleteUser(userId);
  if (!error) {
    message.success($t("page.system.user.deleteSuccess"));
    loadUsers();
  }
}

onMounted(async () => {
  if (hasRolePermission.value) {
    await loadRoles();
  }
  loadUsers();
});
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4">
    <NCard :bordered="false" shadow="sm" class="flex-1">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-18px font-bold">{{ $t("page.system.user.title") }}</span>
          <NButton
            v-if="hasAuth('system:user:create')"
            type="primary"
            @click="openCreate"
          >
            <template #icon>
              <SvgIcon icon="mdi:account-plus-outline" />
            </template>
            {{ $t("page.system.user.addUser") }}
          </NButton>
        </div>
      </template>

      <div class="flex flex-col h-full gap-4">
        <div class="flex flex-wrap gap-3 items-center">
          <NInput
            v-model:value="keyword"
            clearable
            :placeholder="$t('page.system.user.searchPlaceholder')"
            style="width: 260px"
            @keyup.enter="loadUsers"
          />
          <NSelect
            v-if="hasRolePermission"
            v-model:value="role"
            clearable
            :options="roleOptions"
            :placeholder="$t('page.system.user.role')"
            style="width: 220px"
          />
          <NButton type="primary" @click="loadUsers">
            <template #icon>
              <SvgIcon icon="mdi:magnify" />
            </template>
            {{ $t("common.search") }}
          </NButton>
        </div>

        <NDataTable
          :columns="columns"
          :data="users"
          :loading="loading"
          :pagination="{ pageSize: 10 }"
          :row-key="(row) => row.userId"
          flex-height
          class="flex-1"
        />
      </div>
    </NCard>

    <NModal
      v-model:show="showModal"
      preset="card"
      :title="
        editingUserId === null
          ? $t('page.system.user.addUser')
          : $t('page.system.user.editUser')
      "
      style="width: min(560px, 92vw)"
    >
      <NForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-placement="left"
        label-width="86"
      >
        <NFormItem :label="$t('page.system.user.userName')" path="userName">
          <NInput
            v-model:value="form.userName"
            :disabled="editingUserId !== null"
            :placeholder="$t('page.system.user.userNameRequired')"
          />
        </NFormItem>
        <NFormItem :label="$t('page.system.user.passwordPlaceholder')" path="password">
          <NInput
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            :placeholder="
              editingUserId === null
                ? $t('page.system.user.passwordPlaceholder')
                : $t('page.system.user.passwordEditPlaceholder')
            "
          />
        </NFormItem>
        <NFormItem :label="$t('page.system.user.nickname')" path="nickname">
          <NInput
            v-model:value="form.nickname"
            :placeholder="$t('page.system.user.nicknamePlaceholder')"
          />
        </NFormItem>
        <NFormItem
          v-if="hasRolePermission"
          :label="$t('page.system.user.role')"
          path="role"
        >
          <NSelect v-model:value="form.role" :options="roleOptions" />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showModal = false">{{ $t("common.cancel") }}</NButton>
          <NButton type="primary" @click="handleSubmit">
            {{
              $t("common.confirm")
            }}
          </NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>
