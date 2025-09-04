<script setup lang="ts">
import { ref } from 'vue';
import { useMessage } from 'naive-ui';
import { cutBin } from '@/service/api';

const message = useMessage();
// 数据模型
interface Item {
  label: string;
  width: number;
  height: number;
  quantity: number;
}

interface Material {
  name: string;
  width: number;
  height: number;
  count: number;
}

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

const items = ref<Item[]>([]);
const materials = ref<Material[]>([]);
const results = ref<Api.Cut.BinResult[]>([]);

// 用于保存 canvas 引用
const canvases = ref<(HTMLCanvasElement | null)[]>([]);

const loading = ref(false);

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

  const existingIndex = items.value.findIndex((item: Item) => item.label === label.value);
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
    name: materialName.value,
    width: materialWidth.value,
    height: materialHeight.value,
    count: materialCount.value
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

  const expandedItems = items.value.flatMap((item: Item) => {
    return Array.from({ length: item.quantity }, (_, i) => ({
      label: `${item.label}_${i + 1}`,
      width: item.width,
      height: item.height
    }));
  });

  const materialData = materials.value.map((m: Material) => ({
    name: m.name,
    width: m.width,
    height: m.height,
    availableCount: m.count
  }));

  try {
    loading.value = true;
    const data = await cutBin({
      items: expandedItems,
      materials: materialData,
      width: newMaterialWidth.value,
      height: newMaterialHeight.value
    });
    const { data: reslut } = data;
    if (!reslut || reslut.length === 0) {
      message.warning('无法使用现有材料完成所有切割项目，将使用新材料。');
    } else {
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
      <section class="mb-6 border rounded-lg bg-gray-50 p-4">
        <h3 class="mb-3 text-lg font-semibold">裁剪尺寸</h3>
        <div class="flex flex-wrap items-center gap-3">
          <input v-model="label" type="text" placeholder="标签" class="border rounded px-3 py-2" />
          <input
            v-model.number="width"
            type="number"
            placeholder="宽(cm)"
            step="0.1"
            min="0.1"
            class="w-24 border rounded px-3 py-2"
          />
          <input
            v-model.number="height"
            type="number"
            placeholder="高(cm)"
            step="0.1"
            min="0.1"
            class="w-24 border rounded px-3 py-2"
          />
          <input
            v-model.number="quantity"
            type="number"
            placeholder="数量"
            class="w-20 border rounded px-3 py-2"
            min="1"
          />
          <NButton type="primary" @click="addItem">添加尺寸</NButton>
        </div>
      </section>

      <!-- 切割项目列表 -->
      <section class="mb-6">
        <h3 class="mb-2 text-lg font-semibold">切割项目</h3>
        <table class="w-full border-collapse text-sm">
          <thead>
            <tr class="bg-gray-100">
              <th class="border px-3 py-2">标签</th>
              <th class="border px-3 py-2">宽(cm)</th>
              <th class="border px-3 py-2">高(cm)</th>
              <th class="border px-3 py-2">数量</th>
              <th class="border px-3 py-2">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in items" :key="index" class="hover:bg-gray-50">
              <td class="border px-3 py-2">{{ item.label }}</td>
              <td class="border px-3 py-2">{{ item.width.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ item.height.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ item.quantity }}</td>
              <td class="border px-3 py-2">
                <button class="text-sm text-red-600" @click="removeItem(index)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <!-- 添加剩余材料 -->
      <section class="mb-6 border rounded-lg bg-gray-50 p-4">
        <h3 class="mb-3 text-lg font-semibold">库存材料</h3>
        <div class="flex flex-wrap items-center gap-3">
          <input v-model="materialName" type="text" placeholder="材料名称" class="border rounded px-3 py-2" />
          <input
            v-model.number="materialWidth"
            type="number"
            placeholder="宽(cm)"
            step="0.1"
            min="0.1"
            class="w-24 border rounded px-3 py-2"
          />
          <input
            v-model.number="materialHeight"
            type="number"
            placeholder="高(cm)"
            step="0.1"
            min="0.1"
            class="w-24 border rounded px-3 py-2"
          />
          <input
            v-model.number="materialCount"
            type="number"
            placeholder="数量"
            class="w-20 border rounded px-3 py-2"
            min="1"
          />
          <NButton type="primary" @click="addMaterial">添加材料</NButton>
        </div>
      </section>

      <!-- 剩余材料列表 -->
      <section class="mb-6">
        <h3 class="mb-2 text-lg font-semibold">剩余材料</h3>
        <table class="w-full border-collapse text-sm">
          <thead>
            <tr class="bg-gray-100">
              <th class="border px-3 py-2">名称</th>
              <th class="border px-3 py-2">宽(cm)</th>
              <th class="border px-3 py-2">高(cm)</th>
              <th class="border px-3 py-2">数量</th>
              <th class="border px-3 py-2">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(material, index) in materials" :key="index" class="hover:bg-gray-50">
              <td class="border px-3 py-2">{{ material.name }}</td>
              <td class="border px-3 py-2">{{ material.width.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ material.height.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ material.count }}</td>
              <td class="border px-3 py-2">
                <button class="text-sm text-red-600" @click="removeMaterial(index)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

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

<style scoped></style>
