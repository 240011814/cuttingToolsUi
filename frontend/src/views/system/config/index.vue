<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useMessage } from "naive-ui";
import { request } from "@/service/request";

defineOptions({ name: "SystemConfig" });

const message = useMessage();
const loading = ref(false);
const savingRegister = ref(false);
const saving2fa = ref(false);
const savingTelegram = ref(false);
const savingTimeout = ref(false);
const savingSmtp = ref(false);
const sendingTestEmail = ref(false);
const showTestEmailModal = ref(false);
const testEmailForm = ref({ email: "", subject: "", content: "" });
const registerEnabled = ref(false);
const admin2faEnabled = ref(false);
const telegramEnabled = ref(true);
const telegramBotToken = ref("");
const showTelegramToken = ref(false);
const telegramWebhookUrl = ref("");
const aiTimeoutMinutes = ref(5);
const aiTlsHandshakeTimeout = ref(15);
const aiResponseHeaderTimeout = ref(30);
const httpTimeoutSeconds = ref(30);
const smtpEnabled = ref(false);
const smtpHost = ref("");
const smtpPort = ref(587);
const smtpEncryption = ref("ssl");
const smtpUser = ref("");
const smtpPassword = ref("");
const showSmtpPassword = ref(false);
const smtpFrom = ref("");
const smtpFromName = ref("");

