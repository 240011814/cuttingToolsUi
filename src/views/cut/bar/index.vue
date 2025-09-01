<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { NButton, NCard, NDataTable, NInputNumber, NModal, NSpin } from 'naive-ui';
import { CutBar } from '@/service/api';

const itemsData = ref<{ length: number; qty: number }[]>([]);
const materialsData = ref<{ length: number; qty: number }[]>([]);

const itemLength = ref<number | null>(null);
const itemQty = ref<number | null>(null);
const matLength = ref<number | null>(null);
const matQty = ref<number | null>(null);
const newMaterialLength = ref(600);
const loss = ref(0.2);
const utilizationWeight = ref(4);

const cutResult = ref<Api.Cut.BarResult[] | null>(null);
const loading = ref(false);
const scaleFactor = ref(1);

const canvasWrapper = ref<HTMLDivElement | null>(null);
const containerWidth = ref(800); // 动态容器宽度

// item 表格
const itemColumns = [
  { title: '长度(cm)', key: 'length' },
  { title: '数量', key: 'qty' },
  {
    title: '操作',
    key: 'actions',
    render(_row: any, index: number) {
      return h(
        NButton,
        {
          size: 'small',
          type: 'error',
          onClick: () => removeFromList('items', index)
        },
        { default: () => '删除' }
      );
    }
  }
];

// material 表格
const materialColumns = [
  { title: '长度(cm)', key: 'length' },
  { title: '数量', key: 'qty' },
  {
    title: '操作',
    key: 'actions',
    render(_row: any, index: number) {
      return h(
        NButton,
        {
          size: 'small',
          type: 'error',
          onClick: () => removeFromList('materials', index)
        },
        { default: () => '删除' }
      );
    }
  }
];

function addItem() {
  if (itemLength.value && itemQty.value && itemQty.value > 0) {
    itemsData.value.push({ length: itemLength.value, qty: itemQty.value });
    itemLength.value = null;
    itemQty.value = null;
  }
}

function addMaterial() {
  if (matLength.value && matQty.value && matQty.value > 0) {
    materialsData.value.push({ length: matLength.value, qty: matQty.value });
    matLength.value = null;
    matQty.value = null;
  }
}

function removeFromList(list: 'items' | 'materials', index: number) {
  if (list === 'items') itemsData.value.splice(index, 1);
  else materialsData.value.splice(index, 1);
}

function clearAll() {
  itemsData.value = [];
  materialsData.value = [];
  cutResult.value = null;
}

// 获取数据
async function fetchData() {
  loading.value = true;
  const items: number[] = itemsData.value.flatMap(i => Array(i.qty).fill(i.length));

  const materials: number[] = materialsData.value.flatMap(i => Array(i.qty).fill(i.length));
  try {
    const data = await CutBar({
      items,
      materials,
      newMaterialLength: newMaterialLength.value,
      loss: loss.value,
      utilizationWeight: utilizationWeight.value
    });
    const { data: reslut } = data;
    cutResult.value = reslut;
  } catch {
  } finally {
    loading.value = false;
  }
}

// 统计信息
const result = computed(() => {
  if (!cutResult.value || cutResult.value.length === 0) {
    return {
      totalMaterials: 0,
      totalLength: 0,
      totalUsed: 0,
      totalRemaining: 0,
      usagePercent: '0.00'
    };
  }

  const totalMaterials = cutResult.value.length;
  let totalLength = 0;
  let totalUsed = 0;
  let totalRemaining = 0;

  cutResult.value.forEach(item => {
    totalLength += item.totalLength;
    totalUsed += item.used;
    totalRemaining += item.remaining;
  });

  return {
    totalMaterials,
    totalLength,
    totalUsed,
    totalRemaining,
    usagePercent: ((totalUsed / totalLength) * 100).toFixed(2)
  };
});

// 颜色池
const randomColors = Array.from({ length: 50 }, (_, i) => `hsl(${(i * 30) % 360}, 70%, 50%)`);

// 缩放 + 拖动
onMounted(() => {
  if (canvasWrapper.value) {
    containerWidth.value = canvasWrapper.value.clientWidth;

    let isDragging = false;
    let startX = 0;
    let scrollLeft = 0;

    canvasWrapper.value.addEventListener('wheel', e => {
      e.preventDefault();
      scaleFactor.value += e.deltaY * -0.001;
      scaleFactor.value = Math.min(Math.max(0.5, scaleFactor.value), 3);
    });
    canvasWrapper.value.addEventListener('mousedown', e => {
      isDragging = true;
      startX = e.pageX - canvasWrapper.value!.offsetLeft;
      scrollLeft = canvasWrapper.value!.scrollLeft;
    });
    canvasWrapper.value.addEventListener('mouseup', () => {
      isDragging = false;
    });
    canvasWrapper.value.addEventListener('mouseleave', () => {
      isDragging = false;
    });
    canvasWrapper.value.addEventListener('mousemove', e => {
      if (!isDragging) return;
      e.preventDefault();
      const x = e.pageX - canvasWrapper.value!.offsetLeft;
      const walk = (x - startX) * 1.5;
      canvasWrapper.value!.scrollLeft = scrollLeft - walk;
    });
  }
});
</script>

