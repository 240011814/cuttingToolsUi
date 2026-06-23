<script setup lang="ts">
import { computed } from 'vue';
import { NButton, NPopconfirm, NTag } from 'naive-ui';
import { useAuth } from '@/hooks/business/auth';
import type { Api } from '@/service/api';

interface Props {
  item: Api.ModelScenario.Item;
  expanded: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  toggle: [];
  edit: [item: Api.ModelScenario.Item];
  delete: [id: number];
}>();

const { hasAuth } = useAuth();

const isModel = computed(() => props.item.type === 'model');
const icon = computed(() => isModel.value ? 'mdi:brain' : 'mdi:movie-open-outline');
const detailLabel = computed(() => isModel.value ? '适用场景' : '应对策略');
const detailEmoji = computed(() => isModel.value ? '🎯' : '🛡️');
</script>

<template>
  <div
    class="group relative cursor-pointer rounded-xl border border-gray-200 bg-white p-5 shadow-sm transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
    :class="{ 'border-l-4 border-l-gray-800 dark:border-l-gray-400': !expanded, 'ring-2 ring-blue-200 dark:ring-blue-800': expanded }"
    @click="emit('toggle')"
  >
    <!-- Header: icon + name + category -->
    <div class="flex items-center gap-3">
      <div
        class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg"
        :class="isModel ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-purple-50 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400'"
      >
        <SvgIcon :icon="icon" class="text-xl" />
      </div>
      <div class="flex-1 min-w-0">
        <div class="text-[15px] font-bold truncate text-gray-800 dark:text-gray-200">
          {{ item.name }}
        </div>
      </div>
      <NTag v-if="item.category" size="small" :bordered="false" :type="isModel ? 'info' : 'warning'">
        {{ item.category }}
      </NTag>
    </div>

    <!-- Summary -->
    <div v-if="item.summary" class="mt-3 text-[13px] leading-relaxed text-gray-500 dark:text-gray-400 line-clamp-2">
      {{ item.summary }}
    </div>

    <!-- Expanded detail -->
    <Transition name="expand">
      <div v-if="expanded" class="mt-4 pt-4 border-t border-gray-100 dark:border-gray-700" @click.stop>
        <div v-if="item.description" class="mb-3">
          <div class="text-xs font-bold mb-1 text-gray-600 dark:text-gray-400">📝 详细描述</div>
          <div class="text-[13px] leading-relaxed whitespace-pre-wrap text-gray-700 dark:text-gray-300">
            {{ item.description }}
          </div>
        </div>
        <div v-if="item.detail" class="mb-3">
          <div class="text-xs font-bold mb-1 text-gray-600 dark:text-gray-400">{{ detailEmoji }} {{ detailLabel }}</div>
          <div class="text-[13px] leading-relaxed whitespace-pre-wrap text-gray-700 dark:text-gray-300">
            {{ item.detail }}
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <NButton v-if="hasAuth('model_scenario:update')" size="small" type="primary" @click.stop="emit('edit', item)">
            编辑
          </NButton>
          <NPopconfirm v-if="hasAuth('model_scenario:delete')" @positive-click="emit('delete', item.id)">
            <template #trigger>
              <NButton size="small" type="error" @click.stop>删除</NButton>
            </template>
            确认删除？
          </NPopconfirm>
        </div>
      </div>
    </Transition>

    <!-- Hover actions hint -->
    <div class="absolute top-3 right-3 flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100" @click.stop>
      <NButton v-if="hasAuth('model_scenario:update')" size="tiny" quaternary @click="emit('edit', item)">
        <template #icon><SvgIcon icon="mdi:pencil-outline" /></template>
      </NButton>
      <NPopconfirm v-if="hasAuth('model_scenario:delete')" @positive-click="emit('delete', item.id)">
        <template #trigger>
          <NButton size="tiny" quaternary type="error">
            <template #icon><SvgIcon icon="mdi:delete-outline" /></template>
          </NButton>
        </template>
        确认删除？
      </NPopconfirm>
    </div>
  </div>
</template>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  margin-top: 0;
  padding-top: 0;
}
.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 500px;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
