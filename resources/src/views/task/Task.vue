<template>
  <v-container fluid class="py-6 px-8">
    <v-card class="rounded-lg">
      <!-- loading spinner -->
      <div
        v-if="loading"
        class="h-full d-flex flex-grow-1 align-center justify-center"
        style="min-height: 400px"
      >
        <v-progress-circular
          indeterminate
          color="primary"
        ></v-progress-circular>
      </div>

      <div v-else>
        <!-- 标题栏 -->
        <h6 class="text-h6 font-weight-bold pa-5 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-format-list-checks</v-icon>
          <span class="flex-fill">任务</span>
          <v-btn icon variant="text" size="small" @click="refreshTasks">
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
          <v-btn
            color="primary"
            variant="tonal"
            size="small"
            @click="addTaskBefore"
          >
            <v-icon left>mdi-plus</v-icon>
            新建任务
          </v-btn>
        </h6>

        <!-- 任务列表 -->
        <v-table class="pa-3">
          <thead>
            <tr>
              <th
                class="text-left"
                v-for="header in headers"
                :key="header.title"
              >
                {{ header.title }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in paginatedTasks" :key="item.id">
              <td class="font-weight-bold">
                <v-chip color="primary" size="small" class="font-weight-bold">
                  {{ item.id }}
                </v-chip>
              </td>
              <td class="font-weight-bold">{{ item.name }}</td>
              <td>{{ item.nextId === null ? "-" : item.nextId }}</td>
              <td>
                <code v-if="item.cron">{{ item.cron }}</code>
                <span v-else class="text-secondary">-</span>
              </td>
              <td>{{ formatStartTime(item.startTime) }}</td>
              <td>
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
                <svg
                  v-else
                  width="20"
                  height="20"
                  viewBox="0 0 48 48"
                  fill="#000000"
                >
                  <path
                    d="M42.02 12.71l-1.38-1.42a1 1 0 00-1.41 0L18.01 32.5l-9.9-9.89a1 1 0 00-1.41 0l-1.41 1.41a1 1 0 000 1.41l10.6 10.61 1.42 1.41a1 1 0 001.41 0l1.41-1.41 21.92-21.92a1 1 0 00-.03-1.41z"
                    fill="#000000"
                  />
                </svg>
              </td>
              <td>
                <v-switch
                  color="primary"
                  @change="edit(item)"
                  v-model="item.enable"
                  density="compact"
                  inset
                  hide-details
                ></v-switch>
              </td>
              <td>
                <v-switch
                  color="primary"
                  @change="edit(item)"
                  v-model="item.apiEnable"
                  density="compact"
                  inset
                  hide-details
                ></v-switch>
              </td>
              <td>
                <v-icon
                  class="mr-2"
                  v-if="!item.running"
                  @click="startTask(item)"
                  size="small"
                >
                  mdi-play
                </v-icon>
                <v-icon
                  class="mr-2"
                  v-else
                  @click="stopTask(item)"
                  size="small"
                >
                  mdi-stop
                </v-icon>
                <v-icon class="mr-2" @click="editTaskBefore(item)" size="small">
                  mdi-pencil
                </v-icon>
                <v-icon @click="deleteTask(item)" size="small">
                  mdi-delete
                </v-icon>
              </td>
            </tr>
            <tr v-if="taskData.length === 0">
              <td colspan="9" class="text-center text-secondary pa-8">
                暂无数据
              </td>
            </tr>
          </tbody>
        </v-table>

        <!-- 分页 -->
        <div class="text-center pa-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            :total-visible="7"
            density="compact"
            @update:model-value="handlePageChange"
          ></v-pagination>
          <div class="mt-2 text-caption text-secondary">
            共 {{ taskData.length }} 个任务
          </div>
        </div>
      </div>
    </v-card>
  </v-container>

  <v-dialog v-model="taskDialog" max-width="600px">
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
            <v-col cols="12" sm="12">
              <v-text-field
                label="任务名"
                variant="outlined"
                density="comfortable"
                v-model="taskForm.name"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
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
                :disabled="taskForm.processId == null"
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
                :disabled="taskForm.operationTarget == null"
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
                placeholder="* * * * *"
                v-model="taskForm.cron"
              />
            </v-col>

            <v-col cols="12" v-if="!isAdd">
              <v-text-field
                label="API"
                variant="outlined"
                density="comfortable"
                readonly
                v-model="apiUrl"
                append-inner-icon="mdi-content-copy"
                @click:append-inner="copyToClipboard"
              >
                <!-- 把按钮放进输入框右侧 -->
                <template v-slot:append>
                  <v-btn
                    v-if="taskForm?.key == null"
                    @click="changeApi"
                    color="primary"
                    variant="tonal"
                    size="small"
                    icon="mdi-plus"
                  >
                  </v-btn>
                  <v-btn
                    v-else
                    @click="changeApi"
                    color="primary"
                    variant="tonal"
                    size="small"
                    icon="mdi-refresh"
                  >
                  </v-btn>
                </template>
              </v-text-field>
            </v-col>
          </v-row>
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
import { getProcessList } from "~/src/api/process";
import {
  addTask,
  changeTaskKey,
  editTask,
  getTaskAll,
  getTaskById,
  deleteTaskById,
  startTaskById,
  stopTaskById,
} from "~/src/api/task";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import axios from "axios";
import { onBeforeUnmount } from "vue";
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

const version = ref(0);
// 表头
const headers = [
  { title: "任务ID", key: "id" },
  { title: "任务名", key: "name" },
  { title: "下一步 ID", key: "nextId" },
  { title: "定时任务", key: "cron" },
  { title: "开始时间", key: "startTime" },
  { title: "状态", key: "running" },
  { title: "启用定时任务", key: "enable" },
  { title: "启用API", key: "apiEnable" },
  {
    title: "操作",
    key: "operate",
    headerProps: { style: "min-width: 150px;" },
  },
];
const apiUrl = computed(() =>
  taskForm.value?.key ? urlBase.value + taskForm.value.key : "未创建api"
);

// 任务数据
const taskData = ref<TaskItem[]>([]);
const loading = ref(false);

// 分页
const currentPage = ref(1);
const pageSize = ref(10);

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(taskData.value.length / pageSize.value);
});

