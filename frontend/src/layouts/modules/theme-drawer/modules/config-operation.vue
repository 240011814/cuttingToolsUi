<script setup lang="ts">
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';

defineOptions({
  name: 'ConfigOperation'
});

const themeStore = useThemeStore();

async function handleSave() {
  const ok = await themeStore.saveToServerImmediate();
  if (ok) {
    window.$message?.success($t('theme.configOperation.saveSuccessMsg'));
  }
}

async function handleReset() {
  await themeStore.resetStore();
  window.$message?.success($t('theme.configOperation.resetSuccessMsg'));
}
</script>

<template>
  <div class="w-full flex justify-between">
    <NButton type="error" ghost @click="handleReset">{{ $t('theme.configOperation.resetConfig') }}</NButton>
    <NButton type="primary" @click="handleSave">{{ $t('theme.configOperation.saveConfig') }}</NButton>
  </div>
</template>

<style scoped></style>
