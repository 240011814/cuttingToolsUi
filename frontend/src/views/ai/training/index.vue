<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useAuth } from '@/hooks/business/auth';

const router = useRouter();
const { hasAuth } = useAuth();

interface TrainingModule {
  key: string;
  title: string;
  description: string;
  icon: string;
  color: string;
  route: string;
  permission: string;
}

const modules: TrainingModule[] = [
  {
    key: 'chat',
    title: '英语训练',
    description: '通过模拟真实生活场景练习地道英语口语表达',
    icon: 'mdi:translate-variant',
    color: '#2080f0',
    route: '/ai/chat',
    permission: 'ai:chat:view'
  },
  {
    key: 'decision',
    title: '决策训练',
    description: '学习60+决策模型，提升在工作生活中的决策能力',
    icon: 'mdi:scale-balance',
    color: '#8a6d3b',
    route: '/ai/decision',
    permission: 'ai:decision:view'
  },
  {
    key: 'social',
    title: '社交训练',
    description: '练习聊天破冰、安慰、拒绝等40+沟通场景',
    icon: 'mdi:account-group-outline',
    color: '#7c3aed',
    route: '/ai/social',
    permission: 'ai:social:view'
  },
  {
    key: 'emergency',
    title: '应急训练',
    description: '突发应变与反应力训练，掌握应变策略',
    icon: 'mdi:incognito',
    color: '#d9534f',
    route: '/ai/emergency',
    permission: 'ai:emergency:view'
  }
];

function goToModule(route: string) {
  router.push(route);
}
</script>

<template>
  <div class="h-full overflow-auto p-6">
    <div class="mx-auto max-w-4xl">
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-800 dark:text-gray-200">训练中心</h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">选择训练模块，开始你的提升之旅</p>
      </div>

      <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
        <template v-for="item in modules" :key="item.key">
          <div
            v-if="hasAuth(item.permission)"
            class="group cursor-pointer rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
            @click="goToModule(item.route)"
          >
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg"
                :style="{ backgroundColor: item.color + '15', color: item.color }"
              >
                <SvgIcon :icon="item.icon" class="text-2xl" />
              </div>
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">
                  {{ item.title }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ item.description }}
                </p>
              </div>
              <SvgIcon
                icon="mdi:chevron-right"
                class="text-xl text-gray-400 transition-transform group-hover:translate-x-1"
              />
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
