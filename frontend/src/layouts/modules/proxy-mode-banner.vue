<script setup lang="ts">
import { computed } from 'vue';
import { NButton } from 'naive-ui';
import { useAuthStore } from '@/store/modules/auth';
import { $t } from '@/locales';

defineOptions({
  name: 'ProxyModeBanner'
});

const authStore = useAuthStore();

const isProxyMode = computed(() => authStore.proxyMode);

const userName = computed(() => authStore.userInfo.userName);

function handleExitProxy() {
  authStore.resetStore();
}
</script>

<template>
  <div
    v-if="isProxyMode"
    class="flex items-center justify-between px-4 py-1.5 bg-orange-500 text-white text-sm dark:bg-orange-600"
  >
    <span>
      ⚠️ {{ $t('proxy.desc', { userName }) }}
    </span>
    <NButton size="tiny" quaternary color="#fff" @click="handleExitProxy">
      {{ $t('proxy.exit') }}
    </NButton>
  </div>
</template>
