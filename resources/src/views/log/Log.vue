<template>
  <v-container fluid class="py-6 px-8">
    <!-- 日志查看工具栏 -->
    <v-card class="mb-6 rounded-2xl elevation-3">
      <!-- 顶部标题和操作按钮 -->
      <div class="pa-4 d-flex align-center justify-space-between flex-wrap">
        <div class="d-flex align-center mb-2 mb-sm-0">
          <v-icon size="40" color="primary" class="mr-3">mdi-text-box-search</v-icon>
          <span class="text-h5 font-weight-bold text-primary">日志查看</span>
        </div>

        <div class="d-flex align-center ga-3 flex-wrap">
          <v-btn
            color="success"
            variant="flat"
            class="rounded-lg px-4"
            @click="openTerminalView"
          >
            <v-icon start>mdi-console</v-icon>
            终端视图
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            class="rounded-lg px-4"
            @click="refreshLogs"
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
                <!-- 进程名筛选 -->
                <v-col cols="12" sm="6" md="4">
                  <v-autocomplete
                    label="进程名"
                    variant="outlined"
                    density="comfortable"
                    v-model="searchForm.name"
                    :items="processList"
                    clearable
                    prepend-inner-icon="mdi-application"
                  />
                </v-col>

                <!-- 日志内容搜索 -->
                <v-col cols="12" sm="6" md="4">
                  <v-text-field
                    label="日志内容"
                    variant="outlined"
                    density="comfortable"
                    v-model="searchForm.log"
                    clearable
                    prepend-inner-icon="mdi-text-search"
                  />
                </v-col>

                <!-- 使用类型 -->
                <v-col cols="12" sm="6" md="4">
                  <v-text-field
                    label="使用类型"
                    variant="outlined"
                    density="comfortable"
                    v-model="searchForm.using"
                    clearable
                    prepend-inner-icon="mdi-tag"
                  />
                </v-col>

                <!-- 开始时间 -->
                <v-col cols="12" sm="6" md="4">
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
                <v-col cols="12" sm="6" md="4">
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

                <!-- 排序选择 -->
                <v-col cols="12" sm="6" md="4">
                  <v-select
                    label="排序方式"
                    variant="outlined"
                    density="comfortable"
                    v-model="searchForm.sort"
                    :items="sortOptions"
                    item-title="label"
                    item-value="value"
                    prepend-inner-icon="mdi-sort"
                  />
                </v-col>

                <!-- 操作按钮 -->
                <v-col cols="12" sm="6" md="4" class="d-flex align-center ga-2">
                  <v-btn color="primary" @click="searchLogs" :loading="loading">
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

    <!-- 日志列表 -->
    <v-card class="rounded-2xl elevation-2">
      <v-data-table
        :headers="headers"
        :items="logData"
        :items-length="totalLogs"
        :loading="loading"
        v-model:page="currentPage"
        v-model:items-per-page="pageSize"
        item-key="id"
        class="text-body-2 log-table-seamless"
        density="compact"
      >
        <!-- 自定义列渲染 -->
        <template #item.log="{ item }">
          <div class="log-content" v-html="convertAnsiToHtml(item.log)"></div>
        </template>

        <template #item.time="{ item }">
          <span class="text-caption">{{ formatTime(item.time) }}</span>
        </template>

        <template #item.name="{ item }">
          <v-chip color="primary" size="x-small" variant="tonal">
            {{ item.name }}
          </v-chip>
        </template>

        <template #item.using="{ item }">
          <v-chip color="secondary" size="x-small" variant="tonal">
            {{ item.using || '-' }}
          </v-chip>
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
              共 {{ totalLogs }} 条日志，每页 {{ pageSize }} 条
            </div>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- 终端日志查看器 -->
    <LogTerminal ref="logTerminalRef" :search-form="searchForm" />
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { getLog } from "~/src/api/log";
import type { GetLogReq, ProcessLog } from "~/src/types/log/log";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { AnsiUp } from "ansi_up";
import LogTerminal from "~/src/components/log/LogTerminal.vue";
import { getProcessList } from "~/src/api/process";

const snackbarStore = useSnackbarStore();
const logTerminalRef = ref<InstanceType<typeof LogTerminal> | null>(null);
const ansiConverter = new AnsiUp();

// 进程列表
const processList = ref<string[]>([]);

// 排序选项
const sortOptions = [
  { label: "自动", value: "" },
  { label: "时间正序", value: "asc" },
  { label: "时间倒序", value: "desc" },
];

// 表头定义 - 日志内容放在最左边
const headers = [
  { title: "日志内容", key: "log", sortable: false },
  { title: "时间", key: "time", width: "150px" },
  { title: "进程名", key: "name", width: "30px" },
  { title: "类型", key: "using", width: "30px" },
];

// 数据
const logData = ref<ProcessLog[]>([]);
const totalLogs = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const loading = ref(false);

