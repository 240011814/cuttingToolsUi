<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { NResult, NTag, NSpin } from "naive-ui";
import { fetchSharedHistory } from "@/service/api";
import type { TrainingHistory } from "@/service/api";
import { $t } from "@/locales";
import MarkdownIt from "markdown-it";

const route = useRoute();
const token = route.params.token as string;

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
});

const renderMarkdown = (content: string) => {
  return md.render(content || "");
};

const loading = ref(true);
const error = ref(false);
const history = ref<TrainingHistory | null>(null);

const typeMap: Record<string, string> = {
  ai_chat: $t("route.ai_chat"),
  ai_decision: $t("route.ai_decision"),
  ai_social: $t("route.ai_social"),
  ai_emergency: $t("route.ai_emergency")
};

const loadHistory = async () => {
  loading.value = true;
  error.value = false;
  try {
    const { data } = await fetchSharedHistory(token);
    if (data) {
      history.value = data;
    } else {
      error.value = true;
    }
  } catch {
    error.value = true;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadHistory();
});
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="max-w-3xl mx-auto py-6 px-4">
      <NSpin :show="loading">
        <template v-if="error">
          <NResult status="404" :title="$t('page.share.notFound')" class="mt-20" />
        </template>
        <template v-else-if="history">
          <!-- Header -->
          <div class="mb-6">
            <h1 class="text-xl font-bold text-gray-800 dark:text-gray-100 mb-2">{{ history.title }}</h1>
            <NTag type="info" :bordered="false" size="small">
              {{ typeMap[history.training_type] || history.training_type }}
            </NTag>
          </div>

          <!-- Messages -->
          <div class="flex flex-col gap-6">
            <template v-for="(msg, index) in history.messages" :key="index">
              <div v-if="msg.role !== 'system'" class="flex flex-col gap-2">
                <!-- User: avatar right, bubble right -->
                <template v-if="msg.role === 'user'">
                  <div class="flex justify-end">
                    <div class="w-8 h-8 rounded-full bg-[#18a058] flex items-center justify-center text-white text-xs font-bold">
                      U
                    </div>
                  </div>
                  <div class="flex justify-end">
                    <div
                      class="p-3 rounded-2xl rounded-tr-none whitespace-pre-wrap leading-relaxed shadow-sm text-sm bg-[#18a058] text-white"
                    >
                      <!-- eslint-disable-next-line vue/no-v-html -->
                      <div v-html="renderMarkdown(msg.content)"></div>
                    </div>
                  </div>
                </template>
                <!-- AI: avatar left, bubble left -->
                <template v-else>
                  <div class="flex justify-start">
                    <div class="w-8 h-8 rounded-full bg-[#2080f0] flex items-center justify-center text-white text-xs font-bold">
                      AI
                    </div>
                  </div>
                  <div class="flex justify-start">
                    <div
                      class="p-3 rounded-2xl rounded-tl-none whitespace-pre-wrap leading-relaxed shadow-sm text-sm bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200"
                    >
                      <!-- eslint-disable-next-line vue/no-v-html -->
                      <div v-html="renderMarkdown(msg.content)"></div>
                    </div>
                  </div>
                </template>
              </div>
            </template>
          </div>
        </template>
      </NSpin>
    </div>
  </div>
</template>

<style scoped></style>
