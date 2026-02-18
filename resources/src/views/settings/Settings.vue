<template>
  <v-container fluid class="settings-container pa-4 pa-sm-6 pa-md-8">
    <!-- 修改密码卡片 -->
    <v-card class="mb-4 mb-sm-6 rounded-xl elevation-2">
      <v-card-title class="d-flex align-center py-3 py-sm-4">
        <v-icon color="warning" class="mr-2">mdi-lock-reset</v-icon>
        <span class="text-h6 text-sm-h5">修改密码</span>
      </v-card-title>
      <v-divider></v-divider>
  <v-card-text class="pa-3 pa-sm-4 pa-md-6 d-flex flex-column align-center">
    <div class="password-input-wrapper">
      <v-text-field
        v-model="newPasswd1"
        label="输入新密码"
        :append-inner-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
        :type="show1 ? 'text' : 'password'"
        @click:append-inner="show1 = !show1"
        variant="outlined"
        density="comfortable"
        class="password-input mb-2"
        hint="密码长度不能小于4位"
        :rules="[rules.required, rules.min]"
        persistent-hint
      ></v-text-field>
      <v-text-field
        v-model="newPasswd2"
        label="确认新密码"
        :append-inner-icon="show2 ? 'mdi-eye' : 'mdi-eye-off'"
        :type="show2 ? 'text' : 'password'"
        @click:append-inner="show2 = !show2"
        variant="outlined"
        density="comfortable"
        class="password-input"
        :rules="[rules.required, rules.min, rules.match]"
        persistent-hint
      ></v-text-field>
    </div>
  </v-card-text>
      <v-divider></v-divider>
      <v-card-actions class="pa-3 pa-sm-4">
        <v-btn
          color="primary"
          variant="flat"
          @click="changePasswd"
          :loading="passwordLoading"
          class="text-none"
          prepend-icon="mdi-check"
        >
          修改密码
        </v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>

    <!-- 系统配置卡片 - 仅管理员可见 -->
    <v-card
      v-if="isAdmin"
      class="rounded-xl elevation-2"
      :loading="configLoading"
    >
      <v-card-title class="d-flex align-center justify-space-between py-3 py-sm-4 flex-wrap gap-2">
        <div class="d-flex align-center">
          <v-icon color="success" class="mr-2">mdi-tune</v-icon>
          <span class="text-h6 text-sm-h5">系统配置</span>
        </div>
        <v-btn
          color="primary"
          variant="tonal"
          @click="handleStorageReload"
          :loading="configLoading"
          class="text-none"
          prepend-icon="mdi-reload"
          size="small"
        >
          刷新存储引擎
        </v-btn>
      </v-card-title>
      <v-divider></v-divider>
      <v-card-text class="pa-3 pa-sm-4 pa-md-6">
        <v-alert
          type="warning"
          variant="tonal"
          class="mb-4"
          density="compact"
          icon="mdi-alert"
        >
          部分配置需要重启后才能生效，错误的配置可能会导致崩溃
        </v-alert>

        <v-row dense>
          <v-col
            v-for="(item, index) in configList"
            :key="index"
            cols="12"
            sm="6"
            md="4"
            lg="3"
          >
            <v-card
              variant="outlined"
              class="pa-3 config-item fill-height"
            >
              <div class="d-flex align-center justify-space-between flex-wrap gap-2">
                <div style="flex: 1; min-width: 120px;">
                  <div class="text-subtitle-2 font-weight-medium text-break">
                    {{ item.describe }}
                  </div>
                  <div class="text-caption text-grey mt-1 text-break">
                    {{ item.key }}
                  </div>
                </div>
                <div class="config-input" style="flex: 0 0 auto;">
                  <v-switch
                    v-if="typeof item.value === 'boolean'"
                    v-model="configForm[item.key]"
                    color="primary"
                    density="compact"
                    hide-details
                    inset
                    class="ma-0"
                  ></v-switch>
                  <v-text-field
                    v-else
                    v-model="configForm[item.key]"
                    variant="outlined"
                    density="compact"
                    hide-details
                    style="min-width: 180px; max-width: 260px;"
                    :placeholder="'默认: ' + item.default"
                    class="config-text-field"
                  ></v-text-field>
                </div>
              </div>
            </v-card>
          </v-col>
        </v-row>

        <!-- 空状态提示 -->
        <v-row v-if="configList.length === 0" class="mt-4">
          <v-col cols="12" class="text-center text-grey">
            <v-icon size="48" class="mb-2">mdi-database-off</v-icon>
            <div>暂无配置项</div>
          </v-col>
        </v-row>
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions class="pa-3 pa-sm-4 flex-wrap gap-2">
        <v-btn
          color="primary"
          variant="flat"
          @click="handleSetConfig"
          :loading="configLoading"
          class="text-none"
          prepend-icon="mdi-content-save"
        >
          保存配置
        </v-btn>
        <v-btn
          color="grey"
          variant="tonal"
          @click="handleGetConfig"
          class="text-none"
          prepend-icon="mdi-refresh"
        >
          刷新配置
        </v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { getConfig, setConfig, configReload } from "~/src/api/config";