async function loadConfig() {
  loading.value = true;
  try {
    const { data } = await request<any[]>({
      url: "/api/admin/system-config",
    });
    if (data) {
      const registerConfig = data.find((c: any) => c.key === "register_enabled");
      registerEnabled.value = registerConfig?.value === "true";

      const admin2faConfig = data.find((c: any) => c.key === "admin_2fa_enabled");
      admin2faEnabled.value = admin2faConfig?.value === "true";

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

      const smtpEnabledConfig = data.find((c: any) => c.key === "smtp_enabled");
      smtpEnabled.value = smtpEnabledConfig?.value === "true";

      const smtpHostConfig = data.find((c: any) => c.key === "smtp_host");
      smtpHost.value = smtpHostConfig?.value || "";

      const smtpPortConfig = data.find((c: any) => c.key === "smtp_port");
      smtpPort.value = smtpPortConfig ? Number(smtpPortConfig.value) : 587;

      const smtpEncryptionConfig = data.find((c: any) => c.key === "smtp_encryption");
      smtpEncryption.value = smtpEncryptionConfig?.value || "ssl";

      const smtpUserConfig = data.find((c: any) => c.key === "smtp_user");
      smtpUser.value = smtpUserConfig?.value || "";

      const smtpPasswordConfig = data.find((c: any) => c.key === "smtp_password");
      smtpPassword.value = smtpPasswordConfig?.value || "";

      const smtpFromConfig = data.find((c: any) => c.key === "smtp_from");
      smtpFrom.value = smtpFromConfig?.value || "";

      const smtpFromNameConfig = data.find((c: any) => c.key === "smtp_from_name");
      smtpFromName.value = smtpFromNameConfig?.value || "";
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

async function handleToggleSmtp(val: boolean) {
  savingSmtp.value = true;
  try {
    await saveConfig("smtp_enabled", val ? "true" : "false", "SMTP邮件服务开关");
    message.success(val ? "SMTP 邮件服务已启用" : "SMTP 邮件服务已禁用");
  } catch (err: any) {
    smtpEnabled.value = !val;
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingSmtp.value = false;
  }
}

async function handleSaveSmtp() {
  savingSmtp.value = true;
  try {
    await saveConfig("smtp_host", smtpHost.value, "SMTP服务器地址");
    await saveConfig("smtp_port", String(smtpPort.value), "SMTP服务器端口");
    await saveConfig("smtp_encryption", smtpEncryption.value, "SMTP加密方式(none/ssl/starttls)");
    await saveConfig("smtp_user", smtpUser.value, "SMTP用户名");
    await saveConfig("smtp_password", smtpPassword.value, "SMTP密码");
    await saveConfig("smtp_from", smtpFrom.value, "发件人邮箱地址");
    await saveConfig("smtp_from_name", smtpFromName.value, "发件人显示名称");
    message.success("SMTP 配置已保存");
  } catch (err: any) {
    message.error(`保存失败: ${err?.message || "未知错误"}`);
  } finally {
    savingSmtp.value = false;
  }
}

function openTestEmailModal() {
  testEmailForm.value = { email: "", subject: "测试邮件", content: "这是一封 SMTP 邮件服务的测试邮件。\n\n如果您收到此邮件，说明 SMTP 配置正确。" };
  showTestEmailModal.value = true;
}

async function handleSendTestEmail() {
  if (!testEmailForm.value.email) {
    message.warning("请输入收件人邮箱");
    return;
  }
  sendingTestEmail.value = true;
  try {
    await request({
      url: "/api/admin/system-config/test-email",
      method: "post",
      data: {
        email: testEmailForm.value.email,
        subject: testEmailForm.value.subject,
        content: testEmailForm.value.content,
      },
    });
    message.success("测试邮件已发送，请检查收件箱");
    showTestEmailModal.value = false;
  } catch (err: any) {
    message.error(`发送失败: ${err?.message || "未知错误"}`);
  } finally {
    sendingTestEmail.value = false;
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

          <!-- SMTP 邮件配置 -->
          <div class="p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between mb-4">
              <div>
                <div class="font-bold text-gray-800 dark:text-gray-200">SMTP 邮件服务</div>
                <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  配置 SMTP 邮件服务器，用于发送系统通知邮件。
                </div>
              </div>
              <NSwitch
                v-model:value="smtpEnabled"
                :loading="savingSmtp"
                @update:value="handleToggleSmtp"
              >
                <template #checked>开启</template>
                <template #unchecked>关闭</template>
              </NSwitch>
            </div>
            <NForm label-placement="left" label-width="120">
              <NGrid :cols="2" :x-gap="12" :y-gap="8">
                <NFormItemGi label="SMTP 主机" path="smtpHost">
                  <NInput
                    v-model:value="smtpHost"
                    placeholder="smtp.example.com"
                    :disabled="!smtpEnabled"
                  />
                </NFormItemGi>
                <NFormItemGi label="SMTP 端口" path="smtpPort">
                  <NInputNumber
                    v-model:value="smtpPort"
                    :min="1"
                    :max="65535"
                    :disabled="!smtpEnabled"
                    size="small"
                  />
                </NFormItemGi>
                <NFormItemGi label="加密方式" path="smtpEncryption">
                  <NSelect
                    v-model:value="smtpEncryption"
                    :options="[
                      { label: 'SSL/TLS', value: 'ssl' },
                      { label: 'STARTTLS', value: 'starttls' },
                      { label: '无加密', value: 'none' }
                    ]"
                    :disabled="!smtpEnabled"
                  />
                </NFormItemGi>
                <NFormItemGi label="用户名" path="smtpUser">
                  <NInput
                    v-model:value="smtpUser"
                    placeholder="SMTP 用户名"
                    :disabled="!smtpEnabled"
                  />
                </NFormItemGi>
                <NFormItemGi label="密码" path="smtpPassword">
                  <NInput
                    v-model:value="smtpPassword"
                    :type="showSmtpPassword ? 'text' : 'password'"
                    placeholder="SMTP 密码"
                    :disabled="!smtpEnabled"
                  >
                    <template #suffix>
                      <div
                        class="cursor-pointer text-gray-400 hover:text-gray-600"
                        :class="showSmtpPassword ? 'i-mdi:eye-off' : 'i-mdi:eye'"
                        @click="showSmtpPassword = !showSmtpPassword"
                      />
                    </template>
                  </NInput>
                </NFormItemGi>
                <NFormItemGi label="发件人邮箱" path="smtpFrom">
                  <NInput
                    v-model:value="smtpFrom"
                    placeholder="noreply@example.com"
                    :disabled="!smtpEnabled"
                  />
                </NFormItemGi>
                <NFormItemGi label="发件人名称" path="smtpFromName">
                  <NInput
                    v-model:value="smtpFromName"
                    placeholder="系统通知"
                    :disabled="!smtpEnabled"
                  />
                </NFormItemGi>
              </NGrid>
              <NFormItem class="mt-4">
                <NSpace>
                  <NButton type="primary" :loading="savingSmtp" :disabled="!smtpEnabled" @click="handleSaveSmtp">
                    保存 SMTP 配置
                  </NButton>
                  <NButton type="info" :disabled="!smtpEnabled" @click="openTestEmailModal">
                    <template #icon>
                      <SvgIcon icon="mdi:email-fast-outline" />
                    </template>
                    发送测试邮件
                  </NButton>
                </NSpace>
              </NFormItem>
            </NForm>
          </div>
        </div>
      </NSpin>
    </NCard>

    <!-- 发送测试邮件弹窗 -->
    <NModal v-model:show="showTestEmailModal" preset="card" title="发送测试邮件" style="width: 500px">
      <NForm label-placement="left" label-width="80">
        <NFormItem label="收件邮箱" required>
          <NInput v-model:value="testEmailForm.email" placeholder="请输入收件人邮箱" />
        </NFormItem>
        <NFormItem label="邮件主题">
          <NInput v-model:value="testEmailForm.subject" placeholder="测试邮件" />
        </NFormItem>
        <NFormItem label="邮件内容">
          <NInput
            v-model:value="testEmailForm.content"
            type="textarea"
            placeholder="请输入邮件内容"
            :rows="6"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showTestEmailModal = false">取消</NButton>
          <NButton type="primary" :loading="sendingTestEmail" @click="handleSendTestEmail">发送</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
