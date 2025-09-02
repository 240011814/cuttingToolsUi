<script setup lang="ts">
import { nextTick, ref } from 'vue';
import html2canvas from 'html2canvas';

const props = defineProps<{
  results: Api.Cut.BinResult[];
  materials: { name: string; width: number; height: number; count: number }[];
}>();

const printArea = ref<HTMLDivElement | null>(null);

async function printResult() {
  if (!printArea.value || !props.results.length) return;

  const el = printArea.value;

  // 临时显示内容
  el.style.position = 'absolute';
  el.style.left = '0';
  el.style.top = '0';
  el.style.width = '190mm'; // A4 宽度
  el.style.display = 'block';
  el.style.opacity = '1';

  await nextTick();

  // 等待 PlaneCanvas 绘制完成
  await new Promise(resolve => {
    setTimeout(resolve, 100);
  });

  try {
    // html2canvas 生成高分辨率图片
    const canvas = await html2canvas(el, {
      scale: 2,
      useCORS: true,
      allowTaint: true,
      logging: false,
      backgroundColor: '#fff',
      width: el.scrollWidth,
      height: el.scrollHeight
    });

    const imgData = canvas.toDataURL('image/png');

    const printWindow = window.open('', '_blank');
    if (!printWindow) return;

    printWindow.document.write(`
      <html>
        <head>
          <title>打印下料图</title>
          <style>
            @page {
              size: A4 portrait;
              margin: 10mm;
            }
            body {
              margin: 0;
              padding: 10mm;
              background: #fff;
            }
            img {
              width: 100%;
              height: auto;
              display: block;
              page-break-after: auto; /* 避免空白页 */
            }
          </style>
        </head>
        <body>
          <img src="${imgData}" onload="window.focus(); window.print(); window.close();" />
        </body>
      </html>
    `);
    printWindow.document.close();
  } finally {
    // 恢复隐藏样式
    el.style.position = 'fixed';
    el.style.left = '-9999px';
    el.style.top = '-9999px';
    el.style.display = 'none';
    el.style.opacity = '0';
  }
}
</script>

<template>
  <!-- 打印按钮 -->
  <NButton type="primary" :disabled="!results || results.length === 0" @click="printResult">打印</NButton>

  <!-- 打印内容区域 -->
  <div ref="printArea" class="print-area">
    <PlaneStats :results="results" />
    <PlaneCanvas :group-data="true" :results="results" :materials="materials" />
  </div>
</template>

<style scoped>
@media print {
  .hidden-print {
    display: none !important;
  }
}

/* 屏幕上隐藏，但保持可渲染 */
.print-area {
  position: fixed;
  left: -9999px;
  top: -9999px;
  width: 800px;
  min-height: 200px;
  z-index: -9999;
  opacity: 0;
  pointer-events: none;
  box-sizing: border-box;
}
</style>
