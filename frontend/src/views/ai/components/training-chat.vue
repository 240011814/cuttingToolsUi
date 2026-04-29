<script setup lang="ts">
import { nextTick, ref } from "vue";
import { useMessage } from "naive-ui";
import { fetchAddVocabulary } from "@/service/api";
import { getAuthorization } from "@/service/request/shared";
import { getServiceBaseURL } from "@/utils/service";

interface VocabSuggestion {
  word?: string;
  phonetic?: string;
  definition?: string;
  example?: string;
  confusingWords?: string;
}

interface ChatMessage {
  role: "user" | "assistant" | "system";
  content: string;
  suggestions?: VocabSuggestion[];
}

const props = withDefaults(
  defineProps<{
    systemPrompt: string;
    initialMessage: string;
    inputPlaceholder?: string;
    assistantColor?: string;
    enableVocabulary?: boolean;
    speechLang?: string;
    speechRate?: number;
  }>(),
  {
    inputPlaceholder: "输入消息... (回车发送，Shift + 回车换行)",
    assistantColor: "#2080f0",
    enableVocabulary: false,
    speechLang: "en-US",
    speechRate: 0.9,
  }
);

const systemMessage: ChatMessage = {
  role: "system",
  content: props.systemPrompt,
};

const messages = ref<ChatMessage[]>([
  { role: "assistant", content: props.initialMessage },
]);
const inputMessage = ref("");
const isGenerating = ref(false);
const scrollbarRef = ref<any>(null);
const message = useMessage();

const showVocabModal = ref(false);
const vocabLoading = ref(false);
const vocabForm = ref({
  word: "",
  phonetic: "",
  definition: "",
  example: "",
  confusingWords: "",
});

const resetVocabForm = () => {
  vocabForm.value = {
    word: "",
    phonetic: "",
    definition: "",
    example: "",
    confusingWords: "",
  };
};

const openVocabModal = () => {
  resetVocabForm();
  showVocabModal.value = true;
};

const handleSelectText = () => {
  if (!props.enableVocabulary) return;

  const selection = window.getSelection()?.toString().trim();
  if (selection && selection.length > 0 && selection.length < 50) {
    vocabForm.value.word = selection;
  }
};

const submitVocab = async () => {
  if (!vocabForm.value.word.trim()) {
    message.warning("请输入单词");
    return;
  }

  vocabLoading.value = true;
  try {
    await fetchAddVocabulary(vocabForm.value);
    message.success("已添加到生词本");
    showVocabModal.value = false;
  } catch (err: any) {
    message.error(`添加失败: ${err?.message || "未知错误"}`);
  } finally {
    vocabLoading.value = false;
  }
};

const scrollToBottom = async () => {
  await nextTick();
  scrollbarRef.value?.scrollTo({ position: "bottom", behavior: "smooth" });
};

const appendAssistantContent = (content: string) => {
  const lastIdx = messages.value.length - 1;
  messages.value[lastIdx] = {
    ...messages.value[lastIdx],
    content: messages.value[lastIdx].content + content,
  };
};

const setAssistantError = (content: string) => {
  const lastIdx = messages.value.length - 1;
  messages.value[lastIdx] = {
    ...messages.value[lastIdx],
    content,
  };
};

