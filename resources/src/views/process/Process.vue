<template>
  <div class="toolbar">
    <ConfirmButton @confirm="startAll" color="#3CB371">全部启动</ConfirmButton>
    <ConfirmButton @confirm="killAll" color="#CD5555">全部停止</ConfirmButton>
    <v-btn
      size="small"
      variant="tonal"
      color="blue"
      @click="processCreateComponent?.createProcessDialog()"
      >创建<v-icon dark right> mdi-plus-circle </v-icon>
    </v-btn>
  </div>
  <v-container>
    <!-- 顶部工具栏 -->

    <!-- 主体网格 -->
    <div class="flex-grid">
      <div v-for="(i, v) in processData" :key="i.uuid" class="responsive-box">
        <ProcessCard :data="i" :index="v" />
      </div>
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
import ConfirmButton from "~/src/components/ConfirmButton.vue";
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
  border-bottom: 1px solid #eee; /* 轻量分隔线 */
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
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  text-align: center;
  padding: 10px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.responsive-box:hover {
  transform: translateY(-6px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}
</style>
