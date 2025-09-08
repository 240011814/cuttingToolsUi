<script setup lang="ts">
import { computed, defineEmits, defineModel, defineOptions, defineProps, ref, watch } from 'vue';
import { enableStatusOptions } from '@/constants/business';
import { editUser } from '@/service/api';
import { useFormRules, useNaiveForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({
  name: 'UserOperateDrawer'
});

interface Props {
  /** the type of operation */
  operateType: NaiveUI.TableOperateType;
  /** the edit row data */
  rowData?: Api.SystemManage.User | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useNaiveForm();
const { defaultRequiredRule } = useFormRules();

const title = computed(() => {
  const titles: Record<NaiveUI.TableOperateType, string> = {
    add: $t('page.manage.user.addUser'),
    edit: $t('page.manage.user.editUser')
  };
  return titles[props.operateType];
});

type Model = Pick<Api.SystemManage.User, 'id' | 'userName' | 'nickName' | 'phone' | 'email' | 'userRoles' | 'status'>;

const model = ref(createDefaultModel());

function createDefaultModel(): Model {
  return {
    id: '',
    userName: '',
    nickName: '',
    phone: '',
    email: '',
    userRoles: [],
    status: '1'
  };
}

type RuleKey = Extract<keyof Model, 'userName'>;

const rules: Record<RuleKey, App.Global.FormRule> = {
  userName: defaultRequiredRule
};

/** the enabled role options */
const roleOptions = ref<CommonType.Option<string>[]>([]);

// async function getRoleOptions() {
//   const { error, data } = await fetchGetAllRoles();

//   if (!error) {
//     const options = data.map(item => ({
//       label: item.roleName,
//       value: item.roleCode
//     }));

//     // the mock data does not have the roleCode, so fill it
//     // if the real request, remove the following code
//     const userRoleOptions = model.value.userRoles.map(item => ({
//       label: item,
//       value: item
//     }));
//     // end

//     roleOptions.value = [...userRoleOptions, ...options];
//   }
// }

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model.value, props.rowData);
  }
}

function closeDrawer() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();
  await editUser({
    nickName: model.value.nickName,
    phone: model.value.phone,
    email: model.value.email,
    enabled: model.value.status === '1',
    id: model.value.id
  });
  window.$message?.success($t('common.updateSuccess'));
  closeDrawer();
  emit('submitted');
}

watch(visible, () => {
  if (visible.value) {
    handleInitModel();
    restoreValidation();
    // getRoleOptions();
  }
});
</script>

<template>
  <NDrawer v-model:show="visible" display-directive="show" :width="360">
    <NDrawerContent :title="title" :native-scrollbar="false" closable>
      <NForm ref="formRef" :model="model" :rules="rules">
        <NFormItem :label="$t('page.manage.user.userName')" path="userName">
          <NInput v-model:value="model.userName" :disabled="true" :placeholder="$t('page.manage.user.form.userName')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.nickName')" path="nickName">
          <NInput v-model:value="model.nickName" :placeholder="$t('page.manage.user.form.nickName')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.userPhone')" path="userPhone">
          <NInput v-model:value="model.phone" :placeholder="$t('page.manage.user.form.userPhone')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.userEmail')" path="email">
          <NInput v-model:value="model.email" :placeholder="$t('page.manage.user.form.userEmail')" />
        </NFormItem>
        <NFormItem :label="$t('page.manage.user.userStatus')" path="status">
          <NRadioGroup v-model:value="model.status">
            <NRadio v-for="item in enableStatusOptions" :key="item.value" :value="item.value" :label="$t(item.label)" />
          </NRadioGroup>
        </NFormItem>
        <NFormItem v-if="false" :label="$t('page.manage.user.userRole')" path="roles">
          <NSelect
            v-model:value="model.userRoles"
            multiple
            :options="roleOptions"
            :placeholder="$t('page.manage.user.form.userRole')"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace :size="16">
          <NButton @click="closeDrawer">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped></style>