<template>
  <div class="p-4">
    <!-- 输入区域 -->
    <NCard title="材料裁剪可视化" size="large" class="mb-4">
      <h3>裁剪尺寸</h3>
      <div class="mb-2 flex items-center gap-2">
        <NInputNumber v-model:value="itemLength" placeholder="长度" class="w-40" />
        <NInputNumber v-model:value="itemQty" placeholder="数量" class="w-32" />
        <NButton type="primary" @click="addItem">添加尺寸</NButton>
      </div>
      <NDataTable :columns="itemColumns" :data="itemsData" />

      <h3 class="mt-6">材料库存</h3>
      <div class="mb-2 flex items-center gap-2">
        <NInputNumber v-model:value="matLength" placeholder="长度" class="w-40" />
        <NInputNumber v-model:value="matQty" placeholder="数量" class="w-32" />
        <NButton type="primary" @click="addMaterial">添加材料</NButton>
      </div>
      <NDataTable :columns="materialColumns" :data="materialsData" />

      <h3 class="mt-6">参数配置</h3>
      <div class="mb-4 flex items-center gap-6">
        <div class="flex items-center gap-2">
          <span class="w-24">新材料长度</span>
          <NInputNumber v-model:value="newMaterialLength" class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">切割损耗</span>
          <NInputNumber v-model:value="loss" class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">利用权重</span>
          <NSlider v-model:value="utilizationWeight" :min="1" :max="8" :step="0.1" class="w-40" />
        </div>
      </div>

      <div class="mt-4 flex gap-2">
        <NButton type="primary" @click="fetchData">开始裁剪</NButton>
        <NButton type="warning" @click="clearAll">清空所有</NButton>
      </div>
    </NCard>

    <!-- 结果统计 -->
    <NCard title="结果统计" size="large" class="mb-4">
      <p>
        材料总数: {{ result.totalMaterials }} 根 | 总长度: {{ result.totalLength }} cm | 已用长度:
        {{ result.totalUsed }} cm | 剩余长度: {{ result.totalRemaining }} cm | 使用率: {{ result.usagePercent }}%
      </p>
    </NCard>

    <!-- 裁剪图示 -->
    <NCard title="裁剪图示" size="large">
      <div ref="canvasWrapper" class="cursor-grab overflow-x-auto border border-gray-300 rounded-md p-4">
        <div class="origin-top-left" :style="{ transform: `scale(${scaleFactor})` }">
          <div v-for="item in cutResult" :key="item.index" class="mb-6">
            <!-- 标签 -->
            <div class="mb-1 font-bold">
              材料 #{{ item.index }} (总长: {{ item.totalLength }}cm, 已用: {{ item.used }}cm, 剩余:
              {{ item.remaining }}cm)
            </div>

            <!-- 条形图 -->
            <div class="h-10 flex">
              <div
                v-for="(cut, idx) in item.cuts"
                :key="idx"
                class="flex items-center justify-center border border-white text-xs text-white"
                :style="{
                  width:
                    cut * (containerWidth / Math.max(...(cutResult ? cutResult.map(d => d.totalLength) : [1]))) + 'px',
                  backgroundColor: randomColors[idx % randomColors.length]
                }"
              >
                {{ cut }}cm
              </div>

              <div
                v-if="item.remaining > 0"
                class="flex items-center justify-center bg-gray-300 text-xs text-black"
                :style="{
                  width:
                    item.remaining *
                      (containerWidth / Math.max(...(cutResult ? cutResult.map(d => d.totalLength) : [1]))) +
                    'px'
                }"
              >
                剩余{{ item.remaining }}cm
              </div>
            </div>
          </div>
        </div>
      </div>
    </NCard>

    <!-- 加载中弹窗 -->
    <NModal v-model:show="loading" preset="dialog" title="计算中...">
      <div class="flex flex-col items-center justify-center p-6">
        <NSpin size="large" />
        <div class="mt-3">正在计算，请稍候...</div>
      </div>
    </NModal>
  </div>
</template>