import { editUser } from "~/src/api/user";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const snackbarStore = useSnackbarStore();

// 密码相关
const show1 = ref(false);
const show2 = ref(false);
const newPasswd1 = ref("");
const newPasswd2 = ref("");
const passwordLoading = ref(false);

// 配置相关
const configLoading = ref(false);
const configList = ref<Array<{
  key: string;
  value: any;
  default: any;
  describe: string;
}>>([]);
const configForm = ref<Record<string, any>>({});

// 验证规则
const rules = {
  required: (v: string) => !!v || "必填项",
  min: (v: string) => (v && v.length >= 4) || "密码长度至少4位",
  match: () => newPasswd1.value === newPasswd2.value || "两次密码不一致",
};

// 判断是否为管理员 (role === 0)
const isAdmin = computed(() => {
  return localStorage.getItem("role") === "0";
});

// 修改密码
const changePasswd = async () => {
  if (newPasswd1.value !== newPasswd2.value) {
    snackbarStore.showErrorMessage("两次密码不同");
    return;
  }

  if (newPasswd1.value.length < 4) {
    snackbarStore.showErrorMessage("密码长度至少4位");
    return;
  }

  passwordLoading.value = true;
  try {
    const resp = await editUser({
      password: newPasswd1.value,
    });
    if (resp.code === 0) {
      snackbarStore.showSuccessMessage("密码修改成功");
      newPasswd1.value = "";
      newPasswd2.value = "";
    }
  } catch (error) {
    console.error("修改密码错误:", error);
    snackbarStore.showErrorMessage("修改密码失败");
  } finally {
    passwordLoading.value = false;
  }
};

// 获取配置
const handleGetConfig = async () => {
  configLoading.value = true;
  try {
    const resp = await getConfig();
    if (resp.code === 0 && resp.data) {
      configList.value = resp.data as typeof configList.value;
      configForm.value = {};
      for (const item of configList.value) {
        configForm.value[item.key] = item.value;
      }
    }
  } catch (error) {
    console.error("获取配置错误:", error);
    snackbarStore.showErrorMessage("获取配置失败");
  } finally {
    configLoading.value = false;
  }
};

// 保存配置
const handleSetConfig = async () => {
  configLoading.value = true;
  try {
    // 将所有值转为字符串
    const data: Record<string, string> = {};
    for (const key in configForm.value) {
      data[key] = String(configForm.value[key]);
    }
    const resp = await setConfig(data);
    if (resp.code === 0) {
      snackbarStore.showSuccessMessage("配置修改成功");
      handleGetConfig();
    }
  } catch (error) {
    console.error("保存配置错误:", error);
    snackbarStore.showErrorMessage("保存配置失败");
  } finally {
    configLoading.value = false;
  }
};

// 重载 ES
const handleStorageReload = async () => {
  configLoading.value = true;
  try {
    const resp = await configReload();
    if (resp.code === 0) {
      snackbarStore.showSuccessMessage("已连接上存储引擎");
    }
  } catch (error) {
    console.error("重载存储引擎错误:", error);
    snackbarStore.showErrorMessage("重载存储引擎失败");
  } finally {
    configLoading.value = false;
  }
};

// 初始化
onMounted(() => {
  if (isAdmin.value) {
    handleGetConfig();
  }
});
</script>

<style scoped>
.settings-container {
  max-width: 1600px;
  margin: 0 auto;
}

.config-item {
  transition: all 0.2s ease;
  height: 100%;
  min-height: 80px;
}

.config-item:hover {
  border-color: rgb(var(--v-theme-primary));
  background-color: rgba(var(--v-theme-primary), 0.02);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

/* 响应式调整 */
@media (max-width: 600px) {
  .config-item {
    min-height: 70px;
  }

  .config-text-field {
    min-width: 120px !important;
    font-size: 14px;
  }
}

@media (max-width: 960px) {
  .gap-2 {
    gap: 8px !important;
  }

  .flex-wrap {
    flex-wrap: wrap;
  }
}

/* 确保文本在小屏幕上不会溢出 */
.text-break {
  word-break: break-word;
  overflow-wrap: break-word;
}

/* 调整输入框样式 */
:deep(.v-input__details) {
  font-size: 12px;
}

/* 优化卡片内边距响应式 */
@media (max-width: 600px) {
  :deep(.v-card-title) {
    font-size: 1rem !important;
  }

  :deep(.v-card-text) {
    padding: 12px !important;
  }

  :deep(.v-card-actions) {
    padding: 12px !important;
  }
}

/* 优化按钮在小屏幕上的显示 */
:deep(.v-btn) {
  letter-spacing: 0;
}

/* 确保警告框提示文字清晰可见 */
:deep(.v-alert) {
  font-size: 0.85rem;
}

/* 密码输入框样式 */
.password-input-wrapper {
  width: 100%;
  max-width: 900px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.password-input {
  width: 100%;
}

/* 优化网格列在小屏幕上的间距 */
.v-row.dense {
  --v-grid-gutter: 8px;
}

@media (max-width: 600px) {
  .v-row.dense {
    --v-grid-gutter: 4px;
  }
}
</style>
