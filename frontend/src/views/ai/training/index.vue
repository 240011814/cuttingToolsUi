<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import { fetchCustomTrainingList, fetchCustomTrainingDetail, fetchCreateCustomTraining, fetchUpdateCustomTraining, fetchDeleteCustomTraining } from '@/service/api';
import type { CustomTraining } from '@/service/api';

const router = useRouter();
const message = useMessage();
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

const customTrainings = ref<CustomTraining[]>([]);
const loading = ref(false);

const showModal = ref(false);
const isEdit = ref(false);
const editId = ref(0);
const submitting = ref(false);

const form = ref({
  title: '',
  description: '',
  system_prompt: '',
  icon: 'mdi:robot-outline',
  color: '#2080f0',
  initial_message: '',
  input_placeholder: '输入消息... (回车发送，Shift + 回车换行)',
  speech_lang: 'zh-CN',
  speech_rate: 0.95
});

const iconOptions = [
  { label: '机器人', value: 'mdi:robot-outline' },
  { label: '对话', value: 'mdi:chat-outline' },
  { label: '脑', value: 'mdi:brain' },
  { label: '灯泡', value: 'mdi:lightbulb-outline' },
  { label: '书', value: 'mdi:book-open-outline' },
  { label: '铅笔', value: 'mdi:pencil-outline' },
  { label: '星星', value: 'mdi:star-outline' },
  { label: '心', value: 'mdi:heart-outline' },
  { label: '火箭', value: 'mdi:rocket-launch-outline' },
  { label: '齿轮', value: 'mdi:cog-outline' }
];

const colorOptions = [
  { label: '蓝色', value: '#2080f0' },
  { label: '绿色', value: '#18a058' },
  { label: '红色', value: '#d9534f' },
  { label: '橙色', value: '#f0a020' },
  { label: '紫色', value: '#7c3aed' },
  { label: '棕色', value: '#8a6d3b' },
  { label: '灰色', value: '#666666' },
  { label: '青色', value: '#20c997' }
];

const resetForm = () => {
  form.value = {
    title: '',
    description: '',
    system_prompt: '',
    icon: 'mdi:robot-outline',
    color: '#2080f0',
    initial_message: '',
    input_placeholder: '输入消息... (回车发送，Shift + 回车换行)',
    speech_lang: 'zh-CN',
    speech_rate: 0.95
  };
};

const loadCustomTrainings = async () => {
  loading.value = true;
  try {
    const { data } = await fetchCustomTrainingList();
    if (data) {
      customTrainings.value = data;
    }
  } catch (err: any) {
    console.error('加载自定义训练失败:', err);
  } finally {
    loading.value = false;
  }
};

function goToModule(route: string) {
  router.push(route);
}

function goToCustomTraining(id: number) {
  router.push(`/ai/custom-training/${id}`);
}

function handleCreate() {
  isEdit.value = false;
  editId.value = 0;
  resetForm();
  showModal.value = true;
}

async function handleEdit(id: number) {
  isEdit.value = true;
  editId.value = id;
  try {
    const { data } = await fetchCustomTrainingDetail(id);
    if (data) {
      form.value = {
        title: data.title,
        description: data.description,
        system_prompt: data.system_prompt,
        icon: data.icon || 'mdi:robot-outline',
        color: data.color || '#2080f0',
        initial_message: data.initial_message,
        input_placeholder: data.input_placeholder || '输入消息... (回车发送，Shift + 回车换行)',
        speech_lang: data.speech_lang || 'zh-CN',
        speech_rate: data.speech_rate || 0.95
      };
    }
  } catch (err: any) {
    message.error(`加载失败: ${err?.message || '未知错误'}`);
    return;
  }
  showModal.value = true;
}

async function handleSubmit() {
  if (!form.value.title.trim()) {
    message.warning('请输入训练标题');
    return;
  }
  if (!form.value.system_prompt.trim()) {
    message.warning('请输入系统提示词');
    return;
  }

  submitting.value = true;
  try {
    if (isEdit.value) {
      await fetchUpdateCustomTraining(editId.value, form.value);
      message.success('更新成功');
    } else {
      await fetchCreateCustomTraining(form.value);
      message.success('创建成功');
    }
    showModal.value = false;
    loadCustomTrainings();
  } catch (err: any) {
    message.error(`操作失败: ${err?.message || '未知错误'}`);
  } finally {
    submitting.value = false;
  }
}

async function handleDeleteCustomTraining(id: number) {
  try {
    await fetchDeleteCustomTraining(id);
    message.success('删除成功');
    loadCustomTrainings();
  } catch (err: any) {
    message.error(`删除失败: ${err?.message || '未知错误'}`);
  }
}

onMounted(() => {
  loadCustomTrainings();
});
</script>

