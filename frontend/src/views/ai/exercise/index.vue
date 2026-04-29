<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMessage } from 'naive-ui';
import { fetchGetVocabularyList } from '@/service/api';

import typingSound from '@/assets/sound/typing.mp3';
import errorSound from '@/assets/sound/error.mp3';

const route = useRoute();
const router = useRouter();
const message = useMessage();

// --- Sound Engine ---
const playClick = () => {
  try {
    new Audio(typingSound).play().catch(() => {});
  } catch (e) {}
};

const playError = () => {
  try {
    new Audio(errorSound).play().catch(() => {});
  } catch (e) {}
};

// --- State ---
const isStarted = ref(false);
const loading = ref(false);
const rawWords = ref<any[]>([]);
const currentSentenceIndex = ref(0);
const isFinished = ref(false);
const isPlaying = ref(false);

const activeWordIndex = ref(0);
const currentInput = ref('');
const wordResults = ref<{ typed: string; status: 'pending' | 'correct' | 'error' }[]>([]);
const errorCounts = ref<number[]>([]);

// --- Computed ---
const currentItem = computed(() => rawWords.value[currentSentenceIndex.value] || null);
const targetSentence = computed(() => currentItem.value?.example || '');
const targetWords = computed(() => {
  if (!targetSentence.value) return [];
  return targetSentence.value.trim().split(/\s+/);
});

const progress = computed(() => {
  if (rawWords.value.length === 0) return 0;
  return Math.round((currentSentenceIndex.value / rawWords.value.length) * 100);
});

