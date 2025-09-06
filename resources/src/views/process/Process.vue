<template>
  <v-container>
    <div class="flex-grid">
      <div v-for="(i, v) in processData" class="responsive-box">
        <ProcessCard :data="i" :index="v"></ProcessCard>
      </div>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import ProcessCard from "@/components/process/ProcessCard.vue";
import axios from "axios";
import { getProcessList } from "~/src/api/process";
import { ProcessItem } from "~/src/types/process/process";

const processData = ref<ProcessItem[]>();
const uuid: string = crypto.randomUUID();

const initProcessData = () => {
  getProcessList().then((e) => {
    processData.value = e.data!.sort((a, b) => a.name.localeCompare(b.name));
    getProcessListWait();
  });
};

var cancelTokenSource;
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
.flex-grid {
  display: flex;
  flex-wrap: wrap; /* 自动换行 */
  justify-content: space-between; /* 两边与中间间距均匀 */
  gap: 80px; /* 每个 div 之间的间距 */
}

.responsive-box {
  flex: 1 1 300px; /* 最小宽度 300px */
  min-width: 300px;
  max-width: 100%;
  background: #ffffff; /* 改为白色背景 */
  border-radius: 16px; /* 圆角 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08); /* 柔和阴影 */
  text-align: center;
  padding: 10px; /* 内边距，让内容不贴边 */
  transition: transform 0.2s ease, box-shadow 0.2s ease; /* 交互动画 */
}

/* 悬停效果 */
.responsive-box:hover {
  transform: translateY(-6px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
}
</style>
