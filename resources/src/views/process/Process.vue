<template>
  <!-- <div class="toolbar">

  </div> -->
  <v-container fluid class="py-6 px-8  rounded-lg">
    <!-- 顶部工具栏 -->
    <v-card
      class="pa-4 mb-8 rounded-2xl elevation-3"
    >
      <div class="d-flex align-center justify-space-between flex-wrap">
        <div class="d-flex align-center mb-2 mb-sm-0">
          <v-icon size="40" color="primary" class="mr-3">mdi-application-braces</v-icon>
          <span class="text-h5 font-weight-bold text-primary">进程管理</span>
        </div>

        <div class="d-flex align-center ga-3 flex-wrap">
          <ConfirmButton
            @confirm="startAll"
            color="success"
            variant="flat"
            class="rounded-lg px-4"
          >
            <v-icon start>mdi-play-circle</v-icon>
            全部启动
          </ConfirmButton>

          <ConfirmButton
            @confirm="killAll"
            color="error"
            variant="flat"
            class="rounded-lg px-4"
          >
            <v-icon start>mdi-stop-circle</v-icon>
            全部停止
          </ConfirmButton>

          <v-btn
            size="small"
            variant="flat"
            color="primary"
            class="rounded-lg px-4"
            @click="processCreateComponent?.createProcessDialog()"
          >
            <v-icon start>mdi-plus-circle</v-icon>
            创建
          </v-btn>
        </div>
      </div>
    </v-card>
    <!-- 主体网格 -->
    <div class="flex-grid">
      <v-card v-for="(i, v) in processData" :key="i.uuid" class="responsive-box">
        <ProcessCard :data="i" :index="v" />
      </v-card>
    </div>
  </v-container>
  <ProcessCreate ref="processCreateComponent"></ProcessCreate>
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
const uuid: string = crypto.randomUUID();
const snackbarStore = useSnackbarStore();
const initProcessData = () => {
  getProcessList().then((e) => {
    processData.value = e.data!.sort((a, b) => a.name.localeCompare(b.name));
    getProcessListWait();
  });
};

const startAll = () => {
  startProcessAll().then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("sucess");
    }
  });
};

const killAll = () => {
  killProcessAll().then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("sucess");
    }
  });
};

let cancelTokenSource: any;
const getProcessListWait = () => {
  cancelTokenSource = axios.CancelToken.source();
  axios
    .get("api/process/wait", {
      cancelToken: cancelTokenSource.token,
      headers: {
        Authorization: "bearer " + localStorage.getItem("token"),
        Uuid: uuid,
      },
    })
    .then((response) => {
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
