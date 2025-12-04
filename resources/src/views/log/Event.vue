<template>
  <v-container fluid class="py-6 px-8">
    <!-- 事件查看工具栏 -->
    <v-card class="mb-6 rounded-2xl elevation-3">
      <!-- 顶部标题和操作按钮 -->
      <div class="pa-4 d-flex align-center justify-space-between flex-wrap">
        <div class="d-flex align-center mb-2 mb-sm-0">
          <v-icon size="40" color="primary" class="mr-3">mdi-bell-ring</v-icon>
          <span class="text-h5 font-weight-bold text-primary">系统事件</span>
        </div>

        <div class="d-flex align-center ga-3 flex-wrap">
          <v-btn
            color="primary"
            variant="flat"
            class="rounded-lg px-4"
            @click="refreshEvents"
            :loading="loading"
          >
            <v-icon start>mdi-refresh</v-icon>
            刷新
          </v-btn>
        </div>
      </div>

      <v-divider></v-divider>

      <!-- 筛选条件 -->
      <v-expansion-panels flat>
        <v-expansion-panel>
          <v-expansion-panel-title>
            <v-icon start>mdi-filter</v-icon>
            筛选条件
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <v-container fluid>
              <v-row dense>
                <!-- 进程/任务名筛选 -->
                <v-col cols="12" sm="6" md="3">
                  <v-text-field
                    label="名称"
                    variant="outlined"
                    density="comfortable"
                    v-model="searchForm.name"
                    clearable
                    prepend-inner-icon="mdi-tag"
                  />
                </v-col>

                <!-- 事件类型筛选 -->
                <v-col cols="12" sm="6" md="3">
                  <v-select
                    label="事件类型"
                    variant="outlined"
                    density="comfortable"
                    v-model="searchForm.type"
                    :items="eventTypes"
                    item-title="label"
                    item-value="value"
                    clearable
                    prepend-inner-icon="mdi-shape"
                  />
                </v-col>

                <!-- 开始时间 -->
                <v-col cols="12" sm="6" md="3">
                  <v-text-field
                    label="开始时间"
                    variant="outlined"
                    density="comfortable"
                    type="datetime-local"
                    v-model="searchForm.startTime"
                    clearable
                    prepend-inner-icon="mdi-calendar-start"
                  />
                </v-col>

                <!-- 结束时间 -->
                <v-col cols="12" sm="6" md="3">
                  <v-text-field
                    label="结束时间"
                    variant="outlined"
                    density="comfortable"
                    type="datetime-local"
                    v-model="searchForm.endTime"
                    clearable
                    prepend-inner-icon="mdi-calendar-end"
                  />
                </v-col>

                <!-- 操作按钮 -->
                <v-col cols="12" class="d-flex align-center ga-2">
                  <v-btn color="primary" @click="searchEvents" :loading="loading">
                    <v-icon start>mdi-magnify</v-icon>
                    搜索
                  </v-btn>
                  <v-btn color="grey" variant="tonal" @click="resetSearch">
                    <v-icon start>mdi-refresh</v-icon>
                    重置
                  </v-btn>
                </v-col>
              </v-row>
            </v-container>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>
    </v-card>

    <!-- 事件列表 -->
    <v-card class="rounded-2xl elevation-2">
      <v-data-table
        :headers="headers"
        :items="eventData"
        :loading="loading"
        :items-per-page="pageSize"
        item-key="id"
        class="text-body-2"
        density="comfortable"
      >
        <!-- 事件类型 -->
        <template #item.type="{ item }">
          <v-chip
            :color="getEventTypeColor(item.type)"
            size="small"
            variant="tonal"
          >
            <v-icon start size="small">{{ getEventTypeIcon(item.type) }}</v-icon>
            {{ getEventTypeLabel(item.type) }}
          </v-chip>
        </template>

        <!-- 名称 -->
        <template #item.name="{ item }">
          <v-chip color="primary" size="small" variant="tonal">
            {{ item.name }}
          </v-chip>
        </template>

        <!-- 附加信息 -->
        <template #item.additional="{ item }">
          <div v-if="item.additional" class="additional-info">
            <template v-for="(value, key) in parseAdditional(item.additional)" :key="key">
              <v-chip size="x-small" variant="outlined" class="mr-1 mb-1">
                {{ key }}: {{ value }}
              </v-chip>
            </template>
          </div>
          <span v-else class="text-grey">-</span>
        </template>

        <!-- 时间 -->
        <template #item.createdTime="{ item }">
          <span class="text-caption">{{ formatTime(item.createdTime) }}</span>
        </template>

        <!-- 底部分页 -->
        <template #bottom>
          <div class="text-center pa-4">
            <v-pagination
              v-model="currentPage"
              :length="totalPages"
              :total-visible="7"
              @update:model-value="handlePageChange"
            ></v-pagination>
            <div class="mt-2 text-caption text-grey">
              共 {{ totalEvents }} 条事件，每页 {{ pageSize }} 条
            </div>
          </div>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { getEventList } from "~/src/api/event";
