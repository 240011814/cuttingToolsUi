<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useEcharts } from "@/hooks/common/echarts";
import { fetchDashboardStats } from "@/service/api";
import type { DashboardStats } from "@/service/api";

defineOptions({
  name: "LineChart",
});

const stats = ref<DashboardStats | null>(null);

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: {
    trigger: "axis",
    axisPointer: {
      type: "cross",
      label: {
        backgroundColor: "#6a7985",
      },
    },
  },
  legend: {
    data: ["训练次数"],
    top: "0",
  },
  grid: {
    left: "3%",
    right: "4%",
    bottom: "3%",
    top: "15%",
  },
  xAxis: {
    type: "category",
    boundaryGap: false,
    data: [] as string[],
  },
  yAxis: {
    type: "value",
  },
  series: [
    {
      color: "#8e9dff",
      name: "训练次数",
      type: "line",
      smooth: true,
      stack: "Total",
      areaStyle: {
        color: {
          type: "linear",
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            {
              offset: 0.25,
              color: "#8e9dff",
            },
            {
              offset: 1,
              color: "#fff",
            },
          ],
        },
      },
      emphasis: {
        focus: "series",
      },
      data: [] as number[],
    },
  ],
}));

async function loadData() {
  const { data, error } = await fetchDashboardStats();
  if (!error && data) {
    stats.value = data;
    updateOptions((opts) => {
      opts.xAxis.data = data.training_trend.map((item) => item.date.slice(5));
      opts.series[0].data = data.training_trend.map((item) => item.count);
      return opts;
    });
  }
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" title="近7天训练趋势">
    <div ref="domRef" class="h-360px overflow-hidden"></div>
  </NCard>
</template>

<style scoped></style>
