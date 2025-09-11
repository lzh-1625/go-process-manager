<template>
  <v-card class="pa-4">
    <v-card-title>任务明细（单行）</v-card-title>

    <v-data-table
      :headers="headers"
      :items="taskData"
      :items-per-page="5"
      item-key="id"
      class="elevation-1"
    >
      <!-- 自定义列渲染 -->
      <template #item.nextId="{ item }">
        <span>{{ item.nextId === null ? "-" : item.nextId }}</span>
      </template>

      <template #item.enable="{ item }">
        <v-switch
          color="primary"
          @change="changeEnable(item)"
          v-model="item.enable"
        ></v-switch>
      </template>

      <template #item.apiEnable="{ item }">
        <v-switch
          color="primary"
          @change="edit(item)"
          v-model="item.apiEnable"
        ></v-switch>
      </template>

      <template #item.running="{ item }">
        <svg
          v-if="item.running"
          width="20"
          height="20"
          viewBox="0 0 48 48"
          fill="#000000"
        >
          <path
            fill-rule="evenodd"
            clip-rule="evenodd"
            d="M25 34a1 1 0 011 1v10a1 1 0 01-1 1h-2a1 1 0 01-1-1V35a1 1 0 011-1h2zm8.192-3.636l7.072 7.071a1 1 0 010 1.414l-1.415 1.415a1 1 0 01-1.414 0l-7.071-7.072a1 1 0 010-1.414l1.414-1.414a1 1 0 011.414 0zm-16.97 0l1.414 1.414a1 1 0 010 1.414l-7.071 7.072a1 1 0 01-1.414 0l-1.414-1.415a1 1 0 010-1.414l7.07-7.071a1 1 0 011.415 0zM45 22a1 1 0 011 1v2a1 1 0 01-1 1H35a1 1 0 01-1-1v-2a1 1 0 011-1h10zm-32 0a1 1 0 011 1v2a1 1 0 01-1 1H3a1 1 0 01-1-1v-2a1 1 0 011-1h10zM10.565 7.737l7.071 7.07a1 1 0 010 1.415l-1.414 1.414a1 1 0 01-1.414 0l-7.071-7.071a1 1 0 010-1.414L9.15 7.737a1 1 0 011.414 0zm28.284 0l1.415 1.414a1 1 0 010 1.414l-7.072 7.071a1 1 0 01-1.414 0l-1.414-1.414a1 1 0 010-1.414l7.071-7.071a1 1 0 011.414 0zM25 2a1 1 0 011 1v10a1 1 0 01-1 1h-2a1 1 0 01-1-1V3a1 1 0 011-1h2z"
            fill="#000000"
          />
        </svg>
        <svg v-else width="20" height="20" viewBox="0 0 48 48" fill="#000000">
          <path
            d="M42.02 12.71l-1.38-1.42a1 1 0 00-1.41 0L18.01 32.5l-9.9-9.89a1 1 0 00-1.41 0l-1.41 1.41a1 1 0 000 1.41l10.6 10.61 1.42 1.41a1 1 0 001.41 0l1.41-1.41 21.92-21.92a1 1 0 00-.03-1.41z"
            fill="#000000"
          />
        </svg>
      </template>

      <template #item.startTime="{ item }">
        <span>{{ formatStartTime(item.startTime) }}</span>
      </template>

      <template #item.key="{ item }">
        <code>{{ item.key }}</code>
      </template>

      <!-- 如果需要可以加 actions 列 -->
      <template #item.operate="{ item }">
        <!-- 这里可以放底部说明或按钮 -->
        <v-icon class="mr-2" v-if="!item.running"> mdi-play </v-icon>
        <v-icon class="mr-2" v-else> mdi-stop </v-icon>
        <v-icon class="mr-2"> mdi-pencil </v-icon>
        <v-icon> mdi-delete </v-icon>
      </template>
    </v-data-table>
  </v-card>

  <v-dialog v-model="taskDialog" max-width="500px">
    <v-card>
      <v-card-title class="text-h5">{{
        addTask ? "添加任务" : "修改任务"
      }}</v-card-title>
      <v-card-text style="margin-top: 20px">
        <v-autocomplete
          label="判断条件"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="conditionSelect"
          v-model="taskForm.condition"
        ></v-autocomplete>
        <v-autocomplete
          v-if="taskForm.condition != 3"
          label="判断目标"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="processSelect"
          v-model="taskForm.processId"
        ></v-autocomplete>
        <v-autocomplete
          label="操作目标"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="processSelect"
          v-model="taskForm.operationTarget"
        ></v-autocomplete>
        <v-autocomplete
          label="执行操作"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="operationSelect"
          v-model="taskForm.operation"
        ></v-autocomplete>
        <v-autocomplete
          label="触发目标"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="processSelect"
          v-model="taskForm.triggerTarget"
        ></v-autocomplete>
        <v-autocomplete
          v-if="taskForm.triggerTarget != null"
          label="触发事件"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="eventSelect"
          v-model="taskForm.triggerEvent"
        ></v-autocomplete>
        <v-autocomplete
          label="后续任务"
          item-text="name"
          item-value="value"
          filled
          dense
          :items="taskSelect"
          v-model="taskForm.nextId"
        ></v-autocomplete>
        <v-text-field
          label="定时任务"
          filled
          dense
          v-model="taskForm.cron"
        ></v-text-field>
        <v-text-field
          @click="copyToClipboard"
          :disabled="taskForm?.key == null"
          label="api"
          filled
          dense
          readonly
          :value="taskForm.key != null ? urlBase + taskForm.key : '未创建api'"
        ></v-text-field>

        <v-btn @click="changeApi" color="primary">
          {{ taskForm?.key != null ? "刷新api" : "创建api" }}
          <v-icon right dark>
            {{ taskForm?.key != null ? "mdi-refresh" : "mdi-plus" }}</v-icon
          >
        </v-btn>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="blue darken-1" text @click="taskDialog = false"
          >取消</v-btn
        >
        <v-btn color="blue darken-1" text @click="submit">确认</v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  changeTaskKey,
  editTask,
  editTaskEnable,
  getTaskAll,
  getTaskById,
} from "~/src/api/task";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { TaskItem } from "~/src/types/tassk/task";

