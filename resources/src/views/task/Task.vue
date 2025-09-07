<template>
  <v-card class="pa-4">
    <v-card-title>任务明细（单行）</v-card-title>

    <v-data-table
      :headers="headers"
      :items="items"
      :items-per-page="5"
      item-key="id"
      class="elevation-1"
    >
      <!-- 自定义列渲染 -->
      <template #item.nextId="{ item }">
        <span>{{ item.nextId === null ? "-" : item.nextId }}</span>
      </template>

      <template #item.enable="{ item }">
        <v-chip size="small" variant="tonal">{{
          item.enable ? "启用" : "未启用"
        }}</v-chip>
      </template>

      <template #item.apiEnable="{ item }">
        <v-chip size="small" variant="tonal">{{
          item.apiEnable ? "可用" : "不可用"
        }}</v-chip>
      </template>

      <template #item.running="{ item }">
        <v-chip size="small" variant="tonal">{{
          item.running ? "运行中" : "未运行"
        }}</v-chip>
      </template>

      <template #item.startTime="{ item }">
        <span>{{ formatStartTime(item.startTime) }}</span>
      </template>

      <template #item.key="{ item }">
        <code>{{ item.key }}</code>
      </template>

      <!-- 如果需要可以加 actions 列 -->
      <template #body.append>
        <!-- 这里可以放底部说明或按钮 -->
      </template>
    </v-data-table>
  </v-card>
</template>

<script setup>
import { ref } from "vue";

/**
 * 自定义列头：title -> 显示标题（中文），key -> 对应数据字段名
 * （Vuetify 3 的 headers 使用 title/key 风格）
 */
const headers = [
  { title: "任务ID", key: "id" },
  { title: "任务名", key: "processName" },
  { title: "流程 ID", key: "processId" },
  { title: "条件", key: "condition" },
  { title: "下一步 ID", key: "nextId" },
  { title: "触发进程", key: "targetName" },
  { title: "触发名称", key: "triggerName" },
  { title: "操作", key: "operation" },
  { title: "触发事件", key: "triggerEvent" },
  { title: "触发目标", key: "triggerTarget" },
  { title: "操作目标", key: "operationTarget" },
  { title: "定时任务", key: "cron" },
  { title: "是否启用", key: "enable" },
  { title: "API 可用", key: "apiEnable" },
  { title: "开始时间", key: "startTime" },
  { title: "状态", key: "running" },
];

const rawData = {
  id: 1,
  processId: 0,
  condition: 3,
  nextId: null,
  operation: 0,
  triggerEvent: null,
  triggerTarget: null,
  operationTarget: 112,
  cron: "",
  enable: false,
  apiEnable: true,
  key: "bhCSw2QDdu",
  processName: "",
  targetName: "push",
  triggerName: "",
  startTime: "0001-01-01T00:00:00Z",
  running: false,
};

// 单行数组（v-data-table 接受 items 数组）
const items = ref([rawData]);

/** 格式化开始时间（兼容 null / 空字符串） */
function formatStartTime(v) {
  if (!v) return "-";
  try {
    const d = new Date(v);
    if (Number.isNaN(d.getTime())) return v;
    return d.toLocaleString();
  } catch (e) {
    return v;
  }
}
</script>

<style scoped>
code {
  font-size: 0.85rem;
  background: rgba(0, 0, 0, 0.04);
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
