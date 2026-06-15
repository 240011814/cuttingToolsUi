<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useMessage } from 'naive-ui';
import type { FormInst, FormRules } from 'naive-ui';
import { fetchChangePassword, fetchGetUserProfile, fetchUpdateProfile } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';
import { $t } from '@/locales';
import AppearanceSettings from '@/layouts/modules/theme-drawer/modules/appearance/index.vue';
import LayoutSettings from '@/layouts/modules/theme-drawer/modules/layout/index.vue';
import GeneralSettings from '@/layouts/modules/theme-drawer/modules/general/index.vue';
import PresetSettings from '@/layouts/modules/theme-drawer/modules/preset/index.vue';
import ConfigOperation from '@/layouts/modules/theme-drawer/modules/config-operation.vue';

defineOptions({ name: 'UserProfile' });

const message = useMessage();
const authStore = useAuthStore();
const loading = ref(false);
const profile = ref<Api.Admin.UserProfile | null>(null);
const activeTab = ref('profile');

const profileFormRef = ref<FormInst | null>(null);
const passwordFormRef = ref<FormInst | null>(null);
const themeTab = ref('appearance');

const profileForm = reactive({
  nickname: ''
});

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

const profileRules: FormRules = {
  nickname: [
    {
      required: true,
      message: $t('page.userProfile.nicknameRequired'),
      trigger: ['blur', 'input']
    }
  ]
};

const passwordRules: FormRules = {
  oldPassword: [
    {
      required: true,
      message: $t('page.userProfile.oldPasswordRequired'),
      trigger: ['blur', 'input']
    }
  ],
  newPassword: [
    {
      required: true,
      message: $t('page.userProfile.newPasswordRequired'),
      trigger: ['blur', 'input']
    },
    {
      min: 6,
      message: $t('page.userProfile.passwordMinLength'),
      trigger: ['blur', 'input']
    }
  ],
  confirmPassword: [
    {
      required: true,
      message: $t('page.userProfile.confirmPasswordRequired'),
      trigger: ['blur', 'input']
    },
    {
      validator(_rule: unknown, value: string) {
        return value === passwordForm.newPassword;
      },
      message: $t('page.userProfile.passwordMismatch'),
      trigger: ['blur', 'input']
    }
  ]
};

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
    message.success($t('page.userProfile.updateSuccess'));
    authStore.userInfo.userName = profile.value?.userName || '';
    await loadProfile();
  }
}

async function handleChangePassword() {
  await passwordFormRef.value?.validate();
  const { error } = await fetchChangePassword({
    oldPassword: passwordForm.oldPassword,
    newPassword: passwordForm.newPassword
  });
  if (!error) {
    message.success($t('page.userProfile.passwordChangeSuccess'));
    authStore.resetStore();
  }
}

// Load profile on mount
loadProfile();
</script>

<template>
  <div class="h-full p-6">
    <NSpin :show="loading">
      <div class="flex gap-6 h-full">
        <!-- 左侧：用户信息卡片 -->
        <NCard class="w-320px flex-shrink-0">
          <div class="flex flex-col items-center py-8">
            <NAvatar :size="100" round class="mb-4">
              <SvgIcon icon="ph:user-circle" class="text-60px" />
            </NAvatar>
            <h2 class="text-24px font-bold mb-2">{{ profile?.nickname || profile?.userName }}</h2>
            <NTag type="info">{{ profile?.role }}</NTag>

            <NDivider />

            <NDescriptions :column="1" label-placement="left" class="w-full px-4">
              <NDescriptionsItem :label="$t('page.userProfile.userName')">
                {{ profile?.userName }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('page.userProfile.nickname')">
                {{ profile?.nickname }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('page.userProfile.lastLoginAt')">
                {{ profile?.lastLoginAt || '-' }}
              </NDescriptionsItem>
              <NDescriptionsItem :label="$t('page.userProfile.createdAt')">
                {{ profile?.createdAt }}
              </NDescriptionsItem>
            </NDescriptions>
          </div>
        </NCard>

        <!-- 右侧：功能区域 -->
        <NCard class="flex-1">
          <NTabs v-model:value="activeTab" type="line" animated>
            <NTabPane name="profile" :tab="$t('page.userProfile.basicInfo')">
              <div class="max-w-600px py-4">
                <NForm
                  ref="profileFormRef"
                  :model="profileForm"
                  :rules="profileRules"
                  label-placement="left"
                  label-width="100"
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
                  <NFormItem :label="$t('page.userProfile.lastLoginAt')">
                    <NInput :value="profile?.lastLoginAt || '-'" disabled />
                  </NFormItem>
                  <NFormItem :label="$t('page.userProfile.createdAt')">
                    <NInput :value="profile?.createdAt" disabled />
                  </NFormItem>
                </NForm>

                <div class="flex justify-start ml-100px mt-4">
                  <NButton type="primary" @click="handleUpdateProfile">
                    <template #icon>
                      <SvgIcon icon="mdi:content-save-outline" />
                    </template>
                    {{ $t('common.save') }}
                  </NButton>
                </div>
              </div>
            </NTabPane>

            <NTabPane name="password" :tab="$t('page.userProfile.changePassword')">
              <div class="max-w-600px py-4">
                <NForm
                  ref="passwordFormRef"
                  :model="passwordForm"
                  :rules="passwordRules"
                  label-placement="left"
                  label-width="100"
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

                <div class="flex justify-start ml-100px mt-4">
                  <NButton type="primary" @click="handleChangePassword">
                    <template #icon>
                      <SvgIcon icon="mdi:lock-check-outline" />
                    </template>
                    {{ $t('page.userProfile.changePassword') }}
                  </NButton>
                </div>
              </div>
            </NTabPane>

            <NTabPane name="theme" :tab="$t('page.userProfile.themeSettings')">
              <div class="py-4">
                <NTabs v-model:value="themeTab" type="segment" size="small" class="mb-16px">
                  <NTab name="appearance" :tab="$t('theme.tabs.appearance')"></NTab>
                  <NTab name="layout" :tab="$t('theme.tabs.layout')"></NTab>
                  <NTab name="general" :tab="$t('theme.tabs.general')"></NTab>
                  <NTab name="preset" :tab="$t('theme.tabs.preset')"></NTab>
                </NTabs>
                <div class="min-h-300px">
                  <KeepAlive>
                    <AppearanceSettings v-if="themeTab === 'appearance'" />
                    <LayoutSettings v-else-if="themeTab === 'layout'" />
                    <GeneralSettings v-else-if="themeTab === 'general'" />
                    <PresetSettings v-else-if="themeTab === 'preset'" />
                  </KeepAlive>
                </div>
                <NDivider />
                <ConfigOperation />
              </div>
            </NTabPane>
          </NTabs>
        </NCard>
      </div>
    </NSpin>
  </div>
</template>

<style scoped></style>
