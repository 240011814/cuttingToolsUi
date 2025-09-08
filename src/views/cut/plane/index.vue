<script setup lang="ts">
import { h, ref } from 'vue';
import { NButton, useMessage } from 'naive-ui';
import { cutBin } from '@/service/api';

const message = useMessage();
// 数据模型

// 响应式数据
const label = ref('');
const group = ref(false);
const width = ref<number | null>(null);
const height = ref<number | null>(null);
const quantity = ref(1);
const newMaterialHeight = ref(200);
const newMaterialWidth = ref(200);
const materialName = ref('');
const materialWidth = ref<number | null>(null);
const materialHeight = ref<number | null>(null);
const materialCount = ref(1);
const saveData = ref<Api.Cut.RecordRequest | null>(null);
const items = ref<Api.Cut.Item[]>([]);
const materials = ref<Api.Cut.Item[]>([]);
const results = ref<Api.Cut.BinResult[]>([]);

// 用于保存 canvas 引用
const canvases = ref<(HTMLCanvasElement | null)[]>([]);

const loading = ref(false);

// item 表格
const itemColumns = [
  { title: '标签', key: 'label' },
  { title: '宽(cm)', key: 'width' },
  { title: '高(cm)', key: 'height' },
  { title: '数量', key: 'quantity' },
  {
    title: '操作',
    key: 'actions',
    render(_row: any, index: number) {
      return h(
        NButton,
        {
          size: 'small',
          type: 'error',
          onClick: () => removeItem(index)
        },
        { default: () => '删除' }
      );
    }
  }
];

// material 表格
const materialColumns = [
  { title: '标签', key: 'label' },
  { title: '宽(cm)', key: 'width' },
  { title: '高(cm)', key: 'height' },
  { title: '数量', key: 'quantity' },
  {
    title: '操作',
    key: 'actions',
    render(_row: any, index: number) {
      return h(
        NButton,
        {
          size: 'small',
          type: 'error',
          onClick: () => removeMaterial(index)
        },
        { default: () => '删除' }
      );
    }
  }
];

// 添加项目
function addItem() {
  if (
    !label.value ||
    width.value === null ||
    height.value === null ||
    quantity.value < 1 ||
    width.value <= 0 ||
    height.value <= 0
  ) {
    message.error('请输入有效的项目参数！');
    return;
  }

  const existingIndex = items.value.findIndex((item: Api.Cut.Item) => item.label === label.value);
  if (existingIndex !== -1) {
    items.value[existingIndex].quantity = quantity.value;
  } else {
    items.value.push({
      label: label.value,
      width: width.value,
      height: height.value,
      quantity: quantity.value
    });
  }

  clearItemInputs();
}

// 添加材料
function addMaterial() {
  if (
    !materialName.value ||
    materialWidth.value === null ||
    materialHeight.value === null ||
    materialCount.value < 1 ||
    materialWidth.value <= 0 ||
    materialHeight.value <= 0
  ) {
    message.error('请输入有效的材料参数！');
    return;
  }

  materials.value.push({
    label: materialName.value,
    width: materialWidth.value,
    height: materialHeight.value,
    quantity: materialCount.value
  });

  clearMaterialInputs();
}

// 删除项目
function removeItem(index: number) {
  items.value.splice(index, 1);
}

// 删除材料
function removeMaterial(index: number) {
  materials.value.splice(index, 1);
}

// 清空所有
function clearAll() {
  items.value = [];
  materials.value = [];
  results.value = [];
  canvases.value = [];
}

// 清空输入框
function clearItemInputs() {
  label.value = '';
  width.value = null;
  height.value = null;
  quantity.value = 1;
}

function clearMaterialInputs() {
  materialName.value = '';
  materialWidth.value = null;
  materialHeight.value = null;
  materialCount.value = 1;
}

