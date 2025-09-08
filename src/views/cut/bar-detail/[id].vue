<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { NCard, NDataTable, NInputNumber } from 'naive-ui';
const route = useRoute();
const request: Api.Cut.BarRequest & {
  rowItems: Api.Cut.BarItem[];
  rowMaterials: Api.Cut.BarItem[];
} = JSON.parse(route.query.request as string) as Api.Cut.BarRequest & {
  rowItems: Api.Cut.BarItem[];
  rowMaterials: Api.Cut.BarItem[];
};
const response: Api.Cut.BarResult[] = JSON.parse(route.query.response as string) as Api.Cut.BarResult[];
const itemsData = ref<Api.Cut.BarItem[]>(request.rowItems || []);
const materialsData = ref<Api.Cut.BarItem[]>(request.rowMaterials || []);

const newMaterialLength = ref(request.newMaterialLength || 600);
const loss = ref(request.loss || 0);
const utilizationWeight = ref(request.utilizationWeight || 1);
const group = ref(false);
const cutResult = ref<Api.Cut.BarResult[] | null>(response || null);
const scaleFactor = ref(1);
const canvasWrapper = ref<HTMLDivElement | null>(null);
const containerWidth = ref(800); // 动态容器宽度

// item 表格
const itemColumns = [
  { title: '长度(cm)', key: 'length' },
  { title: '数量', key: 'quantity' }
];

// material 表格
const materialColumns = [
  { title: '长度(cm)', key: 'length' },
  { title: '数量', key: 'quantity' }
];

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

  cutResult.value.forEach((item: Api.Cut.BarResult) => {
    totalLength += item.totalLength;
    totalUsed += item.used;
    totalRemaining += item.remaining;
  });

  return {
    totalMaterials,
    totalLength: totalLength.toFixed(2),
    totalUsed: totalUsed.toFixed(2),
    totalRemaining: totalRemaining.toFixed(2),
    usagePercent: ((totalUsed / totalLength) * 100).toFixed(2)
  };
});

const processedResult = computed(() => {
  if (!cutResult.value) return [];

  // 按 cuts + remaining 来归一化 key
  const map = new Map<string, any>();

  if (group.value === false) {
    return cutResult.value;
  }
  cutResult.value.forEach((item: Api.Cut.BarResult) => {
    const cutsKey = item.cuts
      .slice()
      .sort((a, b) => a - b)
      .join(',');
    const key = `${cutsKey}|${item.remaining}`;
    if (!map.has(key)) {
      map.set(key, { ...item, count: 1 });
    } else {
      map.get(key).count = map.get(key).count + 1;
    }
  });

  // 转成数组并排序 (例如按 remaining 从小到大)
  return Array.from(map.values()).sort((a, b) => a.remaining - b.remaining);
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

    canvasWrapper.value.addEventListener('wheel', (e: WheelEvent) => {
      e.preventDefault();
      scaleFactor.value += e.deltaY * -0.001;
      scaleFactor.value = Math.min(Math.max(0.5, scaleFactor.value), 3);
    });
    canvasWrapper.value.addEventListener('mousedown', (e: MouseEvent) => {
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
    canvasWrapper.value.addEventListener('mousemove', (e: MouseEvent) => {
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
      <NDataTable :columns="itemColumns" :data="itemsData" />
      <h3 class="mt-6">材料库存</h3>
      <NDataTable :columns="materialColumns" :data="materialsData" />

      <h3 class="mt-6">参数配置</h3>
      <div class="mb-4 flex items-center gap-6">
        <div class="flex items-center gap-2">
          <span class="w-24">新材料长度</span>
          <NInputNumber v-model:value="newMaterialLength" disabled class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">切割损耗</span>
          <NInputNumber v-model:value="loss" disabled class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">利用率权重</span>
          <NSlider v-model:value="utilizationWeight" disabled :min="1" :max="8" :step="0.1" class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">聚合显示</span>
          <NSwitch v-model:value="group" class="w-40" />
        </div>
      </div>

      <div class="mt-4 flex gap-2">
        <BarPrinter :data="cutResult" />
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
          <div v-for="item in processedResult" :key="item.index" class="mb-6">
            <!-- 标签 -->
            <div class="mb-1 font-bold">
              材料 #{{ item.index }} (总长: {{ item.totalLength }}cm, 已用: {{ item.used }}cm, 剩余:
              {{ item.remaining }}cm) * {{ item.count || 1 }} 根
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
  </div>
</template>
