<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useMessage } from "naive-ui";
import { request } from "@/service/request";

defineOptions({ name: "SystemConfig" });

const message = useMessage();
const loading = ref(false);
const saving = ref(false);
const registerEnabled = ref(false);
const mem0Enabled = ref(true);
const mem0ApiKey = ref("");
const mem0BaseUrl = ref("");
const showApiKey = ref(false);

async function loadConfig() {
  loading.value = true;
  try {
    const { data } = await request<any[]>({
      url: "/api/admin/system-config",
    });
    if (data) {
      const registerConfig = data.find((c: any) => c.key === "register_enabled");
      registerEnabled.value = registerConfig?.value === "true";

      const mem0EnabledConfig = data.find((c: any) => c.key === "mem0_enabled");
      mem0Enabled.value = mem0EnabledConfig?.value !== "false";

      const apiKeyConfig = data.find((c: any) => c.key === "mem0_api_key");
      mem0ApiKey.value = apiKeyConfig?.value || "";

      const baseUrlConfig = data.find((c: any) => c.key === "mem0_base_url");
      mem0BaseUrl.value = baseUrlConfig?.value || "https://api.mem0.ai/v1";
    }
  } catch (err: any) {
    message.error(`加载配置失败: ${err?.message || "未知错误"}`);
  } finally {
    loading.value = false;
  }
}

async function saveConfig(key: string, value: string, remark: string) {
  await request({
    url: "/api/admin/system-config",
    method: "put",
    data: { key, value, remark },
  });
}

async function handleToggleRegister(val: boolean) {
  saving.value = true;
  try {
    await saveConfig("register_enabled", val ? "true" : "false", "注册功能开关");
    message.success(val ? "注册功能已开启" : "注册功能已关闭");
  } catch (err: any) {
    registerEnabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    saving.value = false;
  }
}

async function handleToggleMem0(val: boolean) {
  saving.value = true;
  try {
    await saveConfig("mem0_enabled", val ? "true" : "false", "Mem0 记忆服务开关");
    message.success(val ? "Mem0 记忆服务已开启" : "Mem0 记忆服务已关闭");
  } catch (err: any) {
    mem0Enabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    saving.value = false;
  }
}

async function handleSaveMem0() {
  saving.value = true;
  try {
    await saveConfig("mem0_api_key", mem0ApiKey.value, "Mem0 API 密钥");
    await saveConfig("mem0_base_url", mem0BaseUrl.value, "Mem0 API 地址");
    message.success("Mem0 配置已保存并生效");
  } catch (err: any) {
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  loadConfig();
});
</script>

<template>
  <div class="h-full overflow-auto p-6">
    <NCard :bordered="false" shadow="sm" title="系统配置">
      <NSpin :show="loading">
        <div class="space-y-6">
          <!-- 注册开关 -->
          <div
            class="flex items-center justify-between p-4 rounded-lg border border-gray-200 dark:border-gray-700"
          >
            <div>
              <div class="font-bold text-gray-800 dark:text-gray-200">注册功能</div>
              <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                控制登录页面是否显示注册按钮。关闭后新用户无法自行注册。
              </div>
            </div>
            <NSwitch
              v-model:value="registerEnabled"
              :loading="saving"
              @update:value="handleToggleRegister"
            >
              <template #checked>开启</template>
              <template #unchecked>关闭</template>
            </NSwitch>
          </div>

          <!-- Mem0 配置 -->
          <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between mb-4">
              <div>
                <div class="font-bold text-gray-800 dark:text-gray-200">Mem0 记忆服务</div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  关闭后将停止所有记忆功能，包括对话记忆保存和记忆搜索。
                </div>
              </div>
              <NSwitch
                v-model:value="mem0Enabled"
                :loading="saving"
                @update:value="handleToggleMem0"
              >
                <template #checked>开启</template>
                <template #unchecked>关闭</template>
              </NSwitch>
            </div>
            <NForm label-placement="left" label-width="100">
              <NFormItem label="API Key">
                <NInput
                  v-model:value="mem0ApiKey"
                  :type="showApiKey ? 'text' : 'password'"
                  placeholder="输入 Mem0 API Key"
                  :disabled="!mem0Enabled"
                >
                  <template #suffix>
                    <div
                      class="cursor-pointer text-gray-400 hover:text-gray-600"
                      :class="showApiKey ? 'i-mdi:eye-off' : 'i-mdi:eye'"
                      @click="showApiKey = !showApiKey"
                    />
                  </template>
                </NInput>
              </NFormItem>
              <NFormItem label="Base URL">
                <NInput
                  v-model:value="mem0BaseUrl"
                  placeholder="https://api.mem0.ai/v1"
                  :disabled="!mem0Enabled"
                />
              </NFormItem>
              <NFormItem>
                <NButton type="primary" :loading="saving" :disabled="!mem0Enabled" @click="handleSaveMem0">
                  保存 Mem0 配置
                </NButton>
              </NFormItem>
            </NForm>
          </div>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>
