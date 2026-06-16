<script setup lang="ts">
import { nextTick, ref, onBeforeUnmount, onMounted } from "vue";
import { useMessage, NDrawer, NDrawerContent, NModal, NInput } from "naive-ui";
import { useAppStore } from "@/store/modules/app";
import {
  fetchAddVocabulary,
  fetchAddNote,
  fetchHistoryDetail,
  fetchUpdateFavorite,
  fetchUpdateHistoryTitle,
  fetchGenerateShareToken,
} from "@/service/api";
import { fetchGetAIModels, fetchGetUserPrompt, fetchChatStream } from "@/service/api/ai";
import { fetchSearchMemories } from "@/service/api/memory";
import { useAuth } from "@/hooks/business/auth";
import MarkdownIt from "markdown-it";
import texmath from "markdown-it-texmath";
import katex from "katex";
import "katex/dist/katex.min.css";
import { useRoute } from "vue-router";
import PromptEditor from "./prompt-editor.vue";

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
  renderedContent?: string;
  suggestions?: VocabSuggestion[];
  isError?: boolean;
}

const props = withDefaults(
  defineProps<{
    systemPrompt: string;
    initialMessage: string;
    moduleKey: string;
    trainingType?: string;
    customTrainingId?: number | null;
    inputPlaceholder?: string;
    assistantColor?: string;
    enableVocabulary?: boolean;
    speechLang?: string;
    speechRate?: number;
  }>(),
  {
    trainingType: "",
    customTrainingId: null,
    inputPlaceholder: "输入消息... (回车发送，Shift + 回车换行)",
    assistantColor: "#2080f0",
    enableVocabulary: false,
    speechLang: "en-US",
    speechRate: 0.9,
  }
);

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
}).use(texmath, { engine: katex, delimiters: 'dollars' });
const { hasAuth } = useAuth();
const appStore = useAppStore();

const renderMarkdown = (content: string) => {
  return md.render(content).trim();
};

function formatDisplayContent(content: string) {
  if (!props.enableVocabulary) return content;

  return content.replace(/<vocabs>[\s\S]*?<\/vocabs>/g, "").trim();
}

const renderMessageContent = (content: string) => {
  return renderMarkdown(formatDisplayContent(content));
};

const createChatMessage = (role: ChatMessage["role"], content: string): ChatMessage => {
  return {
    role,
    content,
    renderedContent: renderMessageContent(content),
  };
};

const toApiMessages = (items: ChatMessage[]) => {
  return items.map(({ role, content }) => ({ role, content }));
};

const systemMessage = ref<ChatMessage>({
  role: "system",
  content: props.systemPrompt,
});

const messages = ref<ChatMessage[]>([
  createChatMessage("assistant", props.initialMessage),
]);

const showPromptEditor = ref(false);
const memorySearchQuery = ref("用户已经训练过的场景和学习进度和用户的偏好");
const memorySearchTopK = ref(30);
const mem0Enabled = ref(true);

async function refreshPrompt() {
  const { data } = await fetchGetUserPrompt(props.moduleKey);
  if (data) {
    systemMessage.value.content = data.effective_prompt || props.systemPrompt;
    if (data.memory_search_query) {
      memorySearchQuery.value = data.memory_search_query;
    }
    if (data.memory_search_top_k) {
      memorySearchTopK.value = data.memory_search_top_k;
    }
    mem0Enabled.value = data.mem0_enabled !== false;
  }
}

async function loadMemories() {
  if (!mem0Enabled.value) return;
  try {
    const query = memorySearchQuery.value?.trim() || "用户已经训练过的场景和学习进度和用户的偏好";
    const { data: memories } = await fetchSearchMemories(
      query,
      memorySearchTopK.value
    );
    if (memories && memories.length > 0) {
      const memoryText = memories.map((m: any) => `- ${m.memory}`).join("\n");
      systemMessage.value = {
        role: "system",
        content: `${systemMessage.value.content}\n\n---\n以下是基于「${query}」检索到的用户记忆，回答时请参考：\n${memoryText}\n---`,
      };
    }
  } catch {
    // Memory load failure should not block chat
  }
}
const historyId = ref<number>(0);
const isFavorite = ref(false);
const shareToken = ref<string | null>(null);
const historyTitle = ref("");
const showTitleModal = ref(false);
const editTitle = ref("");
const inputMessage = ref("");
const isGenerating = ref(false);
const scrollbarRef = ref<any>(null);
const message = useMessage();
let scrollFrame = 0;

