<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useMessage } from 'naive-ui';
import { CutBin } from '@/service/api';

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

interface ResultBin {
  binId: number;
  materialType?: string;
  materialWidth: number; // 米
  materialHeight: number; // 米
  utilization: number;
  pieces: {
    label: string;
    x: number; // 米
    y: number; // 米
    w: number; // 米
    h: number; // 米
  }[];
}

// 响应式数据
const label = ref('');
const width = ref<number | null>(null);
const height = ref<number | null>(null);
const quantity = ref(1);

const materialName = ref('');
const materialWidth = ref<number | null>(null);
const materialHeight = ref<number | null>(null);
const materialCount = ref(1);

const items = ref<Item[]>([]);
const materials = ref<Material[]>([]);
const results = ref<ResultBin[]>([]);

// 用于保存 canvas 引用
const canvases = ref<(HTMLCanvasElement | null)[]>([]);

const loading = ref(false);
// 统计计算
const totalItems = computed(() => {
  return items.value.reduce((sum, item) => sum + item.quantity, 0);
});

const totalItemArea = computed(() => {
  return items.value.reduce((sum, item) => sum + item.width * item.height * item.quantity, 0);
});

// 判断是否为旧材料
const isRemainderMaterial = (bin: ResultBin) => {
  if (!bin.materialType) return false;
  return materials.value.some(m => bin.materialType?.startsWith(m.name));
};

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

  const existingIndex = items.value.findIndex(item => item.label === label.value);
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

  const expandedItems = items.value.flatMap(item => {
    return Array.from({ length: item.quantity }, (_, i) => ({
      label: `${item.label}_${i + 1}`,
      width: item.width / 100,
      height: item.height / 100
    }));
  });

  const materialData = materials.value.map(m => ({
    name: m.name,
    width: m.width / 100,
    height: m.height / 100,
    availableCount: m.count
  }));

  try {
    loading.value = true;
    const data = await CutBin({
      items: expandedItems,
      materials: materialData
    });
    results.value = data;

    // 延迟绘制，确保 canvas 已渲染
    setTimeout(() => {
      drawAllBins();
    }, 100);
  } catch {
  } finally {
    loading.value = false;
  }

  // 绘制所有结果
  function drawAllBins() {
    const maxMaterialWidth = Math.max(...results.value.map(b => b.materialWidth));
    const maxMaterialHeight = Math.max(...results.value.map(b => b.materialHeight));
    const maxCanvasSize = 400;
    const scale = Math.min(maxCanvasSize / (maxMaterialWidth * 100), maxCanvasSize / (maxMaterialHeight * 100)) * 100;

    canvases.value.forEach((canvas, index) => {
      const bin = results.value[index];
      if (!canvas) return;

      const ctx = canvas.getContext('2d')!;
      const widthCm = bin.materialWidth * 100;
      const heightCm = bin.materialHeight * 100;
      const widthPx = (widthCm * scale) / 100;
      const heightPx = (heightCm * scale) / 100;

      canvas.width = widthPx;
      canvas.height = heightPx;

      const isRemainder = isRemainderMaterial(bin);

      // 背景
      ctx.fillStyle = isRemainder ? '#e8f5e8' : '#e3f2fd';
      ctx.fillRect(0, 0, widthPx, heightPx);

      // 边框
      ctx.strokeStyle = '#333';
      ctx.lineWidth = 2;
      ctx.strokeRect(0, 0, widthPx, heightPx);

      // 网格线 (10cm)
      ctx.strokeStyle = '#bbb';
      ctx.lineWidth = 1;
      for (let x = 0; x <= widthCm; x += 10) {
        const px = (x * scale) / 100;
        ctx.beginPath();
        ctx.moveTo(px, 0);
        ctx.lineTo(px, heightPx);
        ctx.stroke();
      }
      for (let y = 0; y <= heightCm; y += 10) {
        const py = (y * scale) / 100;
        ctx.beginPath();
        ctx.moveTo(0, py);
        ctx.lineTo(widthPx, py);
        ctx.stroke();
      }

      // 绘制每个 piece
      bin.pieces.forEach(piece => {
        const hue = Math.floor(Math.random() * 360);
        const color = `hsl(${hue}, 70%, 80%)`;

        const x = (piece.x * 100 * scale) / 100;
        const y = (piece.y * 100 * scale) / 100;
        const w = (piece.w * 100 * scale) / 100;
        const h = (piece.h * 100 * scale) / 100;

        ctx.fillStyle = color;
        ctx.fillRect(x, y, w, h);

        ctx.strokeStyle = '#333';
        ctx.lineWidth = 1;
        ctx.strokeRect(x, y, w, h);

        // 标签
        ctx.fillStyle = 'rgba(0,0,0,0.8)';
        const labelWidth = Math.min(w - 4, 120);
        ctx.fillRect(x + 2, y + 2, labelWidth, 36);

        ctx.fillStyle = 'white';
        ctx.font = '12px Arial';
        ctx.fillText(piece.label, x + 6, y + 16);

        const sizeText = `${(piece.w * 100).toFixed(1)}×${(piece.h * 100).toFixed(1)}cm`;
        ctx.font = '11px Arial';
        ctx.fillText(sizeText, x + 6, y + 30);
      });
    });
  }

  // 回车快捷键支持
  onMounted(() => {
    window.addEventListener('keypress', e => {
      if (e.key === 'Enter') {
        const active = document.activeElement;
        if (['INPUT', 'TEXTAREA'].includes(active?.tagName || '')) {
          if (active && ['label', 'width', 'height', 'quantity'].includes(active.id)) {
            addItem();
          } else if (
            active &&
            ['materialName', 'materialWidth', 'materialHeight', 'materialCount'].includes(active.id)
          ) {
            addMaterial();
          }
        }
      }
    });
  });
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

      <!-- 操作按钮 -->
      <div class="mb-6 flex gap-3">
        <NButton type="primary" @click="runOptimization">开始裁剪</NButton>
        <NButton type="warning" @click="clearAll">清空所有</NButton>
      </div>
    </NCard>

    <!-- 统计信息 -->
    <NCard title="结果统计" size="large" class="mb-4">
      <p>总项目数: {{ totalItems }} (面积: {{ totalItemArea.toFixed(1) }} cm²), 剩余材料: {{ materials.length }} 种</p>
    </NCard>

    <!-- 优化结果 -->
    <div v-if="results.length" id="bins" class="mt-8 space-y-6">
      <h3 class="text-xl font-semibold">优化结果: 使用 {{ results.length }} 块材料, 放置 {{ totalItems }} 个项目</h3>
      <div
        v-for="(bin, index) in results"
        :key="index"
        class="bin-card overflow-hidden border rounded-lg"
        :class="{
          'border-green-400 bg-green-50': isRemainderMaterial(bin),
          'border-blue-400 bg-blue-50': !isRemainderMaterial(bin)
        }"
      >
        <div class="bg-gray-100 p-3">
          <h3 class="text-gray-800 font-semibold">{{ bin.materialType || '材料' }}</h3>
          <p class="text-sm text-gray-600">
            ID: {{ bin.binId }} | 尺寸: {{ (bin.materialWidth * 100).toFixed(1) }}×{{
              (bin.materialHeight * 100).toFixed(1)
            }}cm | 利用率: {{ bin.utilization.toFixed(1) }}%
            <span v-if="isRemainderMaterial(bin)" class="text-green-600">♻️ 剩余材料</span>
            <span v-else class="text-blue-600">🆕 新材料</span>
          </p>
        </div>
        <canvas :ref="el => (canvases[index] = el as HTMLCanvasElement | null)" class="block bg-white"></canvas>
      </div>
    </div>
    <NModal v-model:show="loading" preset="dialog" title="计算中...">
      <div class="flex flex-col items-center justify-center p-6">
        <NSpin size="large" />
        <div class="mt-3">正在计算，请稍候...</div>
      </div>
    </NModal>
  </div>
</template>

<style scoped></style>
