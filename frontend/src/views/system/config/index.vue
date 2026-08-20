<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useMessage } from "naive-ui";
import { request } from "@/service/request";

defineOptions({ name: "SystemConfig" });

const message = useMessage();
const loading = ref(false);
const savingRegister = ref(false);
const saving2fa = ref(false);
const savingMem0 = ref(false);
const savingTelegram = ref(false);
const savingTimeout = ref(false);
const savingWebSearch = ref(false);
const webSearchApiKey = ref("");
const showWebSearchKey = ref(false);
const registerEnabled = ref(false);
const admin2faEnabled = ref(false);
const mem0Enabled = ref(true);
const mem0ApiKey = ref("");
const mem0BaseUrl = ref("");
const showApiKey = ref(false);
const telegramEnabled = ref(true);
const telegramBotToken = ref("");
const showTelegramToken = ref(false);
const telegramWebhookUrl = ref("");
const aiTimeoutMinutes = ref(5);
const aiTlsHandshakeTimeout = ref(15);
const aiResponseHeaderTimeout = ref(30);
const httpTimeoutSeconds = ref(30);

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

      const admin2faConfig = data.find((c: any) => c.key === "admin_2fa_enabled");
      admin2faEnabled.value = admin2faConfig?.value === "true";

      const apiKeyConfig = data.find((c: any) => c.key === "mem0_api_key");
      mem0ApiKey.value = apiKeyConfig?.value || "";

      const baseUrlConfig = data.find((c: any) => c.key === "mem0_base_url");
      mem0BaseUrl.value = baseUrlConfig?.value || "https://api.mem0.ai/v1";

      const telegramEnabledConfig = data.find((c: any) => c.key === "telegram_enabled");
      telegramEnabled.value = telegramEnabledConfig?.value !== "false";

      const telegramTokenConfig = data.find((c: any) => c.key === "telegram_bot_token");
      telegramBotToken.value = telegramTokenConfig?.value || "";

      const telegramWebhookUrlConfig = data.find((c: any) => c.key === "telegram_webhook_url");
      telegramWebhookUrl.value = telegramWebhookUrlConfig?.value || "";

      const aiTimeoutConfig = data.find((c: any) => c.key === "ai_timeout_minutes");
      aiTimeoutMinutes.value = aiTimeoutConfig ? Number(aiTimeoutConfig.value) : 5;

      const aiTlsConfig = data.find((c: any) => c.key === "ai_tls_handshake_timeout");
      aiTlsHandshakeTimeout.value = aiTlsConfig ? Number(aiTlsConfig.value) : 15;

      const aiResponseConfig = data.find((c: any) => c.key === "ai_response_header_timeout");
      aiResponseHeaderTimeout.value = aiResponseConfig ? Number(aiResponseConfig.value) : 30;

      const httpTimeoutConfig = data.find((c: any) => c.key === "http_timeout_seconds");
      httpTimeoutSeconds.value = httpTimeoutConfig ? Number(httpTimeoutConfig.value) : 30;

      const webSearchApiKeyConfig = data.find((c: any) => c.key === "web_search_api_key");
      webSearchApiKey.value = webSearchApiKeyConfig?.value || "";
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
  savingRegister.value = true;
  try {
    await saveConfig("register_enabled", val ? "true" : "false", "注册功能开关");
    message.success(val ? "注册功能已开启" : "注册功能已关闭");
  } catch (err: any) {
    registerEnabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingRegister.value = false;
  }
}

async function handleToggleAdmin2FA(val: boolean) {
  saving2fa.value = true;
  try {
    await saveConfig("admin_2fa_enabled", val ? "true" : "false", "管理员二次验证(TOTP)开关");
    message.success(val ? "管理员二次验证已开启" : "管理员二次验证已关闭");
  } catch (err: any) {
    admin2faEnabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    saving2fa.value = false;
  }
}

async function handleToggleMem0(val: boolean) {
  savingMem0.value = true;
  try {
    await saveConfig("mem0_enabled", val ? "true" : "false", "Mem0 记忆服务开关");
    message.success(val ? "Mem0 记忆服务已开启" : "Mem0 记忆服务已关闭");
  } catch (err: any) {
    mem0Enabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingMem0.value = false;
  }
}