const modelOptions = ref<{ label: string; value: string }[]>([]);
const selectedModel = ref("");

async function loadModels() {
  const { data } = await fetchGetAIModels();
  if (data && data.length > 0) {
    modelOptions.value = data.map((m) => ({
      label: m.display_name,
      value: m.model_code,
    }));
    // 默认选择第一个模型（或标记为  default的模型）
    const defaultModel = data.find((m) => m.is_default) || data[0];
    selectedModel.value = defaultModel.model_code;
  }
}

const showVocabModal = ref(false);
const vocabLoading = ref(false);
const vocabForm = ref({
  word: "",
  phonetic: "",
  definition: "",
  example: "",
  confusingWords: "",
});

const showNoteModal = ref(false);
const noteLoading = ref(false);
const noteForm = ref({
  title: "",
  category: "",
  content: "",
});

const route = useRoute();
const routeTitleMap: Record<string, string> = {
  ai_chat: "英语训练",
  ai_decision: "决策训练",
  ai_social: "社交训练",
  ai_emergency: "应急训练",
  ai_exercise: "练习",
};

const openNoteModal = (content: string) => {
  const routeName = (route.name as string) || "";
  const defaultCategory = (routeTitleMap[routeName] || "未分类").replace("训练", "");

  // 提取前20个字符作为默认标题
  let defaultTitle = content.trim().slice(0, 20);
  if (content.trim().length > 20) defaultTitle += "...";

  noteForm.value = {
    title: defaultTitle,
    category: defaultCategory,
    content: formatDisplayContent(content),
  };
  showNoteModal.value = true;
};

const submitNote = async () => {
  if (!noteForm.value.title.trim()) {
    message.warning("请输入标题");
    return;
  }
  if (!noteForm.value.category.trim()) {
    message.warning("请输入分类");
    return;
  }
  if (!noteForm.value.content.trim()) {
    message.warning("请输入内容");
    return;
  }

  noteLoading.value = true;
  try {
    await fetchAddNote(noteForm.value);
    message.success("笔记添加成功");
    showNoteModal.value = false;
  } catch (err: any) {
    message.error(`添加失败: ${err?.message || "未知错误 "}`);
  } finally {
    noteLoading.value = false;
  }
};

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

const scrollToBottom = async (behavior: ScrollBehavior = "auto") => {
  await nextTick();
  scrollbarRef.value?.scrollTo({ position: "bottom", behavior });
};

const scheduleScrollToBottom = () => {
  if (scrollFrame) return;

  scrollFrame = window.requestAnimationFrame(() => {
    scrollFrame = 0;
    scrollToBottom();
  });
};

const appendAssistantContent = (content: string) => {
  const lastIdx = messages.value.length - 1;
  const nextContent = messages.value[lastIdx].content + content;
  messages.value[lastIdx] = {
    ...messages.value[lastIdx],
    content: nextContent,
    renderedContent: renderMessageContent(nextContent),
  };
};

const setAssistantError = (content: string) => {
  const lastIdx = messages.value.length - 1;
  messages.value[lastIdx] = {
    ...messages.value[lastIdx],
    content,
    renderedContent: renderMessageContent(content),
    isError: true,
  };
};

const copyToClipboard = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content);
    message.success("已复制");
  } catch {
    message.error("复制失败");
  }
};

