<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  results: Api.Cut.BinResult[];
}>();

// 总项目数
const totalItems = computed(() =>
  props.results.reduce((sum: number, bin: Api.Cut.BinResult) => sum + bin.pieces.length, 0)
);

// 项目面积
const totalItemArea = computed(() =>
  props.results.reduce(
    (sum: number, bin: Api.Cut.BinResult) => sum + bin.pieces.reduce((s, p) => s + ((p.w / 100) * p.h) / 100, 0),
    0
  )
);

// 材料面积
const totalMaterialArea = computed(() =>
  props.results.reduce(
    (sum: number, bin: Api.Cut.BinResult) => sum + ((bin.materialWidth / 100) * bin.materialHeight) / 100,
    0
  )
);

// 总体利用率
const overallUtilization = computed(() => {
  if (totalMaterialArea.value === 0) return 0;
  return (totalItemArea.value / totalMaterialArea.value) * 100;
});
</script>

<template>
  <NCard v-if="results.length" title="结果统计" size="large" class="mb-4">
    <p>
      总项目数: {{ totalItems }} (面积: {{ totalItemArea.toFixed(1) }} m²) | 材料使用面积:
      {{ totalMaterialArea.toFixed(1) }} m² | 总体利用率: {{ overallUtilization.toFixed(1) }}% | 使用材料数:
      {{ results.length }}
    </p>
  </NCard>
</template>