async function handleSaveMem0() {
  savingMem0.value = true;
  try {
    await saveConfig("mem0_api_key", mem0ApiKey.value, "Mem0 API 密钥");
    await saveConfig("mem0_base_url", mem0BaseUrl.value, "Mem0 API 地址");
    message.success("Mem0 配置已保存并生效");
  } catch (err: any) {
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingMem0.value = false;
  }
}

async function handleToggleTelegram(val: boolean) {
  savingTelegram.value = true;
  try {
    await saveConfig("telegram_enabled", val ? "true" : "false", "Telegram Bot 开关");
    message.success(val ? "Telegram Bot 已启用" : "Telegram Bot 已禁用");
  } catch (err: any) {
    telegramEnabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingTelegram.value = false;
  }
}

async function handleSaveTelegram() {
  savingTelegram.value = true;
  try {
    await saveConfig("telegram_bot_token", telegramBotToken.value, "Telegram Bot Token");
    await saveConfig("telegram_webhook_url", telegramWebhookUrl.value, "Telegram Webhook 回调地址 (留空使用 Long Polling 模式)");
    message.success("Telegram 配置已保存，Bot 将自动重启");
  } catch (err: any) {
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingTelegram.value = false;
  }
}

async function handleSaveTimeout() {
  savingTimeout.value = true;
  try {
    await saveConfig("ai_timeout_minutes", String(aiTimeoutMinutes.value), "AI 请求超时时间（分钟）");
    await saveConfig("ai_tls_handshake_timeout", String(aiTlsHandshakeTimeout.value), "AI TLS 握手超时时间（秒）");
    await saveConfig("ai_response_header_timeout", String(aiResponseHeaderTimeout.value), "AI 响应头超时时间（秒）");
    await saveConfig("http_timeout_seconds", String(httpTimeoutSeconds.value), "HTTP 请求超时时间（秒）");
    message.success("超时配置已保存");
  } catch (err: any) {
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingTimeout.value = false;
  }
}

