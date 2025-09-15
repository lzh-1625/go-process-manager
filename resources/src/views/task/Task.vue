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
        <v-icon class="mr-2" @click="editTaskBefore(item)"> mdi-pencil </v-icon>
        <v-icon> mdi-delete </v-icon>
      </template>
    </v-data-table>
  </v-card>

  <v-dialog v-model="taskDialog" max-width="600px" persistent>
    <v-card class="rounded-xl">
      <!-- 标题 -->
      <v-card-title class="text-h6 font-weight-medium">
        {{ isAdd ? "添加任务" : "修改任务" }}
      </v-card-title>

      <v-divider></v-divider>

      <!-- 表单内容 -->
      <v-card-text class="pt-6">
        <v-container fluid>
          <v-row dense>
            <v-col cols="12" sm="6">
              <v-autocomplete
                label="判断条件"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="conditionSelect"
                v-model="taskForm.condition"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                :disabled="taskForm.condition == 3"
                label="判断目标"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="processSelect"
                v-model="taskForm.processId"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                label="操作目标"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="processSelect"
                v-model="taskForm.operationTarget"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                label="执行操作"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="operationSelect"
                v-model="taskForm.operation"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                label="触发目标"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="processSelect"
                v-model="taskForm.triggerTarget"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                :disabled="taskForm.triggerTarget == null"
                label="触发事件"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="eventSelect"
                v-model="taskForm.triggerEvent"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                label="后续任务"
                item-title="name"
                item-value="value"
                variant="outlined"
                density="comfortable"
                :items="taskSelect"
                v-model="taskForm.nextId"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-text-field
                label="定时任务"
                variant="outlined"
                density="comfortable"
                v-model="taskForm.cron"
              />
            </v-col>

            <v-col cols="12">
              <v-text-field
                label="API"
                variant="outlined"
                density="comfortable"
                readonly
                v-model="apiUrl"
                append-inner-icon="mdi-content-copy"
                @click:append-inner="copyToClipboard"
              />
            </v-col>
          </v-row>

          <div class="d-flex justify-end mt-3">
            <v-btn @click="changeApi" color="primary" variant="tonal">
              {{ taskForm?.key != null ? "刷新 API" : "创建 API" }}
              <v-icon end>
                {{ taskForm?.key != null ? "mdi-refresh" : "mdi-plus" }}
              </v-icon>
            </v-btn>
          </div>
        </v-container>
      </v-card-text>

      <v-divider></v-divider>

      <!-- 底部操作按钮 -->
      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="taskDialog = false">取消</v-btn>
        <v-btn color="primary" @click="submit">确认</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ref, onMounted } from "vue";
import {
  addTask,
  changeTaskKey,
  editTask,
  editTaskEnable,
  getTaskAll,
  getTaskById,
} from "~/src/api/task";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { TaskItem } from "~/src/types/tassk/task";

const snackbarStore = useSnackbarStore();

// 弹窗 & 表单
const taskDialog = ref(false);
const taskForm = ref<Partial<TaskItem>>({});
const isAdd = ref(false);

// 下拉框选项
const taskSelect = ref<any[]>([]);
const processSelect = ref<any[]>([]);
const eventSelect = ref<any[]>([]);
const operationSelect = ref<any[]>([]);
const conditionSelect = ref<any[]>([]);

// 映射表
const conditionMap = {
  0: "运行中",
  1: "已停止",
  2: "错误",
  3: "无条件",
};
const operationMap = {
  0: "异步启动",
  1: "异步停止",
  2: "完成启动",
  3: "完成停止",
};
const eventMap = {
  0: "停止",
  1: "启动",
  2: "异常",
};

const urlBase = ref(`${window.location.origin}/api/task/api-key/`);

// 表头
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
const apiUrl = computed(() =>
  taskForm.value?.key ? urlBase.value + taskForm.value.key : "未创建api"
);
// 任务数据
const taskData = ref<TaskItem[]>([]);

// 格式化时间
const formatStartTime = (v: Date) => {
  if (!v) return "-";
  try {
    if (Number.isNaN(v.getTime())) return v;
    return v.toLocaleString();
  } catch {
    return v;
  }
};

onMounted(() => {
  initTask();
});

// 打开添加任务弹窗
const addTaskBefore = () => {
  isAdd.value = true;
  taskForm.value = {}; // 清空表单
  taskDialog.value = true;
};

// 打开编辑任务弹窗
const editTaskBefore = (row: TaskItem) => {
  isAdd.value = false;
  getTaskById(row.id).then((res) => {
    taskForm.value = res.data ?? {};
  });
  taskDialog.value = true;
};

// 复制 API 地址
const copyToClipboard = () => {
  if (!taskForm.value?.key) return;
  navigator.clipboard
    .writeText(urlBase.value + taskForm.value.key)
    .then(() => {
      snackbarStore.showSuccessMessage("复制成功");
    })
    .catch((err) => {
      console.error("复制失败:", err);
    });
};

// 初始化任务 & 下拉框选项
const initTask = () => {
  getTaskAll().then((res) => {
    const list = res.data ?? [];
    taskData.value = list;

    // 任务选择
    taskSelect.value = list.map((t) => ({
      name: t.id,
      value: t.id,
    }));
    taskSelect.value.push({ name: "无", value: null });

    // 进程选择
    processSelect.value = list.map((t) => ({
      name: t.processName,
      value: t.processId,
    }));

    // 操作/事件/条件
    operationSelect.value = Object.entries(operationMap).map(
      ([value, name]) => ({
        name,
        value: parseInt(value),
      })
    );

    eventSelect.value = Object.entries(eventMap).map(([value, name]) => ({
      name,
      value: parseInt(value),
    }));

    conditionSelect.value = Object.entries(conditionMap).map(
      ([value, name]) => ({
        name,
        value: parseInt(value),
      })
    );
  });
};

// 修改任务
const edit = (item: TaskItem) => {
  editTask(item).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("修改成功");
      initTask();
    }
  });
};

// 切换启用状态
const changeEnable = (item: TaskItem) => {
  editTaskEnable({ id: item.id, enable: item.enable }).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("修改成功");
      initTask();
    }
  });
};

// 生成/刷新 API Key
const changeApi = () => {
  if (!taskForm.value?.id) return;
  changeTaskKey(taskForm.value.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("API 更新成功");
      getTaskById(taskForm.value?.id).then((e) => {
        Object.assign(taskForm.value, e.data);
      });
    }
  });
};

// 提交
const submit = () => {
  if (isAdd.value) {
    addTask(taskForm.value).then((res) => {
      if (res.code === 0) {
        taskDialog.value = false;
        snackbarStore.showSuccessMessage("添加成功");
        initTask();
      }
    });
  } else {
    editTask(taskForm.value).then((res) => {
      if (res.code === 0) {
        taskDialog.value = false;
        snackbarStore.showSuccessMessage("修改成功");
        initTask();
      }
    });
  }
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