const parseVocabSuggestions = () => {
  const lastIdx = messages.value.length - 1;
  const lastMsg = messages.value[lastIdx];
  if (lastMsg.role !== "assistant") return;

  const nextMessage: ChatMessage = {
    ...lastMsg,
    renderedContent: renderMessageContent(lastMsg.content),
  };

  if (!props.enableVocabulary) {
    messages.value[lastIdx] = nextMessage;
    return;
  }

  const match = lastMsg.content.match(/<vocabs>([\s\S]*?)<\/vocabs>/);
  if (!match?.[1]) {
    messages.value[lastIdx] = nextMessage;
    return;
  }

  try {
    const vocabs = JSON.parse(match[1]);
    if (Array.isArray(vocabs) && vocabs.length > 0) {
      messages.value[lastIdx] = {
        ...nextMessage,
        suggestions: vocabs,
      };
      return;
    }
  } catch (error) {
    console.error("Failed to parse suggested vocabs:", error);
  }

  messages.value[lastIdx] = nextMessage;
};

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isGenerating.value) return;

  const userText = inputMessage.value;
  inputMessage.value = "";

  messages.value.push(createChatMessage("user", userText));
  messages.value.push(createChatMessage("assistant", ""));

  scrollToBottom();
  isGenerating.value = true;

  try {
    const history = messages.value
      .slice(0, -1)
      .filter((item) => item.content.trim() && !item.isError);
    const routeName = props.trainingType || (route.name as string) || "ai_chat";

    const response = await fetchChatStream({
      history_id: historyId.value,
      training_type: routeName,
      custom_training_id: props.customTrainingId || undefined,
      model: selectedModel.value,
      messages: toApiMessages([systemMessage.value, ...history]),
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
    let eventType = "message"; // 1. 移到循环外，持久化状态

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split(/\r?\n/);
      buffer = lines.pop() || "";

      for (let line of lines) {
        line = line.trim();
        if (!line) continue;

        if (line.startsWith("event:")) {
          eventType = line.replace(/^event:\s*/, "").trim();
          continue;
        }

        if (line.startsWith("data:")) {
          const dataStr = line.replace(/^data:\s*/, "").trim();

          try {
            const dataObj = JSON.parse(dataStr);

            if (eventType == "history_id") {
              historyId.value = dataObj.history_id;
              if (dataObj.title) {
                historyTitle.value = dataObj.title;
              }
            }

            if (dataObj.error) {
              setAssistantError(`AI 服务错误: ${dataObj.error}`);
              return;
            }

            if (dataObj.content) {
              appendAssistantContent(dataObj.content);
              scheduleScrollToBottom();
            }
          } catch (e) {
            console.warn("Parse error:", e);
          }
        }
      }
    }
  } catch (err: any) {
    setAssistantError(`连接 AI 服务失败: ${err?.message || "未知错误"}。`);
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

const loadHistory = async (id: number) => {
  try {
    const { data } = await fetchHistoryDetail(id);
    if (data) {
      historyId.value = data.id;
      historyTitle.value = data.title;
      messages.value = (data.messages || [])
        .filter((msg: any) => msg.role !== "system")
        .map((msg: any) => createChatMessage(msg.role, msg.content));
      isFavorite.value = data.is_favorite;
      shareToken.value = data.share_token || null;
    }
  } catch (err: any) {
    message.error(`加载历史记录失败: ${err?.message || "未知错误"}`);
  }
};

const handleToggleFavorite = async () => {
  if (!historyId.value) {
    message.warning("请先发送消息后再收藏");
    return;
  }
  try {
    await fetchUpdateFavorite(historyId.value, !isFavorite.value);
    isFavorite.value = !isFavorite.value;
    message.success(isFavorite.value ? "已收藏" : "已取消收藏");
  } catch (err: any) {
    message.error(`操作失败: ${err?.message || "未知错误"}`);
  }
};

const getShareUrl = (token: string) => {
  const isHashMode = import.meta.env.VITE_ROUTER_HISTORY_MODE === 'hash';
  const basePath = isHashMode ? '/#/' : '/';
  return `${window.location.origin}${basePath}share/${token}`;
};

const handleShare = async () => {
  if (!historyId.value) {
    message.warning("请先发送消息后再分享");
    return;
  }
  try {
    if (shareToken.value) {
      const shareUrl = getShareUrl(shareToken.value);
      await navigator.clipboard.writeText(shareUrl);
      message.success("分享链接已复制到剪贴板");
    } else {
      const { data } = await fetchGenerateShareToken(historyId.value);
      if (data?.share_token) {
        shareToken.value = data.share_token;
        const shareUrl = getShareUrl(data.share_token);
        await navigator.clipboard.writeText(shareUrl);
        message.success("分享链接已复制到剪贴板");
      }
    }
  } catch (err: any) {
    message.error(`操作失败: ${err?.message || "未知错误"}`);
  }
};

const handleOpenEditTitle = () => {
  if (!historyId.value) {
    message.warning("请先发送消息后再编辑标题");
    return;
  }
  editTitle.value = historyTitle.value;
  showTitleModal.value = true;
};

const handleSaveTitle = async () => {
  if (!editTitle.value.trim()) {
    message.warning("标题不能为空");
    return;
  }
  try {
    await fetchUpdateHistoryTitle(historyId.value, editTitle.value.trim());
    historyTitle.value = editTitle.value.trim();
    showTitleModal.value = false;
    message.success("标题已更新");
  } catch (err: any) {
    message.error(`操作失败: ${err?.message || "未知错误"}`);
  }
};

onMounted(() => {
  loadModels();
  refreshPrompt().then(() => loadMemories());

  const queryHistoryId = route.query.history_id;
  if (queryHistoryId) {
    loadHistory(Number(queryHistoryId));
  }

  if (scrollbarRef.value) {
    scrollbarRef.value.scrollTo({ top: 999999 });
  }
});

onBeforeUnmount(() => {
  if (scrollFrame) {
    window.cancelAnimationFrame(scrollFrame);
    scrollFrame = 0;
  }
});
</script>

<template>
  <div class="h-full flex-col flex overflow-hidden" :class="appStore.isMobile ? 'p-1 gap-1' : 'p-4 gap-4'">
    <NCard
      class="flex-1 overflow-hidden"
      content-class="p-0 flex flex-col overflow-hidden"
      :bordered="false"
      :shadow="appStore.isMobile ? false : 'sm'"
    >
      <!-- Header inside content to avoid NCard layout issues in flex-1 -->
      <div
        class="border-b border-gray-100 dark:border-gray-800 flex items-center justify-between"
        :class="appStore.isMobile ? 'px-2 py-1.5 flex-wrap gap-1' : 'p-4'"
      >
        <div class="flex items-center gap-1 min-w-0 flex-1">
          <span class="font-bold text-gray-600 dark:text-gray-300 truncate text-sm">{{
            historyTitle || "AI 训练对话"
          }}</span>
          <NButton v-if="!appStore.isMobile" quaternary size="small" @click="handleOpenEditTitle">
            <template #icon>
              <SvgIcon icon="mdi:pencil-outline" class="text-lg" />
            </template>
          </NButton>
          <NButton
            quaternary
            size="small"
            :type="isFavorite ? 'warning' : 'default'"
            @click="handleToggleFavorite"
          >
            <template #icon>
              <SvgIcon
                :icon="isFavorite ? 'mdi:star' : 'mdi:star-outline'"
                class="text-lg"
              />
            </template>
          </NButton>
          <NButton
            quaternary
            size="small"
            :type="shareToken ? 'info' : 'default'"
            @click="handleShare"
          >
            <template #icon>
              <SvgIcon icon="mdi:share-variant" class="text-lg" />
            </template>
          </NButton>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <NButton
            v-if="hasAuth('ai:prompt:manage') && !appStore.isMobile"
            type="primary"
            size="small"
            class="rounded-lg shadow-sm"
            @click="showPromptEditor = true"
          >
            <div class="flex items-center gap-1 px-1">
              <div class="i-mdi:cog-outline" />
              <span v-if="!appStore.isMobile">设置</span>
            </div>
          </NButton>
          <NSelect
            v-model:value="selectedModel"
            :options="modelOptions"
            size="small"
            class="shadow-sm"
            :style="{ width: appStore.isMobile ? '110px' : '140px' }"
          />
        </div>
      </div>

      <NScrollbar ref="scrollbarRef" class="flex-1 bg-gray-50/50 dark:bg-dark" :class="appStore.isMobile ? 'px-1.5 py-1' : 'p-4'">
        <div class="flex flex-col pb-4" :class="appStore.isMobile ? 'gap-4' : 'gap-6'">
          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="flex items-start"
            :class="[msg.role === 'user' ? 'flex-row-reverse' : 'flex-row', appStore.isMobile ? 'gap-2' : 'gap-3']"
          >
            <!-- PC: 水平布局（头像+气泡并排） -->
            <template v-if="!appStore.isMobile">
              <NAvatar
                :color="msg.role === 'user' ? '#6bb8e8' : assistantColor"
                round
                size="large"
                class="shrink-0 self-start"
              >
                {{ msg.role === "user" ? "U" : "AI" }}
              </NAvatar>
              <div class="flex flex-col gap-1 max-w-[80%]">
                <div class="group/btn">
                  <div
                    class="p-4 text-[15px] rounded-2xl whitespace-pre-wrap leading-relaxed shadow-sm"
                    :class="
                      msg.role === 'user'
                        ? 'bg-[#e8f4fd] text-gray-800 rounded-tr-none dark:bg-blue-900/40 dark:text-gray-200'
                        : 'bg-white text-gray-800 rounded-tl-none dark:bg-gray-800 dark:text-gray-200'
                    "
                    @mouseup="msg.role === 'assistant' ? handleSelectText() : undefined"
                  >
                    <!-- eslint-disable-next-line vue/no-v-html -->
                    <div class="msg-content" v-html="msg.renderedContent"></div>
                    <span
                      v-if="
                        isGenerating && index === messages.length - 1 && msg.content === ''
                      "
                      class="inline-block mt-1"
                    >
                      <NSpin size="small" />
                    </span>
                  </div>

                  <div
                    v-if="msg.content"
                    class="flex items-center gap-0.5 mt-1 justify-end opacity-0 group-hover/btn:opacity-100 transition-all duration-200"
                  >
                    <ButtonIcon
                      icon="mdi:content-copy"
                      class="!h-28px !w-28px text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
                      tooltip-content="复制"
                      @click.stop="copyToClipboard(msg.content)"
                    />
                    <ButtonIcon
                      v-if="enableVocabulary && msg.role === 'assistant'"
                      icon="mdi:star-outline"
                      class="!h-28px !w-28px text-gray-400 hover:text-amber-500 dark:text-gray-500 dark:hover:text-amber-400"
                      tooltip-content="添加到生词本"
                      @click.stop="openVocabModal()"
                    />
                    <ButtonIcon
                      v-if="msg.role === 'assistant'"
                      icon="mdi:notebook-edit-outline"
                      class="!h-28px !w-28px text-gray-400 hover:text-blue-500 dark:text-gray-500 dark:hover:text-blue-400"
                      tooltip-content="添加笔记"
                      @click.stop="openNoteModal(msg.content)"
                    />
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
            </template>

            <!-- 移动端: 垂直布局（头像独占一行，气泡占满宽度） -->
            <template v-else>
              <div class="flex flex-col gap-2 w-full">
                <div class="flex" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
                  <div
                    class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0"
                    :style="{ backgroundColor: msg.role === 'user' ? '#6bb8e8' : assistantColor }"
                  >
                    {{ msg.role === "user" ? "U" : "AI" }}
                  </div>
                </div>
                <div class="flex" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
                  <div
                    class="p-3 text-[14px] rounded-2xl whitespace-pre-wrap leading-relaxed shadow-sm w-full"
                    :class="
                      msg.role === 'user'
                        ? 'bg-[#e8f4fd] text-gray-800 rounded-tr-none dark:bg-blue-900/40 dark:text-gray-200'
                        : 'bg-white text-gray-800 rounded-tl-none dark:bg-gray-800 dark:text-gray-200'
                    "
                    @mouseup="msg.role === 'assistant' ? handleSelectText() : undefined"
                  >
                    <!-- eslint-disable-next-line vue/no-v-html -->
                    <div class="msg-content" v-html="msg.renderedContent"></div>
                    <span
                      v-if="
                        isGenerating && index === messages.length - 1 && msg.content === ''
                      "
                      class="inline-block mt-1"
                    >
                      <NSpin size="small" />
                    </span>
                  </div>
                </div>
                <div
                  v-if="msg.content"
                  class="flex items-center gap-0.5 justify-end"
                >
                  <ButtonIcon
                    icon="mdi:content-copy"
                    class="!h-28px !w-28px text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
                    tooltip-content="复制"
                    @click.stop="copyToClipboard(msg.content)"
                  />
                  <ButtonIcon
                    v-if="enableVocabulary && msg.role === 'assistant'"
                    icon="mdi:star-outline"
                    class="!h-28px !w-28px text-gray-400 hover:text-amber-500 dark:text-gray-500 dark:hover:text-amber-400"
                    tooltip-content="添加到生词本"
                    @click.stop="openVocabModal()"
                  />
                  <ButtonIcon
                    v-if="msg.role === 'assistant'"
                    icon="mdi:notebook-edit-outline"
                    class="!h-28px !w-28px text-gray-400 hover:text-blue-500 dark:text-gray-500 dark:hover:text-blue-400"
                    tooltip-content="添加笔记"
                    @click.stop="openNoteModal(msg.content)"
                  />
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
            </template>
          </div>
        </div>
      </NScrollbar>

      <div
        class="border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-dark flex items-stretch"
        :class="appStore.isMobile ? 'p-2 gap-2' : 'p-4 gap-4'"
      >
        <NInput
          v-model:value="inputMessage"
          type="textarea"
          :autosize="appStore.isMobile ? { minRows: 1, maxRows: 4 } : { minRows: 2, maxRows: 6 }"
          :placeholder="inputPlaceholder"
          class="flex-1 shadow-sm"
          :size="appStore.isMobile ? 'medium' : 'large'"
          @keydown="handleEnter"
        />
        <div class="flex flex-col gap-2">
          <NButton
            type="primary"
            :size="appStore.isMobile ? 'medium' : 'large'"
            :loading="isGenerating"
            :disabled="!inputMessage.trim()"
            class="rounded-lg flex-1"
            :class="appStore.isMobile ? 'px-4' : 'px-8'"
            style="align-self: stretch; height: auto"
            @click="sendMessage"
          >
            发送
          </NButton>
        </div>
      </div>
    </NCard>

    <NModal
      v-if="enableVocabulary"
      v-model:show="showVocabModal"
      preset="card"
      title="添加到生词本"
      :style="{ width: appStore.isMobile ? '95vw' : '' }"
      class="max-w-md"
      :segmented="{ content: 'soft' }"
    >
      <NForm :model="vocabForm" label-placement="left" :label-width="appStore.isMobile ? '60' : '80'">
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
          <NButton type="primary" :loading="vocabLoading" @click="submitVocab">
            确认添加
          </NButton>
        </div>
      </template>
    </NModal>

    <NModal
      v-model:show="showNoteModal"
      preset="card"
      title="添加笔记"
      :style="{ width: appStore.isMobile ? '95vw' : '800px' }"
      :segmented="{ content: 'soft' }"
    >
      <NForm :model="noteForm" label-placement="left" :label-width="appStore.isMobile ? '60' : '80'">
        <div :class="appStore.isMobile ? 'flex flex-col gap-2' : 'flex gap-4'">
          <NFormItem label="标题" path="title" class="flex-1">
            <NInput v-model:value="noteForm.title" placeholder="输入笔记标题" />
          </NFormItem>
          <NFormItem label="分类" path="category" :style="appStore.isMobile ? {} : { width: '240px' }">
            <NInput v-model:value="noteForm.category" placeholder="输入笔记分类" />
          </NFormItem>
        </div>
        <NFormItem label="内容" path="content">
          <div :class="appStore.isMobile ? 'flex flex-col gap-4 w-full' : 'grid grid-cols-2 gap-4 w-full'">
            <NInput
              v-model:value="noteForm.content"
              type="textarea"
              :autosize="appStore.isMobile ? { minRows: 6, maxRows: 10 } : { minRows: 12, maxRows: 15 }"
              placeholder="输入笔记内容"
            />
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div
              class="prose dark:prose-invert max-w-none overflow-y-auto p-4 border border-gray-200 dark:border-gray-700 rounded-md bg-gray-50/50 dark:bg-dark-100 text-sm leading-relaxed"
              style="height: 100%; max-height: 350px"
              v-html="renderMarkdown(noteForm.content)"
            ></div>
          </div>
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showNoteModal = false">取消</NButton>
          <NButton type="primary" :loading="noteLoading" @click="submitNote">
            确认添加
          </NButton>
        </div>
      </template>
    </NModal>

    <NModal
      v-model:show="showTitleModal"
      preset="dialog"
      title="编辑标题"
      positive-text="保存"
      negative-text="取消"
      @positive-click="handleSaveTitle"
    >
      <NInput
        v-model:value="editTitle"
        placeholder="请输入标题"
        maxlength="50"
        show-count
      />
    </NModal>

    <NDrawer v-model:show="showPromptEditor" :width="appStore.isMobile ? '85vw' : 600" placement="right">
      <NDrawerContent
        :title="`设置 - ${routeTitleMap[route.name as string] || 'AI 助手'}`"
        closable
        body-content-style="padding: 0; display: flex; flex-direction: column; height: 100%;"
      >
        <PromptEditor
          :module-key="moduleKey"
          :module-name="routeTitleMap[route.name as string]"
          :default-prompt="props.systemPrompt"
          @updated="refreshPrompt"
        />
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
:deep(.msg-content p) {
  margin: 0;
}
:deep(.msg-content p + p) {
  margin-top: 0.5em;
}
</style>