// --- Methods ---
const loadData = async () => {
  loading.value = true;
  try {
    const { data: res } = await fetchGetVocabularyList({ isMastered: false });
    if (res) {
      const { ids, mode } = route.query;
      if (mode === 'all') {
        rawWords.value = res.filter(item => item.example);
      } else if (ids) {
        // 如果指定了 ID，则需要重新获取完整列表（或者单独获取这些 ID），因为上面的请求只拿了未掌握的
        // 为了简单起见，如果传了 ids，我们重新请求一次不带过滤的列表
        const { data: allRes } = await fetchGetVocabularyList();
        if (allRes) {
          const idList = (ids as string).split(',').map(Number);
          rawWords.value = allRes.filter(item => idList.includes(item.id) && item.example);
        }
      } else {
        rawWords.value = res.filter(item => item.example);
      }

      if (rawWords.value.length === 0) {
        message.warning('没找到含有例句的可练习单词');
        router.push({ name: 'ai_vocabulary' });
      }else{
        rawWords.value.forEach(item => {
          item.example = item.example.replace(/\s*\([^)]*[\u4e00-\u9fa5][^)]*\)\s*$/, '')
        });
      }
    }
  } catch (err: any) {
    console.log(err);
    message.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

const initSentence = () => {
  activeWordIndex.value = 0;
  currentInput.value = '';
  wordResults.value = targetWords.value.map(() => ({ typed: '', status: 'pending' }));
  errorCounts.value = targetWords.value.map(() => 0);
  playCurrent(3);
};

const playCurrent = (times = 3) => {
  if (!targetSentence.value || isPlaying.value) return;
  window.speechSynthesis.cancel();

  const voices = window.speechSynthesis.getVoices();
  const englishVoice = voices.find(v => v.lang.startsWith('en')) || null;

  isPlaying.value = true;
  let count = 0;
  const speak = () => {
    if (count < times && !isFinished.value) {
      const utterance = new SpeechSynthesisUtterance(targetSentence.value);
      utterance.voice = englishVoice;
      utterance.lang = 'en-US';
      utterance.rate = 0.85;
      utterance.onend = () => {
        count++;
        if (count < times) {
          setTimeout(speak, 600);
        } else {
          isPlaying.value = false;
        }
      };
      window.speechSynthesis.speak(utterance);
    } else {
      isPlaying.value = false;
    }
  };
  speak();
};

const selectWord = (index: number) => {
  if (activeWordIndex.value === index) return;
  playClick();

  // 离开当前单词前进行校验
  validateWord(activeWordIndex.value, currentInput.value);

  // 切换到目标单词
  activeWordIndex.value = index;
  currentInput.value = wordResults.value[index].typed;
  // 重新进入编辑模式
  wordResults.value[index].status = 'pending';
};

const handleGlobalKeydown = (e: KeyboardEvent) => {
  if (!isStarted.value || isFinished.value) return;

  // 几乎所有按键都触发点击声
  if (e.key.length === 1 || ['Backspace', 'Enter', 'ArrowLeft', 'ArrowRight', ' '].includes(e.key)) {
    playClick();
  }

  // 方向键导航：切换前校验
  if (e.key === 'ArrowLeft') {
    if (activeWordIndex.value > 0) {
      selectWord(activeWordIndex.value - 1);
    }
    return;
  }
  if (e.key === 'ArrowRight') {
    if (activeWordIndex.value < targetWords.value.length - 1) {
      selectWord(activeWordIndex.value + 1);
    }
    return;
  }

  // 重播/提交
  if (e.key === 'Enter') {
    validateWord(activeWordIndex.value, currentInput.value);
    // 如果没全对，可以触发重播
    if (!wordResults.value.every(r => r.status === 'correct')) {
      playCurrent(3);
    }
    return;
  }

  // 空格：校验并尝试跳转到下一个待输入
  if (e.key === ' ') {
    e.preventDefault();
    const oldIndex = activeWordIndex.value;
    validateWord(oldIndex, currentInput.value);

    // 如果还没全对，且刚才校验的是当前焦点词，则移动到下一个
    if (!wordResults.value.every(r => r.status === 'correct') && oldIndex < targetWords.value.length - 1) {
      activeWordIndex.value = oldIndex + 1;
      currentInput.value = wordResults.value[activeWordIndex.value].typed;
    }
    return;
  }

  // 退格
  if (e.key === 'Backspace') {
    if (currentInput.value === '' && activeWordIndex.value > 0) {
      // 回退时不强制校验当前（因为是空的），但回退后的词进入pending
      activeWordIndex.value--;
      currentInput.value = wordResults.value[activeWordIndex.value].typed;
      wordResults.value[activeWordIndex.value].status = 'pending';
    } else {
      currentInput.value = currentInput.value.slice(0, -1);
    }
    return;
  }

  // 字母与符号
  if (e.key.length === 1) {
    currentInput.value += e.key;
  }
};

const validateWord = (index: number, typedValue: string) => {
  if (index < 0 || index >= targetWords.value.length) return;

  const target = targetWords.value[index];
  const typed = typedValue.trim();

  const normalize = (s: string) => s.toLowerCase().replace(/[.,!?;:]/g, '');
  const isCorrect = normalize(target) === normalize(typed);

  if (!isCorrect) {
    // 只有在非空或者用户确实离开时才判定错误
    playError();
    errorCounts.value[index]++;
    if (errorCounts.value[index] >= 6) {
      message.info(`提示：${target}`, { duration: 5000 });
    }
  } else {
    errorCounts.value[index] = 0;
  }

  wordResults.value[index] = {
    typed: typed,
    status: isCorrect ? 'correct' : 'error'
  };

  // 检查整句是否全部正确
  const allCorrect = wordResults.value.every(r => r.status === 'correct');
  if (allCorrect) {
    setTimeout(() => {
      if (wordResults.value.every(r => r.status === 'correct')) {
        if (currentSentenceIndex.value < rawWords.value.length - 1) {
          currentSentenceIndex.value++;
          initSentence();
        } else {
          isFinished.value = true;
        }
      }
    }, 800);
  }
};

const startPractice = () => {
  isStarted.value = true;
  initSentence();
};

const goBack = () => {
  router.push({ name: 'ai_vocabulary' });
};

onMounted(() => {
  loadData();
  window.addEventListener('keydown', handleGlobalKeydown);
  window.speechSynthesis.getVoices();
});

onUnmounted(() => {
  window.speechSynthesis.cancel();
  window.removeEventListener('keydown', handleGlobalKeydown);
});

watch(isFinished, (val) => {
  if (val) {
    window.removeEventListener('keydown', handleGlobalKeydown);
  }
});
</script>

<template>
  <div class="h-full flex flex-col p-4 gap-4 bg-white dark:bg-dark-bg transition-colors duration-300">
    <!-- Start Screen -->
    <div v-if="!isStarted && !loading" class="flex-1 flex-center">
      <div class="max-w-md w-full text-center space-y-10 animate-in fade-in duration-700">
        <div class="space-y-4">
          <h2 class="text-4xl font-bold text-gray-800 dark:text-gray-100 tracking-tight">沉浸式听写训练</h2>
          <p class="text-lg text-gray-500">听发音，还原地道例句，开启听觉与肌肉记忆之旅</p>
        </div>
        <NButton type="primary" size="large" class="w-full h-16 text-xl rd-2xl shadow-xl hover:scale-105 transition-transform" @click="startPractice">
          开始练习
        </NButton>
        <div class="text-xs text-gray-400 flex justify-center gap-6 uppercase tracking-widest">
          <span>Backspace 回退</span>
          <span>Space 跳转</span>
          <span>Enter 重播</span>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <NCard v-else-if="!isFinished" class="flex-1 flex flex-col items-center justify-center relative border-none shadow-none bg-transparent">
      <NProgress
        type="line"
        :percentage="progress"
        :show-indicator="false"
        class="absolute top-0 left-0 w-full"
        processing
        :height="3"
        color="#2080f0"
      />

      <div v-if="loading" class="flex-center h-full">
        <NSpin size="large" />
      </div>

      <div v-else-if="currentItem" class="max-w-5xl w-full flex flex-col items-center gap-24 py-12">
        <div class="text-center">
          <div class="text-2xl text-gray-400 dark:text-gray-500 font-medium tracking-wide">
            {{ currentItem.definition }}
          </div>
        </div>

        <!-- Word Blocks -->
        <div class="w-full px-8">
          <div class="flex flex-wrap justify-center gap-x-8 gap-y-16 font-mono text-4xl leading-none text-center min-h-[240px]">
            <div
              v-for="(word, idx) in targetWords"
              :key="idx"
              class="relative border-b-4 pb-3 transition-all duration-300 flex justify-center items-end cursor-pointer group whitespace-nowrap"
              :style="{ minWidth: (word.length || 1) + 'ch' }"
              :class="[
                activeWordIndex === idx ? 'border-primary scale-105' :
                wordResults[idx].status === 'error' ? 'border-red-500' :
                wordResults[idx].status === 'correct' ? 'border-transparent' : 'border-gray-100 dark:border-gray-800'
              ]"
              @click="selectWord(idx)"
            >
              <span v-if="activeWordIndex === idx" class="text-primary">
                {{ currentInput }}
                <span class="animate-pulse border-r-3 border-primary ml-1 h-10"></span>
              </span>
              <span
                v-else
                class="transition-all duration-300"
                :class="wordResults[idx].status === 'error' ? 'text-red-500' : 'text-gray-800 dark:text-gray-100'"
              >
                {{ wordResults[idx].typed || '&nbsp;' }}
              </span>
            </div>
          </div>
        </div>

        <!-- Status & Help -->
        <div class="mt-12 flex flex-col items-center gap-6 text-gray-300 dark:text-gray-600 select-none">
          <div v-if="isPlaying" class="flex items-center gap-2 animate-bounce text-primary">
            <icon-mdi-volume-high class="text-24px" />
            <span class="text-sm font-bold tracking-tighter">Listening...</span>
          </div>
          <div v-else-if="wordResults.some(r => r.status === 'error')" class="flex items-center gap-2 text-red-400 font-medium">
            <icon-mdi-alert-circle class="text-20px" />
            <span>请修正标红的拼写错误</span>
          </div>
          <div class="flex gap-10 text-xs uppercase tracking-widest font-bold">
            <span class="flex items-center gap-2"><NTag size="small" :bordered="false" round>← →</NTag> 切换单词</span>
            <span class="flex items-center gap-2"><NTag size="small" :bordered="false" round>Space</NTag> 下一个</span>
            <span class="flex items-center gap-2"><NTag size="small" :bordered="false" round>Enter</NTag> 重播</span>
          </div>
        </div>
      </div>
    </NCard>

    <!-- Result -->
    <NCard v-else class="flex-1 flex-center border-none shadow-none bg-transparent">
      <NResult
        status="success"
        title="训练达成！"
        description="例句听写训练已全部完成，您的语感得到了显著提升。"
        class="py-12"
      >
        <template #footer>
          <NSpace justify="center" :size="32">
            <NButton type="primary" size="large" secondary class="px-12 rd-lg" @click="goBack">
              返回生词本
            </NButton>
            <NButton size="large" quaternary class="px-12 rd-lg" @click="() => { currentSentenceIndex = 0; isFinished = false; isStarted = false; loadData(); }">
              再练一遍
            </NButton>
          </NSpace>
        </template>
      </NResult>
    </NCard>
  </div>
</template>

<style scoped>
.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>

<style scoped></style>