<template>
  <div class="h-full overflow-auto p-6">
    <div class="mx-auto max-w-4xl">
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-800 dark:text-gray-200">训练中心</h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">选择训练模块，开始你的提升之旅</p>
      </div>

      <!-- 内置训练模块 -->
      <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mb-8">
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

      <!-- 自定义训练 -->
      <div v-if="hasAuth('ai:custom-training:view')">
        <div class="mb-6 flex items-center justify-between">
          <h2 class="text-xl font-bold text-gray-800 dark:text-gray-200">自定义训练</h2>
          <NButton v-if="hasAuth('ai:custom-training:create')" type="primary" @click="handleCreate">
            <template #icon>
              <SvgIcon icon="mdi:plus" />
            </template>
            新增训练
          </NButton>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <NSpin size="large" />
        </div>

        <div v-else-if="customTrainings.length === 0" class="rounded-xl border border-dashed border-gray-300 p-8 text-center dark:border-gray-600">
          <SvgIcon icon="mdi:robot-outline" class="text-4xl text-gray-400 mb-2" />
          <p class="text-gray-500 dark:text-gray-400">还没有自定义训练</p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">点击上方按钮创建你的专属训练</p>
        </div>

        <div v-else class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div
            v-for="item in customTrainings"
            :key="item.id"
            class="group relative cursor-pointer rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-all hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
            @click="goToCustomTraining(item.id)"
          >
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg"
                :style="{ backgroundColor: (item.color || '#2080f0') + '15', color: item.color || '#2080f0' }"
              >
                <SvgIcon :icon="item.icon || 'mdi:robot-outline'" class="text-2xl" />
              </div>
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">
                  {{ item.title }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                  {{ item.description || '暂无描述' }}
                </p>
              </div>
              <SvgIcon
                icon="mdi:chevron-right"
                class="text-xl text-gray-400 transition-transform group-hover:translate-x-1"
              />
            </div>
            <div
              v-if="hasAuth('ai:custom-training:edit') || hasAuth('ai:custom-training:delete')"
              class="absolute bottom-3 right-3 flex gap-1 opacity-0 transition-opacity group-hover:opacity-100"
              @click.stop
            >
              <NButton v-if="hasAuth('ai:custom-training:edit')" size="small" quaternary @click="handleEdit(item.id)">
                <template #icon>
                  <SvgIcon icon="mdi:pencil-outline" />
                </template>
              </NButton>
              <NPopconfirm v-if="hasAuth('ai:custom-training:delete')" @positive-click="handleDeleteCustomTraining(item.id)">
                <template #trigger>
                  <NButton size="small" quaternary type="error">
                    <template #icon>
                      <SvgIcon icon="mdi:delete-outline" />
                    </template>
                  </NButton>
                </template>
                确定删除这个自定义训练吗？
              </NPopconfirm>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="isEdit ? '编辑训练' : '创建自定义训练'"
      style="width: 600px; max-height: 80vh"
      :segmented="{ content: 'soft', footer: 'soft' }"
    >
      <NForm :model="form" label-placement="left" label-width="80">
        <NFormItem label="训练标题" path="title">
          <NInput v-model:value="form.title" placeholder="例如：商务英语、面试模拟" maxlength="50" show-count />
        </NFormItem>
        <NFormItem label="描述" path="description">
          <NInput v-model:value="form.description" type="textarea" :autosize="{ minRows: 2, maxRows: 3 }" placeholder="简要描述训练内容" maxlength="200" show-count />
        </NFormItem>
        <NFormItem label="系统提示" path="system_prompt">
          <NInput v-model:value="form.system_prompt" type="textarea" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="定义AI的角色、行为和训练流程" />
        </NFormItem>
        <NFormItem label="欢迎语" path="initial_message">
          <NInput v-model:value="form.initial_message" type="textarea" :autosize="{ minRows: 2, maxRows: 3 }" placeholder="AI的开场白" />
        </NFormItem>
        <div class="grid grid-cols-2 gap-4">
          <NFormItem label="图标" path="icon">
            <NSelect v-model:value="form.icon" :options="iconOptions" />
          </NFormItem>
          <NFormItem label="颜色" path="color">
            <NSelect v-model:value="form.color" :options="colorOptions" />
          </NFormItem>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <NFormItem label="语言" path="speech_lang">
            <NSelect v-model:value="form.speech_lang" :options="[
              { label: '中文', value: 'zh-CN' },
              { label: '英文', value: 'en-US' },
              { label: '日文', value: 'ja-JP' }
            ]" />
          </NFormItem>
          <NFormItem label="语速" path="speech_rate">
            <NSlider v-model:value="form.speech_rate" :min="0.5" :max="1.5" :step="0.05" />
          </NFormItem>
        </div>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>
