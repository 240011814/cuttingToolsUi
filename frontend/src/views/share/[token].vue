<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { NCard, NResult, NAvatar, NTag, NSpin } from "naive-ui";
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
    <div class="max-w-4xl mx-auto py-8 px-4">
      <NSpin :show="loading">
        <template v-if="error">
          <NResult status="404" :title="$t('page.share.notFound')" class="mt-20" />
        </template>
        <template v-else-if="history">
          <NCard :bordered="false" shadow="sm">
            <template #header>
              <div class="flex items-center gap-3">
                <span class="text-lg font-bold">{{ history.title }}</span>
                <NTag type="info" :bordered="false" size="small">
                  {{ typeMap[history.training_type] || history.training_type }}
                </NTag>
              </div>
            </template>
            <div class="flex flex-col gap-4">
              <div
                v-for="(msg, index) in history.messages"
                :key="index"
                class="flex flex-col gap-2"
              >
                <template v-if="msg.role !== 'system'">
                  <div
                    class="flex items-start gap-3"
                    :class="msg.role === 'user' ? 'flex-row-reverse' : 'flex-row'"
                  >
                    <NAvatar
                      :color="msg.role === 'user' ? '#18a058' : '#2080f0'"
                      round
                      size="small"
                      class="flex-shrink-0"
                    >
                      {{ msg.role === "user" ? "U" : "AI" }}
                    </NAvatar>
                    <div class="flex flex-col gap-1 max-w-[80%]">
                      <div
                        class="p-3 rounded-2xl whitespace-pre-wrap leading-relaxed shadow-sm text-sm relative"
                        :class="
                          msg.role === 'user'
                            ? 'bg-[#18a058] text-white rounded-tr-none'
                            : 'bg-gray-100 text-gray-800 rounded-tl-none dark:bg-gray-800 dark:text-gray-200'
                        "
                      >
                        <!-- eslint-disable-next-line vue/no-v-html -->
                        <div v-html="renderMarkdown(msg.content)"></div>
                      </div>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </NCard>
        </template>
      </NSpin>
    </div>
  </div>
</template>

<style scoped></style>