async function handleSaveWebSearch() {
  savingWebSearch.value = true;
  try {
    await saveConfig("web_search_api_key", webSearchApiKey.value, "Tavily 联网搜索 API Key");
    message.success("联网搜索配置已保存");
  } catch (err: any) {
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingWebSearch.value = false;
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
              :loading="savingRegister"
              @update:value="handleToggleRegister"
            >
              <template #checked>开启</template>
              <template #unchecked>关闭</template>
            </NSwitch>
          </div>

          <!-- Admin 2FA 开关 -->
          <div
            class="flex items-center justify-between p-4 rounded-lg border border-gray-200 dark:border-gray-700"
          >
            <div>
              <div class="font-bold text-gray-800 dark:text-gray-200">管理员二次验证 (2FA)</div>
              <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                开启后，超级管理员登录时需要输入 TOTP 动态验证码。首次开启时需扫描二维码绑定验证器。
              </div>
            </div>
            <NSwitch
              v-model:value="admin2faEnabled"
              :loading="saving2fa"
              @update:value="handleToggleAdmin2FA"
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
                :loading="savingMem0"
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
                <NButton type="primary" :loading="savingMem0" :disabled="!mem0Enabled" @click="handleSaveMem0">
                  保存 Mem0 配置
                </NButton>
              </NFormItem>
            </NForm>
          </div>

          <!-- Telegram Bot 配置 -->
          <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between mb-4">
              <div>
                <div class="font-bold text-gray-800 dark:text-gray-200">Telegram Bot</div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  启用后用户可绑定 Telegram 使用学习助手功能。配置 Webhook URL 后使用回调模式，留空使用 Long Polling 模式。
                </div>
              </div>
              <NSwitch
                v-model:value="telegramEnabled"
                :loading="savingTelegram"
                @update:value="handleToggleTelegram"
              >
                <template #checked>开启</template>
                <template #unchecked>关闭</template>
              </NSwitch>
            </div>
            <NForm label-placement="left" label-width="100">
              <NFormItem label="Bot Token">
                <NInput
                  v-model:value="telegramBotToken"
                  :type="showTelegramToken ? 'text' : 'password'"
                  placeholder="输入 Telegram Bot Token (从 @BotFather 获取)"
                  :disabled="!telegramEnabled"
                >
                  <template #suffix>
                    <div
                      class="cursor-pointer text-gray-400 hover:text-gray-600"
                      :class="showTelegramToken ? 'i-mdi:eye-off' : 'i-mdi:eye'"
                      @click="showTelegramToken = !showTelegramToken"
                    />
                  </template>
                </NInput>
              </NFormItem>
              <NFormItem label="Webhook URL">
                <NInput
                  v-model:value="telegramWebhookUrl"
                  placeholder="https://your-domain.com (留空使用 Long Polling)"
                  :disabled="!telegramEnabled"
                />
              </NFormItem>
              <NFormItem>
                <NButton type="primary" :loading="savingTelegram" :disabled="!telegramEnabled" @click="handleSaveTelegram">
                  保存 Telegram 配置
                </NButton>
              </NFormItem>
            </NForm>
          </div>

          <!-- 超时配置 -->
          <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="mb-4">
              <div class="font-bold text-gray-800 dark:text-gray-200">超时配置</div>
              <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                配置各服务的超时时间，修改后立即生效。
              </div>
            </div>
            <NForm label-placement="left" label-width="100">
              <NGrid :cols="2" :x-gap="12" :y-gap="8">
                <NFormItemGi label="AI 请求超时">
                  <NInputNumber
                    v-model:value="aiTimeoutMinutes"
                    :min="1"
                    :max="60"
                    :disabled="savingTimeout"
                    size="small"
                  >
                    <template #suffix>分钟</template>
                  </NInputNumber>
                </NFormItemGi>
                <NFormItemGi label="AI TLS 握手">
                  <NInputNumber
                    v-model:value="aiTlsHandshakeTimeout"
                    :min="5"
                    :max="120"
                    :disabled="savingTimeout"
                    size="small"
                  >
                    <template #suffix>秒</template>
                  </NInputNumber>
                </NFormItemGi>
                <NFormItemGi label="AI 响应头超时">
                  <NInputNumber
                    v-model:value="aiResponseHeaderTimeout"
                    :min="5"
                    :max="120"
                    :disabled="savingTimeout"
                    size="small"
                  >
                    <template #suffix>秒</template>
                  </NInputNumber>
                </NFormItemGi>
                <NFormItemGi label="HTTP 请求超时">
                  <NInputNumber
                    v-model:value="httpTimeoutSeconds"
                    :min="5"
                    :max="300"
                    :disabled="savingTimeout"
                    size="small"
                  >
                    <template #suffix>秒</template>
                  </NInputNumber>
                </NFormItemGi>
              </NGrid>
              <NFormItem class="mt-4">
                <NButton type="primary" :loading="savingTimeout" @click="handleSaveTimeout">
                  保存超时配置
                </NButton>
              </NFormItem>
            </NForm>
          </div>

          <!-- 联网搜索配置 -->
          <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between mb-4">
              <div>
                <div class="font-bold text-gray-800 dark:text-gray-200">联网搜索 (Tavily)</div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  配置 Tavily API Key 后，AI Agent 可在对话中调用联网搜索工具获取实时信息。
                </div>
              </div>
            </div>
            <NForm label-placement="left" label-width="100">
              <NFormItem label="API Key">
                <NInput
                  v-model:value="webSearchApiKey"
                  :type="showWebSearchKey ? 'text' : 'password'"
                  placeholder="输入 Tavily API Key (tvly-dev-...)"
                >
                  <template #suffix>
                    <div
                      class="cursor-pointer text-gray-400 hover:text-gray-600"
                      :class="showWebSearchKey ? 'i-mdi:eye-off' : 'i-mdi:eye'"
                      @click="showWebSearchKey = !showWebSearchKey"
                    />
                  </template>
                </NInput>
              </NFormItem>
              <NFormItem>
                <NButton type="primary" :loading="savingWebSearch" @click="handleSaveWebSearch">
                  保存联网搜索配置
                </NButton>
              </NFormItem>
            </NForm>
          </div>
        </div>
      </NSpin>
    </NCard>
  </div>
</template>
