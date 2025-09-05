<script setup lang="ts">
import { ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const request: Api.Cut.BinRequest = JSON.parse(route.query.request as string) as Api.Cut.BinRequest;
const response: Api.Cut.BinResult[] = JSON.parse(route.query.response as string) as Api.Cut.BinResult[];

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

const group = ref(false);

const newMaterialHeight = ref(request.height || 200);
const newMaterialWidth = ref(request.width || 200);
const items = ref<Item[]>([]);
const materials = ref<Material[]>([]);
const results = ref<Api.Cut.BinResult[]>(response || []);
</script>

<template>
  <div class="p-4">
    <NCard title="材料裁剪可视化" size="large" class="mb-4">
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
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in items" :key="index" class="hover:bg-gray-50">
              <td class="border px-3 py-2">{{ item.label }}</td>
              <td class="border px-3 py-2">{{ item.width.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ item.height.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ item.quantity }}</td>
            </tr>
          </tbody>
        </table>
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
            </tr>
          </thead>
          <tbody>
            <tr v-for="(material, index) in materials" :key="index" class="hover:bg-gray-50">
              <td class="border px-3 py-2">{{ material.name }}</td>
              <td class="border px-3 py-2">{{ material.width.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ material.height.toFixed(1) }}</td>
              <td class="border px-3 py-2">{{ material.count }}</td>
            </tr>
          </tbody>
        </table>
      </section>

      <h3 class="mt-6">参数配置</h3>
      <div class="mb-4 flex items-center gap-6">
        <div class="flex items-center gap-2">
          <span class="w-24">新材料长度</span>
          <NInputNumber v-model:value="newMaterialHeight" disabled class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">新材料高度</span>
          <NInputNumber v-model:value="newMaterialWidth" disabled class="w-40" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24">聚合显示</span>
          <NSwitch v-model:value="group" class="w-40" />
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="mt-4 flex gap-2">
        <PlanePrinter :results="results" :materials="materials"></PlanePrinter>
      </div>
    </NCard>

    <PlaneStats :results="results"></PlaneStats>
    <PlaneCanvas :results="results" :group-data="group" :materials="materials"></PlaneCanvas>
  </div>
</template>

<style scoped></style>