const parseVocabSuggestions = () => {
  if (!props.enableVocabulary) return;

  const lastIdx = messages.value.length - 1;
  const lastMsg = messages.value[lastIdx];
  if (lastMsg.role !== "assistant") return;

  const match = lastMsg.content.match(/<vocabs>([\s\S]*?)<\/vocabs>/);
  if (!match?.[1]) return;

  try {
    const vocabs = JSON.parse(match[1]);
    if (Array.isArray(vocabs) && vocabs.length > 0) {
      messages.value[lastIdx] = {
        ...lastMsg,
        suggestions: vocabs,
      };
    }
  } catch (error) {
    console.error("Failed to parse suggested vocabs:", error);
  }
};

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isGenerating.value) return;

  const userText = inputMessage.value;
  inputMessage.value = "";

  messages.value.push({ role: "user", content: userText });
  messages.value.push({ role: "assistant", content: "" });

  scrollToBottom();
  isGenerating.value = true;

  const isHttpProxy = import.meta.env.DEV && import.meta.env.VITE_HTTP_PROXY === "Y";
  const { baseURL } = getServiceBaseURL(import.meta.env, isHttpProxy);

  try {
    const history = messages.value.slice(0, -1).filter((item) => item.content.trim());
    const response = await fetch(`${baseURL}/api/chat`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: getAuthorization() || "",
      },
      body: JSON.stringify({
        messages: [systemMessage, ...history],
      }),
    });

    if (!response.ok) {
      let errorMessage = `请求失败 (HTTP ${response.status})`;
      try {
        const errorData = await response.json();
        if (errorData.error) errorMessage = errorData.error;
      } catch {}
      throw new Error(errorMessage);
    }

    const reader = response.body?.getReader();
    const decoder = new TextDecoder("utf-8");
    if (!reader) throw new Error("无法获取响应流");

    let buffer = "";
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split(/\r?\n/);
      buffer = lines.pop() || "";

      for (let line of lines) {
        line = line.trim();
        if (!line || !line.startsWith("data:")) continue;

        const dataStr = line.replace(/^data:\s*/, "").trim();
        if (dataStr === "[DONE]" || dataStr === '"[DONE]"') break;

        try {
          const dataObj = JSON.parse(dataStr);

          if (dataObj.error) {
            setAssistantError(`AI 服务错误: ${dataObj.error}`);
            break;
          }

          if (dataObj.content) {
            appendAssistantContent(dataObj.content);
            scrollToBottom();
          }
        } catch {
          // Ignore incomplete streaming JSON chunks.
        }
      }
    }
  } catch (err: any) {
    setAssistantError(
      `连接 AI 服务失败: ${
        err?.message || "未知错误"
      }。\n请确认后端服务正常运行且已正确配置 DEEPSEEK_API_KEY。`
    );
  } finally {
    isGenerating.value = false;
    await scrollToBottom();
    parseVocabSuggestions();
  }
};

const handleEnter = (event: KeyboardEvent) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    sendMessage();
  }
};

const handleApplySuggestion = (vocab: VocabSuggestion) => {
  vocabForm.value = {
    word: vocab.word || "",
    phonetic: vocab.phonetic || "",
    definition: vocab.definition || "",
    example: vocab.example || "",
    confusingWords: vocab.confusingWords || "",
  };
  showVocabModal.value = true;
};

const formatDisplayContent = (content: string) => {
  if (!props.enableVocabulary) return content;

  return content.replace(/<vocabs>[\s\S]*?<\/vocabs>/g, "").trim();
};

