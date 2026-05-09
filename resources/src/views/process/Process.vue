<template>
  <v-container fluid class="py-6 px-8 rounded-lg">
    <!-- 主体网格 -->
    <div v-if="processData && processData.length > 0" class="flex-grid">
      <v-card
        v-for="(i, v) in processData"
        :key="i.uuid"
        class="responsive-box"
      >
        <ProcessCard :data="i" :index="v" />
      </v-card>
    </div>

    <!-- 空状态界面 -->
    <v-card
      v-else-if="initFirst"
      class="pa-12 rounded-2xl elevation-2 text-center"
      style="min-height: 400px"
    >
      <div
        class="d-flex flex-column align-center justify-center"
        style="height: 100%"
      >
        <v-icon size="120" color="grey-lighten-1" class="mb-6">
          mdi-application-outline
        </v-icon>
        <h3 class="text-h5 font-weight-medium text-grey-darken-1 mb-3">
          暂无进程
        </h3>
        <p v-permission="1" class="text-body-1 text-grey mb-8">先创建你的第一个进程</p>
        <v-btn v-permission="1"

          size="large"
          color="primary"
          variant="elevated"
          class="rounded-lg px-8"
          elevation="4"
          @click="processCreateComponent?.createProcessDialog()"
        >
          <v-icon start size="large" v-permission="1">mdi-plus-circle</v-icon>
          创建进程
        </v-btn>
      </div>
    </v-card>
  </v-container>

  <!-- 右下角悬浮按钮组 -->
  <div class="fab-wrapper">
    <v-speed-dial
      v-model="fab"
      location="bottom end"
      direction="top"
      transition="slide-y-reverse-transition"
      :open-on-hover="false"
    >
      <template v-slot:activator="{ props: activatorProps }">
        <v-btn
          v-bind="activatorProps"
          color="primary"
          size="large"
          icon
          elevation="8"
          class="fab-main"
        >
          <v-icon v-if="!fab" size="28">mdi-menu</v-icon>
          <v-icon v-else size="28">mdi-close</v-icon>
        </v-btn>
      </template>

      <v-tooltip location="start" text="创建进程">
        <template v-slot:activator="{ props: tooltipProps }">
          <v-btn
            v-permission="1"
            v-bind="tooltipProps"
            icon
            color="primary"
            size="default"
            elevation="6"
            @click="handleCreateClick"
            class="fab-action"
          >
            <v-icon size="24">mdi-plus-circle</v-icon>
          </v-btn>
        </template>
      </v-tooltip>

      <v-tooltip location="start" text="全部启动">
        <template v-slot:activator="{ props: tooltipProps }">
          <v-btn
            v-bind="tooltipProps"
            icon
            color="success"
            size="default"
            elevation="6"
            :loading="startingAll"
            :disabled="!processData || processData.length === 0"
            @click="confirmStartAll"
            class="fab-action"
          >
            <v-icon size="24">mdi-play-circle</v-icon>
          </v-btn>
        </template>
      </v-tooltip>

      <v-tooltip location="start" text="全部停止">
        <template v-slot:activator="{ props: tooltipProps }">
          <v-btn
            v-bind="tooltipProps"
            icon
            color="error"
            size="default"
            elevation="6"
            :loading="killingAll"
            :disabled="!processData || processData.length === 0"
            @click="confirmKillAll"
            class="fab-action"
          >
            <v-icon size="24">mdi-stop-circle</v-icon>
          </v-btn>
        </template>
      </v-tooltip>
    </v-speed-dial>
  </div>

  <ProcessCreate ref="processCreateComponent"></ProcessCreate>

  <!-- 全部启动确认对话框 -->
  <v-dialog v-model="startAllDialog" max-width="480">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium d-flex align-center">
        <v-icon color="success" class="mr-2">mdi-play-circle</v-icon>
        确认全部启动
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <div class="text-body-1 mb-3">确定要启动所有进程吗？</div>
        <div class="text-caption text-secondary">
          共 {{ processData?.length || 0 }} 个进程将被启动
        </div>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="startAllDialog = false">取消</v-btn>
        <v-btn color="success" @click="executeStartAll">确认启动</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- 全部停止确认对话框 -->
  <v-dialog v-model="killAllDialog" max-width="480">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium d-flex align-center">
        <v-icon color="error" class="mr-2">mdi-stop-circle</v-icon>
        确认全部停止
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <div class="text-body-1 mb-3">
          确定要停止所有进程吗？此操作将强制终止所有正在运行的进程。
        </div>
        <div class="text-caption text-error">
          共 {{ processData?.length || 0 }} 个进程将被停止
        </div>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="killAllDialog = false">取消</v-btn>
        <v-btn color="error" @click="executeKillAll">确认停止</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import ProcessCard from "@/components/process/ProcessCard.vue";
