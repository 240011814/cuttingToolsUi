<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useMessage } from "naive-ui";
import type { FormInst, FormRules } from "naive-ui";
import {
  fetchChangePassword,
  fetchGetUserProfile,
  fetchUpdateProfile,
} from "@/service/api";
import { useAuthStore } from "@/store/modules/auth";
import { $t } from "@/locales";

defineOptions({ name: "UserProfileModal" });

interface Props {
  show: boolean;
}

interface Emits {
  (e: "update:show", value: boolean): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const message = useMessage();
const authStore = useAuthStore();
const loading = ref(false);
const profile = ref<Api.Admin.UserProfile | null>(null);
const activeTab = ref("profile");

const profileFormRef = ref<FormInst | null>(null);
const passwordFormRef = ref<FormInst | null>(null);

const profileForm = reactive({
  nickname: "",
});

const passwordForm = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});

const profileRules: FormRules = {
  nickname: [
    {
      required: true,
      message: $t("page.userProfile.nicknameRequired"),
      trigger: ["blur", "input"],
    },
  ],
};

const passwordRules: FormRules = {
  oldPassword: [
    {
      required: true,
      message: $t("page.userProfile.oldPasswordRequired"),
      trigger: ["blur", "input"],
    },
  ],
  newPassword: [
    {
      required: true,
      message: $t("page.userProfile.newPasswordRequired"),
      trigger: ["blur", "input"],
    },
    {
      min: 6,
      message: $t("page.userProfile.passwordMinLength"),
      trigger: ["blur", "input"],
    },
  ],
  confirmPassword: [
    {
      required: true,
      message: $t("page.userProfile.confirmPasswordRequired"),
      trigger: ["blur", "input"],
    },
    {
      validator(_rule: unknown, value: string) {
        return value === passwordForm.newPassword;
      },
      message: $t("page.userProfile.passwordMismatch"),
      trigger: ["blur", "input"],
    },
  ],
};

const showModal = computed({
  get: () => props.show,
  set: (val: boolean) => emit("update:show", val),
});

async function loadProfile() {
  loading.value = true;
  const { data, error } = await fetchGetUserProfile();
  if (!error) {
    profile.value = data;
    profileForm.nickname = data.nickname;
  }
  loading.value = false;
}

async function handleUpdateProfile() {
  await profileFormRef.value?.validate();
  const { error } = await fetchUpdateProfile({ nickname: profileForm.nickname });
  if (!error) {
    message.success($t("page.userProfile.updateSuccess"));
    authStore.userInfo.userName = profile.value?.userName || "";
    showModal.value = false;
  }
}

async function handleChangePassword() {
  await passwordFormRef.value?.validate();
  const { error } = await fetchChangePassword({
    oldPassword: passwordForm.oldPassword,
    newPassword: passwordForm.newPassword,
  });
  if (!error) {
    message.success($t("page.userProfile.passwordChangeSuccess"));
    showModal.value = false;
  }
}

function handleClose() {
  activeTab.value = "profile";
  Object.assign(passwordForm, { oldPassword: "", newPassword: "", confirmPassword: "" });
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      loadProfile();
    }
  }
);
</script>

<template>
  <NModal
    v-model:show="showModal"
    preset="card"
    :title="$t('page.userProfile.title')"
    style="width: min(480px, 92vw)"
    @after-leave="handleClose"
  >
    <NSpin :show="loading">
      <NTabs v-model:value="activeTab" type="line" animated>
        <NTabPane name="profile" :tab="$t('page.userProfile.basicInfo')">
          <NForm
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileRules"
            label-placement="left"
            label-width="80"
          >
            <NFormItem :label="$t('page.userProfile.userName')">
              <NInput :value="profile?.userName" disabled />
            </NFormItem>
            <NFormItem :label="$t('page.userProfile.nickname')" path="nickname">
              <NInput
                v-model:value="profileForm.nickname"
                :placeholder="$t('page.userProfile.nicknamePlaceholder')"
              />
            </NFormItem>
          </NForm>
          <div class="flex justify-end mt-4">
            <NButton type="primary" @click="handleUpdateProfile">
              <template #icon>
                <SvgIcon icon="mdi:content-save-outline" />
              </template>
              {{ $t("common.save") }}
            </NButton>
          </div>
        </NTabPane>

        <NTabPane name="password" :tab="$t('page.userProfile.changePassword')">
          <NForm
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-placement="left"
            label-width="80"
          >
            <NFormItem :label="$t('page.userProfile.oldPassword')" path="oldPassword">
              <NInput
                v-model:value="passwordForm.oldPassword"
                type="password"
                show-password-on="click"
                :placeholder="$t('page.userProfile.oldPasswordPlaceholder')"
              />
            </NFormItem>
            <NFormItem :label="$t('page.userProfile.newPassword')" path="newPassword">
              <NInput
                v-model:value="passwordForm.newPassword"
                type="password"
                show-password-on="click"
                :placeholder="$t('page.userProfile.newPasswordPlaceholder')"
              />
            </NFormItem>
            <NFormItem
              :label="$t('page.userProfile.confirmPassword')"
              path="confirmPassword"
            >
              <NInput
                v-model:value="passwordForm.confirmPassword"
                type="password"
                show-password-on="click"
                :placeholder="$t('page.userProfile.confirmPasswordPlaceholder')"
              />
            </NFormItem>
          </NForm>
          <div class="flex justify-end mt-4">
            <NButton type="primary" @click="handleChangePassword">
              <template #icon>
                <SvgIcon icon="mdi:lock-check-outline" />
              </template>
              {{ $t("page.userProfile.changePassword") }}
            </NButton>
          </div>
        </NTabPane>
      </NTabs>
    </NSpin>
  </NModal>
</template>

<style scoped></style>