const snackbarStore = useSnackbarStore();
const taskDialog = ref(false);
const taskForm = ref<Partial<TaskItem>>({});
const addTask = ref(false);
/**
 * 自定义列头：title -> 显示标题（中文），key -> 对应数据字段名
 * （Vuetify 3 的 headers 使用 title/key 风格）
 */
const headers = [
  { title: "任务ID", key: "id" },
  { title: "任务名", key: "processName" },
  { title: "进程ID", key: "processId" },
  { title: "下一步 ID", key: "nextId" },
  { title: "定时任务", key: "cron" },
  { title: "开始时间", key: "startTime" },
  { title: "状态", key: "running" },
  { title: "启用定时任务", key: "enable" },
  { title: "启用API", key: "apiEnable" },
  { title: "操作", key: "operate" },
];

// 单行数组（v-data-table 接受 items 数组）
const taskData = ref<TaskItem[]>();

/** 格式化开始时间（兼容 null / 空字符串） */
const formatStartTime = (v) => {
  if (!v) return "-";
  try {
    const d = new Date(v);
    if (Number.isNaN(d.getTime())) return v;
    return d.toLocaleString();
  } catch (e) {
    return v;
  }
};

onMounted(() => {
  initTask();
});

const initTask = () => {
  getTaskAll().then((e) => {
    taskData.value = e.data!;
  });
};

const edit = (item) => {
  editTask(item).then((e) => {
    if (e.code == 0) {
      snackbarStore.showSuccessMessage("success");
      initTask();
    }
  });
};

const changeEnable = (item) => {
  editTaskEnable({
    id: item.id,
    enable: item.enable,
  }).then((e) => {
    if (e.code == 0) {
      snackbarStore.showSuccessMessage("success");
      initTask();
    }
  });
};

const changeApi = () => {
  changeTaskKey(taskForm.value?.id).then((resp) => {
    if (resp.code == 0) {
      snackbarStore.showSuccessMessage("success");
      getTaskById(taskForm.value?.id).then((e) => {
        Object.assign(taskForm.value, e.data);
      });
    }
  });
};
</script>

<style scoped>
code {
  font-size: 0.85rem;
  background: rgba(0, 0, 0, 0.04);
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
