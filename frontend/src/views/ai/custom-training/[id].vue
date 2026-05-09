<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import TrainingChat from '../components/training-chat.vue';
import { fetchCustomTrainingDetail } from '@/service/api';

const route = useRoute();
const loading = ref(true);
const training = ref<any>(null);

const loadTraining = async () => {
  const id = route.params.id;
  if (id) {
    try {
      const { data } = await fetchCustomTrainingDetail(Number(id));
      if (data) {
        training.value = data;
      }
    } catch (err: any) {
      console.error('加载训练失败:', err);
    } finally {
      loading.value = false;
    }
  }
};

onMounted(() => {
  loadTraining();
});
</script>

<template>
  <div v-if="loading" class="h-full flex items-center justify-center">
    <NSpin size="large" />
  </div>
  <TrainingChat
    v-else-if="training"
    :module-key="'custom_' + training.id"
    :training-type="training.title"
    :system-prompt="training.system_prompt"
    :initial-message="training.initial_message || '你好！我是你的AI训练助手，让我们开始吧。'"
    :input-placeholder="training.input_placeholder || '输入消息... (回车发送，Shift + 回车换行)'"
    :assistant-color="training.color || '#2080f0'"
    :speech-lang="training.speech_lang || 'zh-CN'"
    :speech-rate="training.speech_rate || 0.95"
  />
  <div v-else class="h-full flex items-center justify-center text-gray-500">
    训练不存在或加载失败
  </div>
</template>
