<script setup lang="ts">
import { onMounted } from 'vue';
import { useEcharts } from '@/hooks/common/echarts';
import { fetchDashboardStats } from '@/service/api';

defineOptions({
  name: 'PieChart'
});

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: {
    trigger: 'item'
  },
  legend: {
    bottom: '1%',
    left: 'center',
    itemStyle: {
      borderWidth: 0
    }
  },
  series: [
    {
      color: ['#5da8ff', '#8e9dff', '#fedc69', '#26deca', '#ff6b6b'],
      name: '训练类型',
      type: 'pie',
      radius: ['45%', '75%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 1
      },
      label: {
        show: false,
        position: 'center'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: '12'
        }
      },
      labelLine: {
        show: false
      },
      data: [] as { name: string; value: number }[]
    }
  ]
}));

async function loadData() {
  const { data, error } = await fetchDashboardStats();
  if (!error && data) {
    updateOptions(opts => {
      opts.series[0].data = data.training_type_stats.map(item => ({
        name: item.type,
        value: item.count
      }));
      return opts;
    });
  }
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <NCard :bordered="false" class="card-wrapper" title="训练类型分布">
    <div ref="domRef" class="h-360px overflow-hidden"></div>
  </NCard>
</template>

<style scoped></style>
