<template>
  <v-container fluid class="py-6 px-8">
    <!-- 页面标题 -->
    <!-- 修改密码卡片 -->
    <v-card class="mb-6 rounded-2xl elevation-2">
      <v-card-title class="d-flex align-center">
        <v-icon color="warning" class="mr-2">mdi-lock-reset</v-icon>
        修改密码
      </v-card-title>
      <v-divider></v-divider>
      <v-card-text class="pa-6">
        <v-text-field
          v-model="newPasswd1"
          label="输入新密码"
          :append-inner-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
          :type="show1 ? 'text' : 'password'"
          @click:append-inner="show1 = !show1"
          variant="outlined"
          density="comfortable"
          class="mb-2"
          hint="密码长度不能小于4位"
          :rules="[rules.required, rules.min]"
        ></v-text-field>
        <v-text-field
          v-model="newPasswd2"
          label="确认新密码"
          :append-inner-icon="show2 ? 'mdi-eye' : 'mdi-eye-off'"
          :type="show2 ? 'text' : 'password'"
          @click:append-inner="show2 = !show2"
          variant="outlined"
          density="comfortable"
          :rules="[rules.required, rules.min, rules.match]"
        ></v-text-field>
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions class="pa-4">
        <v-btn
          color="primary"
          variant="flat"
          @click="changePasswd"
          :loading="passwordLoading"
        >
          <v-icon start>mdi-check</v-icon>
          修改密码
        </v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>

    <!-- 系统配置卡片 - 仅管理员可见 -->
    <v-card
      v-if="isAdmin"
      class="rounded-2xl elevation-2"
      :loading="configLoading"
    >
      <v-card-title class="d-flex align-center justify-space-between">
        <div class="d-flex align-center">
          <v-icon color="success" class="mr-2">mdi-tune</v-icon>
          系统配置
        </div>
        <v-btn color="primary" variant="tonal" @click="handleStorageReload" :loading="configLoading">
          <v-icon start>mdi-reload</v-icon>
          刷新存储引擎
        </v-btn>
      </v-card-title>
      <v-divider></v-divider>
      <v-card-text class="pa-4">
        <v-alert type="warning" variant="tonal" class="mb-4" density="compact">
          <v-icon start>mdi-alert</v-icon>
          部分配置需要重启后才能生效，错误的配置可能会导致崩溃
        </v-alert>

        <v-row dense>
            <v-col
              v-for="(item, index) in configList"
              :key="index"
              cols="6"
            >
            <v-card variant="outlined" class="pa-3 mb-2 config-item">
              <div class="d-flex align-center justify-space-between">
                <div>
                  <div class="text-subtitle-2 font-weight-medium">
                    {{ item.describe }}
                  </div>
                  <div class="text-caption text-grey">
                    {{ item.key }}
                  </div>
                </div>
                <div class="config-input">
                  <v-switch
                    v-if="typeof item.value === 'boolean'"
                    v-model="configForm[item.key]"
                    color="primary"
                    density="compact"
                    hide-details
                    inset
                  ></v-switch>
                  <v-text-field
                    v-else
                    v-model="configForm[item.key]"
                    variant="outlined"
                    density="compact"
                    hide-details
                    style="max-width: 200px"
                    :placeholder="'默认: ' + item.default"
                  ></v-text-field>
                </div>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions class="pa-4">
        <v-btn
          color="primary"
          variant="flat"
          @click="handleSetConfig"
          :loading="configLoading"
        >
          <v-icon start>mdi-content-save</v-icon>
          保存配置
        </v-btn>
        <v-btn color="grey" variant="tonal" @click="handleGetConfig">
          <v-icon start>mdi-refresh</v-icon>
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
    const account = localStorage.getItem("name") || "";
    const resp = await editUser({
      account,
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
.config-item {
  transition: all 0.2s ease;
}

.config-item:hover {
  border-color: rgb(var(--v-theme-primary));
  background-color: rgba(var(--v-theme-primary), 0.02);
}

.config-input {
  min-width: 220px;
}
</style>

