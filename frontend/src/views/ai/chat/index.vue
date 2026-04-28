<script setup lang="ts">
import { ref, nextTick } from 'vue';
import { useMessage } from 'naive-ui';
import { getServiceBaseURL } from '@/utils/service';
import { fetchAddVocabulary } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';
import { getAuthorization } from '@/service/request/shared';

interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
  suggestions?: any[];
}

const systemMessage: Message = {
  role: 'system',
  content: `You are a professional AI English Teacher specializing in scenario-based simulation training.
Your goal is to help users practice authentic spoken English through daily life scenarios.

Training Workflow:
1. **Scene Setup**: Start or continue a daily scenario (e.g., ordering food, business meeting, traveling).
2. **Translation Task**: Provide a specific sentence in Chinese and ask the user to translate it into English.
3. **Evaluation & Feedback**: After the user responds, evaluate their translation. Compare it with authentic native expressions, explain grammar/vocabulary points, and provide "Natural Expression" tips.
4. **Progressive Learning**: Move the story forward and provide the next Chinese sentence for the user to translate.

Response Structure:
- Use "地道表达" (Authentic Expression) for corrections.
- Use "💡 重点纠错与地道笔记" for detailed learning points.
- Always include a section "📊 模拟训练进度" to show the current scenario step.
- ALWAYS append identified vocabulary at the end in this format:
<vocabs>[{"word": "word", "phonetic": "...", "definition": "Chinese meaning", "example": "...", "confusingWords": "..."}]</vocabs>

Rules:
- Focus on oral, daily-use English.
- Be encouraging but precise with corrections.
- Do not mention the <vocabs> tag in your natural speech.
- **CRITICAL**: Every time you correct the user or introduce new words (like Sugar, Milk in your notes), you MUST extract them into the JSON format below and append it to the VERY END of your response.
Format Example:
<vocabs>[{"word": "Sugar", "phonetic": "/ˈʃʊɡ.ər/", "definition": "糖", "example": "Do you take sugar? (你要加糖吗？)", "confusingWords": "Shook (摇动), Shocker (令人震惊的事)"}]</vocabs>
If no new words, you can omit it, but if you taught anything, it MUST be there.`
};

const messages = ref<Message[]>([
  { role: 'assistant', content: 'Hello! 我是你的 AI 英语口语老师。我们可以通过模拟真实生活场景来练习地道表达。你想从哪个场景开始？比如：“咖啡店点单”、“酒店入住”或者“入职第一天”。' }
]);
const inputMessage = ref('');
const isGenerating = ref(false);
const scrollbarRef = ref<any>(null);
const message = useMessage();


// 生词本弹窗相关
const showVocabModal = ref(false);
const vocabLoading = ref(false);
const vocabForm = ref({
  word: '',
  phonetic: '',
  definition: '',
  example: '',
  confusingWords: ''
});

const openVocabModal = () => {
  vocabForm.value = {
    word: '',
    phonetic: '',
    definition: '',
    example: '',
    confusingWords: ''
  };
  showVocabModal.value = true;
};

// 如果用户选中了文本，尝试自动填充单词
const handleSelectText = () => {
  const selection = window.getSelection()?.toString().trim();
  if (selection && selection.length > 0 && selection.length < 50) {
    vocabForm.value.word = selection;
  }
};

const submitVocab = async () => {
  if (!vocabForm.value.word.trim()) {
    message.warning('请输入单词');
    return;
  }
  vocabLoading.value = true;
  try {
    await fetchAddVocabulary(vocabForm.value);
    message.success('已添加到生词本');
    showVocabModal.value = false;
  } catch (err: any) {
    message.error(`添加失败: ${err?.message || '未知错误'}`);
  } finally {
    vocabLoading.value = false;
  }
};


