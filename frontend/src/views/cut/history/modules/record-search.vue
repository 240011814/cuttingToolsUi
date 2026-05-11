<script setup lang="ts">
import { computed, defineEmits, defineModel, defineOptions } from 'vue';
import { useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'RecordSearch'
});

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useNaiveForm();

// Ensure startTime and endTime are number | null | undefined for NDatePicker compatibility
const model = defineModel<Api.Cut.CutRecordSearchParams>('model', { required: true });

type RuleKey = Extract<keyof Api.Cut.CutRecordSearchParams, 'name' | 'startTime' | 'endTime' | 'type'>;

const rules = computed<Record<RuleKey, App.Global.FormRule>>(() => {
  return {
    name: { required: false, message: $t('page.cut.inputName'), trigger: 'blur' },
    startTime: { required: false, type: 'date', trigger: 'change' },
    endTime: { required: false, type: 'date', trigger: 'change' },
    type: { required: false, trigger: 'change' }
  };
});

async function reset() {
  await restoreValidation();
  emit('reset');
}

async function search() {
  await validate();
  emit('search');
}
</script>

<template>
  <NCard :bordered="false" size="small" class="card-wrapper">
    <NCollapse>
      <NCollapseItem :title="$t('common.search')" name="record-search">
        <NForm ref="formRef" :model="model" :rules="rules" label-placement="left" :label-width="80">
          <NGrid responsive="screen" item-responsive>
            <!-- 名称 -->
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.cut.name')" path="name" class="pr-24px">
              <NInput v-model:value="model.name" :placeholder="$t('page.cut.inputName')" />
            </NFormItemGi>

            <!-- 类型 -->
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.cut.type')" path="type" class="pr-24px">
              <NSelect
                v-model:value="model.type"
                :options="[
                  { label: '一维', value: '1' },
                  { label: '平面', value: '2' }
                ]"
                :placeholder="$t('page.cut.selectType')"
                clearable
              />
            </NFormItemGi>

            <!-- 开始时间 -->
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.cut.startTime')" path="startTime" class="pr-24px">
              <NDatePicker
                v-model:value="model.startTime"
                type="datetime"
                clearable
                :placeholder="$t('page.cut.inputStartTime')"
              />
            </NFormItemGi>

            <!-- 结束时间 -->
            <NFormItemGi span="24 s:12 m:6" :label="$t('page.cut.endTime')" path="endTime" class="pr-24px">
              <NDatePicker
                v-model:value="model.endTime"
                type="datetime"
                clearable
                :placeholder="$t('page.cut.inputEndTime')"
              />
            </NFormItemGi>

            <!-- 操作按钮 -->
            <NFormItemGi span="24 " class="pr-24px">
              <NSpace class="w-full" justify="end">
                <NButton @click="reset">
                  <template #icon>
                    <icon-ic-round-refresh class="text-icon" />
                  </template>
                  {{ $t('common.reset') }}
                </NButton>
                <NButton type="primary" ghost @click="search">
                  <template #icon>
                    <icon-ic-round-search class="text-icon" />
                  </template>
                  {{ $t('common.search') }}
                </NButton>
              </NSpace>
            </NFormItemGi>
          </NGrid>
        </NForm>
      </NCollapseItem>
    </NCollapse>
  </NCard>
</template>

<style scoped></style>
