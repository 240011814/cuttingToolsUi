<script setup lang="ts">
import { ref, watch } from "vue";

defineOptions({
  name: 'ai_custom-training'
});
import { useRoute } from "vue-router";
import TrainingChat from "../components/training-chat.vue";
import { fetchCustomTrainingDetail } from "@/service/api";

const route = useRoute();
const loading = ref(true);
const training = ref<any>(null);
const chatKey = ref(0);

const loadTraining = async () => {
  const id = route.params.id;
  if (id) {
    loading.value = true;
    try {
      const { data } = await fetchCustomTrainingDetail(Number(id));
      if (data) {
        training.value = data;
        chatKey.value++;
      }
    } catch (err: any) {
      console.error("加载训练失败:", err);
    } finally {
      loading.value = false;
    }
  }
};

watch(
  () => route.params.id,
  () => {
    loadTraining();
  },
  { immediate: true }
);
</script>

<template>
  <div v-if="loading" class="h-full flex items-center justify-center">
    <NSpin size="large" />
  </div>
  <TrainingChat
    v-else-if="training"
    :key="chatKey"
    :module-key="'custom_' + training.id"
    :training-type="training.title"
    :custom-training-id="training.id"
    :system-prompt="training.system_prompt"
    :initial-message="
      training.initial_message || '你好！我是你的AI训练助手，让我们开始吧。'
    "
    :input-placeholder="
      training.input_placeholder || '输入消息... (回车发送，Shift + 回车换行)'
    "
    :assistant-color="training.color || '#2080f0'"
    :speech-lang="training.speech_lang || 'zh-CN'"
    :speech-rate="training.speech_rate || 0.95"
  />
  <div v-else class="h-full flex items-center justify-center text-gray-500">
    训练不存在或加载失败
  </div>
</template>