// 优化主逻辑
async function runOptimization() {
  if (items.value.length === 0) {
    message.error('请先添加至少一个切割项目！');
    return;
  }

  const expandedItems: Api.Cut.Item[] = items.value.flatMap((item: Api.Cut.Item) => {
    const count = item.quantity ?? 0;
    if (count < 1) return [];
    return Array.from({ length: count }, (_, i) => ({
      label: `${item.label}_${i + 1}`,
      width: item.width,
      height: item.height
    }));
  });

  try {
    loading.value = true;
    const request = {
      items: expandedItems,
      materials: materials.value,
      width: newMaterialWidth.value,
      height: newMaterialHeight.value
    };
    const data = await cutBin(request);
    const { data: reslut } = data;
    if (!reslut || reslut.length === 0) {
      message.warning('无法使用现有材料完成所有切割项目，将使用新材料。');
    } else {
      saveData.value = {
        type: '2',
        request: JSON.stringify({ rowItems: items.value, ...request }),
        response: JSON.stringify(reslut),
        name: ``
      };
      results.value = reslut;
    }
  } catch {
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="p-4">
    <NCard title="材料裁剪可视化" size="large" class="mb-4">
      <!-- 添加切割项目 -->

      <h3 class="mb-3 text-lg font-semibold">裁剪尺寸</h3>
      <div class="mb-2 flex items-center gap-2">
        <NInput v-model:value="label" class="input-width" type="text" placeholder="标签" />
        <NInputNumber v-model:value="width" type="number" placeholder="宽(cm)" step="0.1" min="0.1" class="w-40" />
        <NInputNumber v-model:value="height" type="number" placeholder="高(cm)" step="0.1" min="0.1" class="w-40" />
        <NInputNumber v-model:value="quantity" type="number" placeholder="数量" class="w-40" min="1" />
        <NButton type="primary" @click="addItem">添加尺寸</NButton>
      </div>

      <!-- 切割项目列表 -->

      <h3 class="mb-2 text-lg font-semibold">切割项目</h3>
      <NDataTable :columns="itemColumns" :data="items" />

      <!-- 添加剩余材料 -->

      <h3 class="mb-3 text-lg font-semibold">库存材料</h3>
      <div class="mb-2 flex items-center gap-2">
        <NInput v-model:value="materialName" type="text" placeholder="材料名称" class="input-width" />
        <NInputNumber
          v-model:value="materialWidth"
          type="number"
          placeholder="宽(cm)"
          step="0.1"
          min="0.1"
          class="w-40"
        />
        <NInputNumber
          v-model:value="materialHeight"
          type="number"
          placeholder="高(cm)"
          step="0.1"
          min="0.1"
          class="w-40"
        />
        <NInputNumber v-model:value="materialCount" type="number" placeholder="数量" class="w-40" min="1" />
        <NButton type="primary" @click="addMaterial">添加材料</NButton>
      </div>

      <!-- 剩余材料列表 -->

      <h3 class="mb-2 text-lg font-semibold">库存材料</h3>
      <NDataTable :columns="materialColumns" :data="materials" />

      <h3 class="mt-6">参数配置</h3>
      <div class="mb-4 flex items-center gap-6">
        <div class="flex items-center gap-2">
          <span class="w-24">新材料长度</span>
          <NInputNumber v-model:value="newMaterialHeight" class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">新材料高度</span>
          <NInputNumber v-model:value="newMaterialWidth" class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">聚合显示</span>
          <NSwitch v-model:value="group" class="w-40" />
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="mt-4 flex gap-2">
        <NButton type="primary" @click="runOptimization">开始裁剪</NButton>
        <PlanePrinter :results="results" :materials="materials"></PlanePrinter>
        <SaveCutRecord :data="saveData" @saved="saveData = null"></SaveCutRecord>
        <NButton type="warning" @click="clearAll">清空所有</NButton>
      </div>
    </NCard>

    <PlaneStats :results="results"></PlaneStats>
    <PlaneCanvas :results="results" :group-data="group" :materials="materials"></PlaneCanvas>

    <NModal v-model:show="loading" preset="dialog" title="计算中...">
      <div class="flex flex-col items-center justify-center p-6">
        <NSpin size="large" />
        <div class="mt-3">{{ $t('common.loading') }}</div>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.input-width {
  width: 200px;
}
</style>
