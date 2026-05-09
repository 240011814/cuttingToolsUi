<script setup lang="ts">
import { computed } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { useAuthStore } from '@/store/modules/auth';
import { $t } from '@/locales';

defineOptions({
  name: 'HeaderBanner'
});

const appStore = useAppStore();
const authStore = useAuthStore();

const gap = computed(() => (appStore.isMobile ? 0 : 16));

const greetings = [
  '每一次练习，都是向更好的自己迈进',
  '坚持就是胜利，今天也要加油',
  '学习是最好的投资，继续努力',
  '每天进步一点点，积累就是力量',
  '知识改变命运，行动成就未来',
  '保持好奇心，探索无限可能',
  '今天的努力，是明天的收获',
  '相信自己，你比想象中更优秀',
  '学无止境，让我们一起成长',
  '每一次挑战，都是成长的机会',
  '成功源于坚持，继续前行',
  '用心学习，用爱生活',
  '让学习成为一种习惯',
  '今天的付出，明天的回报',
  '保持热情，追逐梦想'
];

const dailyGreeting = computed(() => {
  const today = new Date();
  const seed = today.getFullYear() * 10000 + (today.getMonth() + 1) * 100 + today.getDate();
  const index = seed % greetings.length;
  return greetings[index];
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper">
    <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
      <NGi span="24">
        <div class="flex-y-center">
          <div class="size-72px shrink-0 overflow-hidden rd-1/2">
            <img src="@/assets/imgs/soybean.jpg" class="size-full" />
          </div>
          <div class="pl-12px">
            <h3 class="text-18px font-semibold">
              {{ $t('page.home.greeting', { userName: authStore.userInfo.userName }) }}{{ dailyGreeting }}
            </h3>
          </div>
        </div>
      </NGi>
    </NGrid>
  </NCard>
</template>

<style scoped></style>