const scrollToBottom = async () => {
  await nextTick();
  if (scrollbarRef.value) {
    scrollbarRef.value.scrollTo({ position: 'bottom', behavior: 'smooth' });
  }
};

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isGenerating.value) return;

  const userText = inputMessage.value;
  inputMessage.value = '';

  messages.value.push({ role: 'user', content: userText });
  messages.value.push({ role: 'assistant', content: '' });

  scrollToBottom();
  isGenerating.value = true;
  const _authStore = useAuthStore();

  const isHttpProxy = import.meta.env.DEV && import.meta.env.VITE_HTTP_PROXY === 'Y';
  const { baseURL } = getServiceBaseURL(import.meta.env, isHttpProxy);
  try {
    const history = messages.value.slice(0, -1).filter(m => m.content && m.content.trim() !== '');
    const response = await fetch(`${baseURL}/api/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthorization() || ''
      },
      body: JSON.stringify({
        messages: [systemMessage, ...history]
      })
    });

    if (!response.ok) {
      let errMsg = `请求失败 (HTTP ${response.status})`;
      try {
        const errData = await response.json();
        if (errData.error) errMsg = errData.error;
      } catch {}
      throw new Error(errMsg);
    }

    const reader = response.body?.getReader();
    const decoder = new TextDecoder('utf-8');
    if (!reader) throw new Error('无法获取响应流');

    let buffer = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split(/\r?\n/); // 兼容 \r\n 和 \n
      buffer = lines.pop() || '';

      for (let line of lines) {
        line = line.trim();
        if (!line || !line.startsWith('data:')) continue;

        const dataStr = line.replace(/^data:\s*/, '').trim();
        if (dataStr === '[DONE]' || dataStr === '"[DONE]"') break;

        try {
          const dataObj = JSON.parse(dataStr);
          if (dataObj.error) {
            const lastIdx = messages.value.length - 1;
            messages.value[lastIdx] = { ...messages.value[lastIdx], content: `AI 服务错误: ${dataObj.error}` };
            break;
          }
          if (dataObj.content) {
            const lastIdx = messages.value.length - 1;
            // 核心修复：使用对象展开运算符触发 Vue 的响应式更新
            messages.value[lastIdx] = {
              ...messages.value[lastIdx],
              content: messages.value[lastIdx].content + dataObj.content
            };
            scrollToBottom();
          }
        } catch (e) {
          // 忽略不完整的 JSON 块
        }
      }
    }
  } catch (err: any) {
    const lastIdx = messages.value.length - 1;
    messages.value[lastIdx] = {
      ...messages.value[lastIdx],
      content: `连接 AI 服务失败: ${err?.message || '未知错误'}。\n请确认后端服务正常运行且已正确配置 DEEPSEEK_API_KEY。`
    };
  } finally {
    isGenerating.value = false;
    // 自动滚动到底部
    await nextTick();
    scrollToBottom();
    // 解析并自动保存生词
    autoSaveVocabs();
  }
};

const handleEnter = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
};

const autoSaveVocabs = async () => {
  const lastIdx = messages.value.length - 1;
  const lastMsg = messages.value[lastIdx];
  if (lastMsg.role !== 'assistant') return;

  const match = lastMsg.content.match(/<vocabs>([\s\S]*?)<\/vocabs>/);
  if (match && match[1]) {
    try {
      const vocabs = JSON.parse(match[1]);
      if (Array.isArray(vocabs) && vocabs.length > 0) {
        // 不再自动保存，而是存入建议列表
        messages.value[lastIdx].suggestions = vocabs;
      }
    } catch (e) {
      console.error('Failed to parse suggested vocabs:', e);
    }
  }
};

const handleApplySuggestion = (vocab: any) => {
  vocabForm.value = {
    word: vocab.word || '',
    phonetic: vocab.phonetic || '',
    definition: vocab.definition || '',
    example: vocab.example || '',
    confusingWords: vocab.confusingWords || ''
  };
  showVocabModal.value = true;
};

// 用于显示的计算属性，隐藏 <vocabs> 标签
const formatDisplayContent = (content: string) => {
  return content.replace(/<vocabs>[\s\S]*?<\/vocabs>/g, '').trim();
};

const handlePlay = (text: string) => {
  if (!window.speechSynthesis) {
    message.error('您的浏览器不支持语音播放');
    return;
  }
  window.speechSynthesis.cancel();
  const utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = 'en-US';
  utterance.rate = 0.9;
  window.speechSynthesis.speak(utterance);
};


</script>

<template>
  <div class="h-full flex-col flex gap-4 p-4 overflow-hidden">
    <NCard class="flex-1 overflow-hidden" content-class="p-0 flex flex-col h-full" :bordered="false" shadow="sm">
      <NScrollbar ref="scrollbarRef" class="flex-1 p-4 bg-gray-50/50 dark:bg-dark">
        <div class="flex flex-col gap-6 pb-4">
          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="flex items-start gap-3"
            :class="msg.role === 'user' ? 'flex-row-reverse' : 'flex-row'"
          >
            <NAvatar :color="msg.role === 'user' ? '#18a058' : '#2080f0'" round size="large">
              {{ msg.role === 'user' ? 'U' : 'AI' }}
            </NAvatar>
            <div class="flex flex-col gap-1 max-w-[80%]">
              <div
                class="p-4 rounded-2xl whitespace-pre-wrap leading-relaxed shadow-sm text-[15px] relative group"
                :class="msg.role === 'user' ? 'bg-[#18a058] text-white rounded-tr-none' : 'bg-white text-gray-800 rounded-tl-none dark:bg-gray-800 dark:text-gray-200'"
                @mouseup="msg.role === 'assistant' ? handleSelectText() : null"
              >
                {{ formatDisplayContent(msg.content) }}
                <span v-if="isGenerating && index === messages.length - 1 && msg.content === ''" class="inline-block mt-1">
                  <NSpin size="small" />
                </span>

                <!-- 收藏按钮栏 -->
                <div
                  v-if="msg.role === 'assistant' && msg.content"
                  class="absolute -right-2 -top-4 opacity-0 group-hover:opacity-100 transition-all duration-300 flex gap-1 p-1 bg-white dark:bg-gray-700 rounded-full shadow-md border border-gray-100 dark:border-gray-600 z-10"
                >
                  <ButtonIcon
                    icon="mdi:volume-high"
                    class="text-18px text-primary"
                    tooltip-content="朗读全文"
                    @click.stop="handlePlay(formatDisplayContent(msg.content))"
                  />
                  <div class="w-[1px] h-4 bg-gray-200 dark:bg-gray-600 self-center" />
                  <ButtonIcon
                    icon="mdi:star"
                    class="text-18px text-warning"
                    tooltip-content="添加到生词本"
                    @click.stop="openVocabModal()"
                  />
                </div>
              </div>

              <!-- 智能建议生词区域 -->
              <div v-if="msg.suggestions && msg.suggestions.length > 0" class="flex flex-wrap gap-2 mt-2">
                <span class="text-xs text-gray-400 self-center">智能建议:</span>
                <NTag
                  v-for="(vocab, vIdx) in msg.suggestions"
                  :key="vIdx"
                  size="small"
                  round
                  type="info"
                  check-strategy="child"
                  class="cursor-pointer hover:shadow-sm transition-shadow"
                  @click="handleApplySuggestion(vocab, msg.content)"
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

      <div class="p-4 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-dark flex items-stretch gap-4">
        <NInput
          v-model:value="inputMessage"
          type="textarea"
          :autosize="{ minRows: 2, maxRows: 6 }"
          placeholder="输入消息... (回车发送，Shift + 回车换行)"
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
          style="align-self: stretch; height: auto;"
          @click="sendMessage"
        >
          发送
        </NButton>
      </div>
    </NCard>

    <!-- 添加到生词本弹窗 -->
    <NModal
      v-model:show="showVocabModal"
      preset="card"
      title="添加到生词本"
      class="max-w-md"
      :segmented="{ content: 'soft' }"
    >
      <NForm :model="vocabForm" label-placement="left" label-width="80">
        <NFormItem label="单词" path="word">
          <div class="flex gap-2 w-full">
            <NInput v-model:value="vocabForm.word" placeholder="输入单词" class="flex-1" />
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
          <NButton type="primary" :loading="vocabLoading" @click="submitVocab">确认添加</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
/* 添加一些过渡动画让体验更丝滑 */
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