import type { Event, EventListReq, EventType } from "~/src/types/event/event";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const snackbarStore = useSnackbarStore();

// 事件类型选项
const eventTypes = [
  { label: "进程启动", value: "ProcessStart" },
  { label: "进程停止", value: "ProcessStop" },
  { label: "进程警告", value: "ProcessWarning" },
  { label: "任务启动", value: "TaskStart" },
  { label: "任务停止", value: "TaskStop" },
];

// 表头定义
const headers = [
  { title: "事件类型", key: "type", width: "150px" },
  { title: "名称", key: "name", width: "150px" },
  { title: "附加信息", key: "additional", sortable: false },
  { title: "时间", key: "createdTime", width: "180px" },
];

// 数据
const eventData = ref<Event[]>([]);
const totalEvents = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const loading = ref(false);

// 搜索表单
const searchForm = ref({
  name: "",
  type: "" as EventType | "",
  startTime: "",
  endTime: "",
});

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(totalEvents.value / pageSize.value);
});

// 获取事件类型颜色
const getEventTypeColor = (type: EventType) => {
  const colorMap: Record<EventType, string> = {
    ProcessStart: "success",
    ProcessStop: "error",
    ProcessWarning: "warning",
    TaskStart: "info",
    TaskStop: "secondary",
  };
  return colorMap[type] || "grey";
};

// 获取事件类型图标
const getEventTypeIcon = (type: EventType) => {
  const iconMap: Record<EventType, string> = {
    ProcessStart: "mdi-play-circle",
    ProcessStop: "mdi-stop-circle",
    ProcessWarning: "mdi-alert-circle",
    TaskStart: "mdi-clock-start",
    TaskStop: "mdi-clock-end",
  };
  return iconMap[type] || "mdi-information";
};

// 获取事件类型标签
const getEventTypeLabel = (type: EventType) => {
  const labelMap: Record<EventType, string> = {
    ProcessStart: "进程启动",
    ProcessStop: "进程停止",
    ProcessWarning: "进程警告",
    TaskStart: "任务启动",
    TaskStop: "任务停止",
  };
  return labelMap[type] || type;
};

// 解析附加信息
const parseAdditional = (additional: string): Record<string, string> => {
  try {
    return JSON.parse(additional);
  } catch {
    return {};
  }
};

// 格式化时间
const formatTime = (timestamp: string) => {
  if (!timestamp) return "-";
  const date = new Date(timestamp);
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

// 构建查询参数
const buildQuery = (): EventListReq => {
  const query: EventListReq = {
    page: currentPage.value,
    size: pageSize.value,
  };

  if (searchForm.value.name) {
    query.name = searchForm.value.name;
  }

  if (searchForm.value.type) {
    query.type = searchForm.value.type as EventType;
  }

  if (searchForm.value.startTime) {
    query.startTime = new Date(searchForm.value.startTime).getTime();
  }

  if (searchForm.value.endTime) {
    query.endTime = new Date(searchForm.value.endTime).getTime();
  }

  return query;
};

// 加载事件
const loadEvents = async () => {
  loading.value = true;
  try {
    const query = buildQuery();
    const response = await getEventList(query);

    if (response.code === 0 && response.data) {
      eventData.value = response.data.data || [];
      totalEvents.value = response.data.total || 0;
    } else {
      snackbarStore.showErrorMessage("加载事件失败");
    }
  } catch (error) {
    console.error("加载事件错误:", error);
    snackbarStore.showErrorMessage("加载事件出错");
  } finally {
    loading.value = false;
  }
};

// 搜索事件
const searchEvents = () => {
  currentPage.value = 1;
  loadEvents();
};

// 重置搜索
const resetSearch = () => {
  searchForm.value = {
    name: "",
    type: "",
    startTime: "",
    endTime: "",
  };
  currentPage.value = 1;
  loadEvents();
};

// 刷新事件
const refreshEvents = () => {
  loadEvents();
};

// 处理页码变化
const handlePageChange = (page: number) => {
  currentPage.value = page;
  loadEvents();
};

// 初始化
onMounted(() => {
  loadEvents();
});
</script>

<style scoped>
.additional-info {
  max-width: 400px;
}

:deep(.v-data-table__th) {
  font-weight: 600 !important;
  font-size: 0.875rem !important;
}

:deep(.v-data-table__td) {
  padding: 12px 16px !important;
}
</style>

