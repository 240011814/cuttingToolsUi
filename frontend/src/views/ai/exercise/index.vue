<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

const goBack = () => {
  router.push({ name: 'ai_vocabulary' });
};
</script>

<template>
  <div class="h-full flex items-center justify-center p-4">
    <n-result
      status="info"
      title="练习功能开发中"
      description="生词强化练习模块即将上线，敬请期待！"
    >
      <template #footer>
        <n-button type="primary" @click="goBack">返回生词本</n-button>
      </template>
    </n-result>
  </div>
</template>

<style scoped></style>
