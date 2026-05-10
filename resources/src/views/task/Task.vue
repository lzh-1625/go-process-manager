<template>
  <v-container fluid class="py-6 px-8">
    <v-card class="rounded-lg">
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
        <h6 class="text-h6 font-weight-bold pa-5 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-format-list-checks</v-icon>
          <span class="flex-fill">{{ $t('taskPage.title') }}</span>
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
            {{ $t('taskPage.newTask') }}
          </v-btn>
        </h6>

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
                {{ $t('common.noData') }}
              </td>
            </tr>
          </tbody>
        </v-table>

        <div class="text-center pa-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            :total-visible="7"
            density="compact"
            @update:model-value="handlePageChange"
          ></v-pagination>
          <div class="mt-2 text-caption text-secondary">
            {{ $t('taskPage.totalTasks', { n: taskData.length }) }}
          </div>
        </div>
      </div>
    </v-card>
  </v-container>

  <v-dialog v-model="taskDialog" max-width="600px">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">
        {{ isAdd ? $t('taskPage.addTask') : $t('taskPage.editTask') }}
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-container fluid>
          <v-row dense>
            <v-col cols="12" sm="12">
              <v-text-field
                :label="$t('taskPage.taskName')"
                variant="outlined"
                density="comfortable"
                v-model="taskForm.name"
              />
            </v-col>

            <v-col cols="12" sm="6">
              <v-autocomplete
                :label="$t('taskPage.judgeTarget')"
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
                :label="$t('taskPage.judgeCondition')"
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
                :label="$t('taskPage.operationTarget')"
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
                :label="$t('taskPage.executeOperation')"
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
                :label="$t('taskPage.triggerTarget')"
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
                :label="$t('taskPage.triggerEvent')"
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
                :label="$t('taskPage.nextTask')"
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
                :label="$t('taskPage.cronJob')"
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

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="taskDialog = false">{{ $t('common.cancel') }}</v-btn>
        <v-btn color="primary" @click="submit">{{ $t('common.confirm') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { getProcessList } from "~/src/api/process";
import {
  addTask,
  changeTaskKey,
  editTask,
  getTaskAll,
  GetTaskByID,
  deleteTaskById,
  startTaskById,
  stopTaskById,
} from "~/src/api/task";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import axios from "axios";
import { onBeforeUnmount } from "vue";
import { TaskItem } from "~/src/types/tassk/task";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();

const taskDialog = ref(false);
const taskForm = ref<Partial<TaskItem>>({});
const isAdd = ref(false);

const taskSelect = ref<any[]>([]);
const processSelect = ref<any[]>([]);
const eventSelect = ref<any[]>([]);
const operationSelect = ref<any[]>([]);
const conditionSelect = ref<any[]>([]);

const conditionMap = {
  0: "taskPage.conditionRunning",
  1: "taskPage.conditionStopped",
  2: "taskPage.conditionError",
  3: "taskPage.conditionUnconditional",
};
const operationMap = {
  0: "taskPage.opAsyncStart",
  1: "taskPage.opAsyncStop",
  2: "taskPage.opCompleteStart",
  3: "taskPage.opCompleteStop",
};
const eventMap = {
  0: "taskPage.eventStop",
  1: "taskPage.eventStart",
  2: "taskPage.eventException",
};

const urlBase = ref(`${window.location.origin}/api/task/api-key/`);

const version = ref(0);
const headers = computed(() => [
  { title: t("taskPage.taskId"), key: "id" },
  { title: t("taskPage.taskName"), key: "name" },
  { title: t("taskPage.nextStepId"), key: "nextId" },
  { title: t("taskPage.cronJob"), key: "cron" },
  { title: t("taskPage.startTime"), key: "startTime" },
  { title: t("taskPage.status"), key: "running" },
  { title: t("taskPage.enableCron"), key: "enable" },
  { title: t("taskPage.enableApi"), key: "apiEnable" },
  {
    title: t("common.operation"),
    key: "operate",
    headerProps: { style: "min-width: 150px;" },
  },
]);
const apiUrl = computed(() =>
  taskForm.value?.key ? urlBase.value + taskForm.value.key : t("taskPage.noApi")
);

const taskData = ref<TaskItem[]>([]);
const loading = ref(false);

const currentPage = ref(1);
const pageSize = ref(10);

const totalPages = computed(() => {
  return Math.ceil(taskData.value.length / pageSize.value);
});

const paginatedTasks = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return taskData.value.slice(start, end);
});

const handlePageChange = (page: number) => {
  currentPage.value = page;
};

const formatStartTime = (v: Date) => {
  if (!v) return "-";
  try {
    if (Number.isNaN(v.getTime())) return v;
    return v.toLocaleString();
  } catch {
    return v;
  }
};

const refreshTasks = () => {
  initTask();
};

onMounted(() => {
  initTask();
  getTaskListWait();
});

const addTaskBefore = () => {
  isAdd.value = true;
  taskForm.value = {};
  taskDialog.value = true;
};

const editTaskBefore = (row: TaskItem) => {
  isAdd.value = false;
  GetTaskByID(row.id).then((res) => {
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

const deleteTask = (row: TaskItem) => {
  deleteTaskById(row.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(t("common.operationSuccess"));
    }
  });
};

const copyToClipboard = () => {
  if (!taskForm.value?.key) return;
  navigator.clipboard
    .writeText(urlBase.value + taskForm.value.key)
    .then(() => {
      snackbarStore.showSuccessMessage(t("common.copySuccess"));
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

const initTask = () => {
  loading.value = true;
  getTaskAll()
    .then((res) => {
      const list = res.data ?? [];
      taskData.value = list;

      taskSelect.value = list.map((t) => ({
        name: t.id,
        value: t.id,
      }));
      taskSelect.value.push({ name: t("common.none"), value: null });
    })
    .finally(() => {
      loading.value = false;
    });

  operationSelect.value = Object.entries(operationMap).map(([value, name]) => ({
    name: t(name as string),
    value: parseInt(value),
  }));

  eventSelect.value = Object.entries(eventMap).map(([value, name]) => ({
    name: t(name as string),
    value: parseInt(value),
  }));

  conditionSelect.value = Object.entries(conditionMap).map(([value, name]) => ({
    name: t(name as string),
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
        name: t("common.none"),
        value: null,
      });
    }
  });
};

const edit = (item: TaskItem) => {
  editTask(item).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(t("common.editSuccess"));
    }
  });
};

const startTask = (item: TaskItem) => {
  startTaskById(item.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(t("common.operationSuccess"));
    }
  });
};
const stopTask = (item: TaskItem) => {
  stopTaskById(item.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(t("common.operationSuccess"));
    }
  });
};

const changeApi = () => {
  if (!taskForm.value?.id) return;
  changeTaskKey(taskForm.value.id).then((res) => {
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(t("taskPage.apiUpdateSuccess"));
      GetTaskByID(taskForm.value?.id).then((e) => {
        Object.assign(taskForm.value, e.data);
      });
    }
  });
};

const submit = () => {
  if (isAdd.value) {
    addTask(taskForm.value).then((res) => {
      if (res.code === 0) {
        taskDialog.value = false;
        snackbarStore.showSuccessMessage(t("taskPage.addSuccess"));
      }
    });
  } else {
    editTask(taskForm.value).then((res) => {
      if (res.code === 0) {
        taskDialog.value = false;
        snackbarStore.showSuccessMessage(t("common.editSuccess"));
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
