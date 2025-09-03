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
import { getProcessList } from "~/src/api/process";
import { ProcessItem } from "~/src/types/process/process";

const processData = ref<ProcessItem[]>();

const initProcessData = () => {
  getProcessList().then((e) => {
    processData.value = e.data!;
    console.log(e.data);
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
  gap: 50px; /* 每个 div 之间的间距 */
}

.responsive-box {
  flex: 1 1 300px; /* 最小宽度 500px */
  min-width: 300px; /* 强制最小宽度 */
  max-width: 100%; /* 不超过容器宽度 */
  background: #f5f5f5;
  padding: 16px;
  border-radius: 12px;
  text-align: center;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}
</style>