const handlePlay = (text: string) => {
  if (!window.speechSynthesis) {
    message.error("您的浏览器不支持语音播放");
    return;
  }

  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = props.speechLang;
  utterance.rate = props.speechRate;
  window.speechSynthesis.speak(utterance);
};
</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4 overflow-hidden">
    <NCard
      class="flex-1 overflow-hidden"
      content-class="p-0 flex flex-col h-full"
      :bordered="false"
      shadow="sm"
    >
      <NScrollbar ref="scrollbarRef" class="flex-1 p-4 bg-gray-50/50 dark:bg-dark">
        <div class="flex flex-col gap-6 pb-4">
          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="flex items-start gap-3"
            :class="msg.role === 'user' ? 'flex-row-reverse' : 'flex-row'"
          >
            <NAvatar
              :color="msg.role === 'user' ? '#18a058' : assistantColor"
              round
              size="large"
            >
              {{ msg.role === "user" ? "U" : "AI" }}
            </NAvatar>
            <div class="flex flex-col gap-1 max-w-[80%]">
              <div
                class="p-4 rounded-2xl whitespace-pre-wrap leading-relaxed shadow-sm text-[15px] relative group"
                :class="
                  msg.role === 'user'
                    ? 'bg-[#18a058] text-white rounded-tr-none'
                    : 'bg-white text-gray-800 rounded-tl-none dark:bg-gray-800 dark:text-gray-200'
                "
                @mouseup="msg.role === 'assistant' ? handleSelectText() : undefined"
              >
                {{ formatDisplayContent(msg.content) }}
                <span
                  v-if="
                    isGenerating && index === messages.length - 1 && msg.content === ''
                  "
                  class="inline-block mt-1"
                >
                  <NSpin size="small" />
                </span>

                <div
                  v-if="msg.role === 'assistant' && msg.content"
                  class="absolute -right-2 -top-4 opacity-0 group-hover:opacity-100 transition-all duration-300 flex gap-1 p-1 bg-white dark:bg-gray-700 rounded-full shadow-md border border-gray-100 dark:border-gray-600 z-10"
                >
                  <template v-if="enableVocabulary">
                    <div class="w-[1px] h-4 bg-gray-200 dark:bg-gray-600 self-center" />
                    <ButtonIcon
                      icon="mdi:star"
                      class="text-18px text-warning"
                      tooltip-content="添加到生词本"
                      @click.stop="openVocabModal()"
                    />
                  </template>
                </div>
              </div>

              <div
                v-if="enableVocabulary && msg.suggestions?.length"
                class="flex flex-wrap gap-2 mt-2"
              >
                <span class="text-xs text-gray-400 self-center">智能建议:</span>
                <NTag
                  v-for="(vocab, vocabIndex) in msg.suggestions"
                  :key="vocabIndex"
                  size="small"
                  round
                  type="info"
                  check-strategy="child"
                  class="cursor-pointer hover:shadow-sm transition-shadow"
                  @click="handleApplySuggestion(vocab)"
                >
                  <template #icon>
                    <div class="i-mdi:plus" />
                  </template>
                  {{ vocab.word }}
                </NTag>
              </div>
            </div>
          </div>
        </div>
      </NScrollbar>

      <div
        class="p-4 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-dark flex items-stretch gap-4"
      >
        <NInput
          v-model:value="inputMessage"
          type="textarea"
          :autosize="{ minRows: 2, maxRows: 6 }"
          :placeholder="inputPlaceholder"
          class="flex-1 shadow-sm"
          size="large"
          @keydown="handleEnter"
        />
        <NButton
          type="primary"
          size="large"
          :loading="isGenerating"
          :disabled="!inputMessage.trim()"
          class="px-8 rounded-lg"
          style="align-self: stretch; height: auto"
          @click="sendMessage"
        >
          发送
        </NButton>
      </div>
    </NCard>

    <NModal
      v-if="enableVocabulary"
      v-model:show="showVocabModal"
      preset="card"
      title="添加到生词本"
      class="max-w-md"
      :segmented="{ content: 'soft' }"
    >
      <NForm :model="vocabForm" label-placement="left" label-width="80">
        <NFormItem label="单词" path="word">
          <div class="flex gap-2 w-full">
            <NInput
              v-model:value="vocabForm.word"
              placeholder="输入单词"
              class="flex-1"
            />
            <ButtonIcon
              icon="mdi:volume-high"
              class="text-20px text-primary"
              @click="handlePlay(vocabForm.word)"
            />
          </div>
        </NFormItem>
        <NFormItem label="音标" path="phonetic">
          <NInput v-model:value="vocabForm.phonetic" placeholder="输入音标 (可选)" />
        </NFormItem>
        <NFormItem label="释义" path="definition">
          <NInput
            v-model:value="vocabForm.definition"
            type="textarea"
            :autosize="{ minRows: 2 }"
            placeholder="输入中文释义"
          />
        </NFormItem>
        <NFormItem label="例句" path="example">
          <NInput
            v-model:value="vocabForm.example"
            type="textarea"
            :autosize="{ minRows: 2 }"
            placeholder="输入英文例句及翻译，如：I love coffee. (我爱咖啡。)"
          />
        </NFormItem>
        <NFormItem label="易混淆" path="confusingWords">
          <NInput
            v-model:value="vocabForm.confusingWords"
            type="textarea"
            :autosize="{ minRows: 2 }"
            placeholder="输入易混淆单词及翻译，如：Shook (摇动)"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showVocabModal = false">取消</NButton>
          <NButton type="primary" :loading="vocabLoading" @click="submitVocab"
            >确认添加</NButton
          >
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.flex-col > .flex {
  animation: slideIn 0.3s ease-out forwards;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