// 计算当前页数据
const paginatedTasks = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return taskData.value.slice(start, end);
});

// 处理页码变化
const handlePageChange = (page: number) => {
  currentPage.value = page;
};

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

// 刷新任务列表
const refreshTasks = () => {
  initTask();
};

onMounted(() => {
  initTask();

  getTaskListWait();
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
    if (taskForm.value.processId === 0) {
      taskForm.value.processId = undefined;
    }
    if (taskForm.value.operationTarget === 0) {
      taskForm.value.operationTarget = undefined;
    }
  });
  taskDialog.value = true;
};

// 打开编辑任务弹窗
const deleteTask = (row: TaskItem) => {
  deleteTaskById(row.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("操作成功");
    }
  });
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
let cancelTokenSource: any;

onBeforeUnmount(() => {
  if (cancelTokenSource) {
    cancelTokenSource.cancel("组件已销毁，取消请求");
  }
});

const getTaskListWait = () => {
  cancelTokenSource = axios.CancelToken.source();
  axios
    .get("api/task/all/wait", {
      cancelToken: cancelTokenSource.token,
      headers: {
        Authorization: "bearer " + localStorage.getItem("token"),
        Version: version.value,
      },
    })
    .then((response) => {
      version.value = parseInt(response.headers?.version || "0");
      taskData.value = response.data.data;
      getTaskListWait();
    })
    .catch((error) => {
      console.error("请求错误:", error);
    });
};

// 初始化任务 & 下拉框选项
const initTask = () => {
  loading.value = true;
  getTaskAll()
    .then((res) => {
      const list = res.data ?? [];
      taskData.value = list;

      // 任务选择
      taskSelect.value = list.map((t) => ({
        name: t.id,
        value: t.id,
      }));
      taskSelect.value.push({ name: "无", value: null });
    })
    .finally(() => {
      loading.value = false;
    });
  // 操作/事件/条件
  operationSelect.value = Object.entries(operationMap).map(([value, name]) => ({
    name,
    value: parseInt(value),
  }));

  eventSelect.value = Object.entries(eventMap).map(([value, name]) => ({
    name,
    value: parseInt(value),
  }));

  conditionSelect.value = Object.entries(conditionMap).map(([value, name]) => ({
    name,
    value: parseInt(value),
  }));
  getProcessList().then((resp) => {
    if (resp.code == 0) {
      processSelect.value = resp.data!.map((e) => {
        return {
          name: e.name,
          value: e.uuid,
        };
      });
      processSelect.value.push({
        name: "无",
        value: null,
      });
    }
  });
};

// 修改任务
const edit = (item: TaskItem) => {
  editTask(item).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("修改成功");
    }
  });
};


const startTask = (item: TaskItem) => {
  startTaskById(item.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("修改成功");
    }
  });
};
const stopTask = (item: TaskItem) => {
  stopTaskById(item.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("修改成功");
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
      }
    });
  } else {
    editTask(taskForm.value).then((res) => {
      if (res.code === 0) {
        taskDialog.value = false;
        snackbarStore.showSuccessMessage("修改成功");
      }
    });
  }
};
</script>

<style lang="scss" scoped>
code {
  font-size: 0.85rem;
  background: rgba(0, 0, 0, 0.04);
  padding: 2px 6px;
  border-radius: 4px;
}

.v-table {
  table {
    padding: 4px;
    padding-bottom: 8px;

    th {
      text-transform: uppercase;
      white-space: nowrap;
    }

    td {
      border-bottom: 0 !important;
    }

    tbody {
      tr {
        transition: box-shadow 0.2s, transform 0.2s;

        &:not(.v-data-table__selected):hover {
          box-shadow: 0 3px 15px -2px rgba(0, 0, 0, 0.12);
          transform: translateY(-4px);
        }
      }
    }
  }
}
</style>