// 搜索表单
const searchForm = ref({
  name: "",
  log: "",
  using: "",
  startTime: "",
  endTime: "",
  sort: "",
});

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(totalLogs.value / pageSize.value);
});

// 转换 ANSI 颜色代码为 HTML
const convertAnsiToHtml = (text: string) => {
  if (!text) return "";
  return ansiConverter.ansi_to_html(text).replaceAll("color:rgb(255,255,255)", "color:rgb(160,160,160)");
};

// 格式化时间
const formatTime = (timestamp: number) => {
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
const buildQuery = (): GetLogReq => {
  const query: GetLogReq = {
    page: {
      from: (currentPage.value - 1) * pageSize.value,
      size: pageSize.value,
    },
  };

  // 添加排序
  if (searchForm.value.sort) {
    query.sort = searchForm.value.sort;
  } else {
    query.sort = "desc"; // 默认按时间倒序
  }

  // 添加匹配条件
  const match: any = {};
  if (searchForm.value.name) {
    match.name = searchForm.value.name;
  }
  if (searchForm.value.log) {
    match.log = searchForm.value.log;
  }
  if (searchForm.value.using) {
    match.using = searchForm.value.using;
  }
  if (Object.keys(match).length > 0) {
    query.match = match;
  }

  // 添加时间范围
  if (searchForm.value.startTime || searchForm.value.endTime) {
    query.time = {};
    if (searchForm.value.startTime) {
      query.time.startTime = new Date(searchForm.value.startTime).getTime();
    }
    if (searchForm.value.endTime) {
      query.time.endTime = new Date(searchForm.value.endTime).getTime();
    }
  }

  return query;
};

// 加载日志
const loadLogs = async () => {
  loading.value = true;
  try {
    const query = buildQuery();
    const response = await getLog(query);

    if (response.code === 0 && response.data) {
      logData.value = response.data.data || [];
      totalLogs.value = response.data.total || 0;
    } else {
      snackbarStore.showErrorMessage("加载日志失败");
    }
  } catch (error) {
    console.error("加载日志错误:", error);
    snackbarStore.showErrorMessage("加载日志出错");
  } finally {
    loading.value = false;
  }
};

// 搜索日志
const searchLogs = () => {
  currentPage.value = 1; // 重置到第一页
  loadLogs();
};

// 重置搜索
const resetSearch = () => {
  searchForm.value = {
    name: "",
    log: "",
    using: "",
    startTime: "",
    endTime: "",
    sort: "",
  };
  currentPage.value = 1;
  loadLogs();
};

// 刷新日志
const refreshLogs = () => {
  loadLogs();
};

// 处理页码变化
const handlePageChange = (page: number) => {
  currentPage.value = page;
  loadLogs();
};

// 打开终端视图
const openTerminalView = () => {
  if (logTerminalRef.value) {
    logTerminalRef.value.open();
  }
};

// 加载进程列表
const loadProcessList = async () => {
  try {
    const response = await getProcessList();
    if (response.code === 0 && response.data) {
      // 提取进程名，去重
      processList.value = Array.from(
        new Set(response.data.map((item) => item.name))
      ).sort();
    }
  } catch (error) {
    console.error("加载进程列表错误:", error);
  }
};

// 初始化
onMounted(() => {
  loadProcessList();
  loadLogs();
});
</script>

<style scoped>
.log-content {
  font-size: 0.8rem;
  padding: 3px 10px;
  border-radius: 0;
  font-family: "Consolas", "Monaco", "Courier New", monospace;
  overflow-x: auto;
  max-width: 100%;
  line-height: 1.2;
  display: block;
  margin: -1px 0;
}

/* ANSI 颜色样式 */
:deep(.log-content span) {
  white-space: pre-wrap;
  word-break: break-word;
}

:deep(.v-data-table__th) {
  font-weight: 600 !important;
  font-size: 0.8rem !important;
  text-align: center !important;
}

/* 日志内容列无间距 */
:deep(.log-table-seamless .v-data-table__td:first-child) {
  padding: 0 !important;
  height: auto !important;
  line-height: 1 !important;
}

/* 其他列保持小间距 */
:deep(.log-table-seamless .v-data-table__td:not(:first-child)) {
  padding: 4px 8px !important;
  font-size: 0.75rem;
  vertical-align: middle;
}

/* 移除表格行间距和边框，设置最小高度 */
:deep(.log-table-seamless tbody tr) {
  border: none !important;
  height: 0 !important;
}

:deep(.log-table-seamless .v-data-table__tr) {
  border-bottom: none !important;
  height: 0 !important;
}

:deep(.log-table-seamless tbody tr td) {
  border: none !important;
}

/* 移除hover背景色，避免破坏连续效果 */
:deep(.log-table-seamless tbody tr:hover) {
  background-color: transparent !important;
}

/* 移除表格的内边距和间距 */
:deep(.log-table-seamless .v-table__wrapper) {
  border-spacing: 0 !important;
}

:deep(.log-table-seamless table) {
  border-collapse: collapse !important;
  border-spacing: 0 !important;
}
</style>

