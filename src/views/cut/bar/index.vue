<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import { NButton } from 'naive-ui';
import { CutBar } from '@/service/api';
const itemsData = ref<{ length: number; qty: number }[]>([]);
const materialsData = ref<{ length: number; qty: number }[]>([]);

const itemLength = ref<number | null>(null);
const itemQty = ref<number | null>(null);
const matLength = ref<number | null>(null);
const matQty = ref<number | null>(null);
const newMaterialLength = ref(600);

const loading = ref(false);
const stats = ref('');
const scaleFactor = ref(1);

const canvasWrapper = ref<HTMLDivElement | null>(null);
const bars = ref<HTMLDivElement | null>(null);

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
  stats.value = '';
  if (bars.value) bars.value.innerHTML = '';
}

async function fetchData() {
  loading.value = true;
  const items: number[] = [];
  itemsData.value.forEach(i => {
    for (let n = 0; n < i.qty; n++) items.push(i.length);
  });
  const materials: number[] = [];
  materialsData.value.forEach(m => {
    for (let n = 0; n < m.qty; n++) materials.push(m.length);
  });
  try {
    const data = await CutBar({ items, materials, newMaterialLength: newMaterialLength.value });
    renderBars(data);
  } catch (err) {
    console.error('请求失败:', err);
  } finally {
    loading.value = false;
  }
}

function renderBars(data: Api.Cut.BarResult[] | null) {
  if (!bars.value) return;
  bars.value.innerHTML = '';
  if (data === null || data.length === 0) {
    stats.value = '暂无数据';
    return;
  }

  const totalMaterials = data.length;
  let totalLength = 0;
  let totalUsed = 0;
  let totalRemaining = 0;

  data.forEach(item => {
    totalLength += item.totalLength;
    totalUsed += item.used;
    totalRemaining += item.remaining;
  });

  const usagePercent = ((totalUsed / totalLength) * 100).toFixed(2);

  stats.value = `
    材料总数: ${totalMaterials} 根 |
    总长度: ${totalLength} cm |
    已用长度: ${totalUsed} cm |
    剩余长度: ${totalRemaining} cm |
    使用率: ${usagePercent}%
  `;

  const containerWidth = bars.value.clientWidth || 800;
  const maxLength = Math.max(...data.map(d => d.totalLength));
  const scale = containerWidth / maxLength;

  data.forEach(item => {
    const barWrapper = document.createElement('div');
    barWrapper.className = 'mb-6';

    const label = document.createElement('div');
    label.className = 'font-bold mb-1';
    label.textContent = `材料 #${item.index} (总长: ${item.totalLength}cm, 已用: ${item.used}cm, 剩余: ${item.remaining}cm)`;
    barWrapper.appendChild(label);

    const bar = document.createElement('div');
    bar.className = 'flex h-10';

    item.cuts.forEach((cut: number) => {
      const cutDiv = document.createElement('div');
      cutDiv.style.backgroundColor = getRandomColor();
      cutDiv.style.width = `${cut * scale}px`;
      cutDiv.className = 'flex items-center justify-center text-white text-xs border border-white';
      cutDiv.textContent = `${cut}cm`;
      bar.appendChild(cutDiv);
    });

    if (item.remaining > 0) {
      const remDiv = document.createElement('div');
      remDiv.style.width = `${item.remaining * scale}px`;
      remDiv.className = 'flex items-center justify-center text-black text-xs bg-gray-300';
      remDiv.textContent = `剩余${item.remaining}cm`;
      bar.appendChild(remDiv);
    }

    barWrapper.appendChild(bar);
    bars.value?.appendChild(barWrapper);
  });
}

function getRandomColor() {
  return `hsl(${Math.random() * 360}, 70%, 50%)`;
}

onMounted(() => {
  if (canvasWrapper.value) {
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

      <h3 class="mt-6">新材料长度</h3>
      <NInputNumber v-model:value="newMaterialLength" class="w-40" />

      <div class="mt-4 flex gap-2">
        <NButton type="primary" @click="fetchData">开始裁剪</NButton>
        <NButton type="warning" @click="clearAll">清空所有</NButton>
      </div>
    </NCard>

    <NCard title="结果统计" size="large" class="mb-4">
      <div v-html="stats" />
    </NCard>

    <NCard title="裁剪图示" size="large">
      <div ref="canvasWrapper" class="cursor-grab overflow-x-auto border border-gray-300 rounded-md p-4">
        <div ref="bars" class="origin-top-left" :style="{ transform: `scale(${scaleFactor})` }"></div>
      </div>
    </NCard>

    <NModal v-model:show="loading" preset="dialog" title="计算中...">
      <div class="flex flex-col items-center justify-center p-6">
        <NSpin size="large" />
        <div class="mt-3">正在计算，请稍候...</div>
      </div>
    </NModal>
  </div>
</template>