import axios from "axios";
import {
  getProcessList,
  killProcessAll,
  startProcessAll,
} from "~/src/api/process";
import ProcessCreate from "~/src/components/process/ProcessCreate.vue";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { ProcessItem } from "~/src/types/process/process";
type CreateHandle = {
  createProcessDialog: () => void;
  test: () => void;
};
const processCreateComponent = ref<CreateHandle | null>(null);
const processData = ref<ProcessItem[]>();

const snackbarStore = useSnackbarStore();
const startingAll = ref(false);
const killingAll = ref(false);
const startAllDialog = ref(false);
const killAllDialog = ref(false);
const fab = ref(false);
const initFirst = ref(false);

const version = ref(0);

const initProcessData = () => {
  getProcessList().then((e) => {
    processData.value = e.data?.sort((a, b) => a.name.localeCompare(b.name)) || [];
    initFirst.value = true;
    getProcessListWait();
  });
};

// 显示启动确认对话框
const confirmStartAll = () => {
  if (!processData.value || processData.value.length === 0) return;
  fab.value = false; // 关闭悬浮按钮菜单
  startAllDialog.value = true;
};

// 显示停止确认对话框
const confirmKillAll = () => {
  if (!processData.value || processData.value.length === 0) return;
  fab.value = false; // 关闭悬浮按钮菜单
  killAllDialog.value = true;
};

// 处理创建按钮点击
const handleCreateClick = () => {
  fab.value = false; // 关闭悬浮按钮菜单
  processCreateComponent.value?.createProcessDialog();
};

// 执行全部启动
const executeStartAll = () => {
  startAllDialog.value = false;
  startingAll.value = true;
  startProcessAll()
    .then((e) => {
      if (e.code === 0) {
        snackbarStore.showSuccessMessage("全部进程启动成功");
      } else {
        snackbarStore.showErrorMessage("启动失败");
      }
    })
    .catch(() => {
      snackbarStore.showErrorMessage("启动出错");
    })
    .finally(() => {
      startingAll.value = false;
    });
};

// 执行全部停止
const executeKillAll = () => {
  killAllDialog.value = false;
  killingAll.value = true;
  killProcessAll()
    .then((e) => {
      if (e.code === 0) {
        snackbarStore.showSuccessMessage("全部进程已停止");
      } else {
        snackbarStore.showErrorMessage("停止失败");
      }
    })
    .catch(() => {
      snackbarStore.showErrorMessage("停止出错");
    })
    .finally(() => {
      killingAll.value = false;
    });
};

let cancelTokenSource: any;

onBeforeUnmount(() => {
  if (cancelTokenSource) {
    cancelTokenSource.cancel("组件已销毁，取消请求");
  }
});

const getProcessListWait = () => {
  cancelTokenSource = axios.CancelToken.source();
  axios
    .get("api/process/wait", {
      cancelToken: cancelTokenSource.token,
      headers: {
        Authorization: "bearer " + localStorage.getItem("token"),
        Version: version.value,
      },
    })
    .then((response) => {
      version.value = parseInt(response.headers?.version || "0");
      processData.value = response.data.data.sort((a, b) =>
        a.name.localeCompare(b.name)
      );
      getProcessListWait();
    })
    .catch((error) => {
      console.error("请求错误:", error);
    });
};
onMounted(() => {
  initProcessData();
});
</script>

<style scoped>
/* 悬浮按钮包装器 - 固定定位 */
.fab-wrapper {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 1000;
}

/* 确保 v-speed-dial 不会改变位置 */
.fab-wrapper :deep(.v-speed-dial) {
  position: relative !important;
}

/* 子按钮列表 */
.fab-wrapper :deep(.v-speed-dial__list) {
  position: absolute !important;
  bottom: 76px !important;
  right: 0 !important;
  align-items: flex-end !important;
}

.fab-main {
  width: 64px !important;
  height: 64px !important;
  border-radius: 50% !important;
  transition: transform 0.3s ease, box-shadow 0.3s ease, rotate 0.3s ease !important;
}

.fab-main:hover {
  transform: scale(1.1) !important;
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.25) !important;
}

/* 悬浮按钮打开时的旋转动画 */
.fab-wrapper :deep(.v-speed-dial--active .v-btn--icon) {
  rotate: 45deg;
}

.fab-action {
  margin-bottom: 12px !important;
  transition: transform 0.2s ease, box-shadow 0.2s ease !important;
}

.fab-action:hover {
  transform: scale(1.15) !important;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.2) !important;
}

/* 工具栏样式 */
.toolbar {
  display: flex;
  justify-content: flex-end; /* 靠右对齐 */
  gap: 10px;
  margin-bottom: 20px;
  padding: 4px 0;
}

/* 原来的网格样式 */
.flex-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 80px;
}

.responsive-box {
  flex: 1 1 300px;
  min-width: 300px;
  max-width: 100%;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  text-align: center;
  padding: 10px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.responsive-box:hover {
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}
</style>
