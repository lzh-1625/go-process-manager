<template>
  <v-container fluid class="py-6 px-8">
    <v-card class="rounded-lg">
      <!-- loading spinner -->
      <div
        v-if="loading && logData.length === 0"
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
          <v-icon color="primary" class="mr-2">mdi-text-box-search</v-icon>
          <span class="flex-fill">日志查看</span>
          <v-btn
            icon
            variant="text"
            size="small"
            @click="openTerminalView"
            class="mr-1"
          >
            <v-icon>mdi-console</v-icon>
          </v-btn>
          <v-btn icon variant="text" size="small" @click="refreshLogs">
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
          <v-btn
            icon
            variant="text"
            size="small"
            @click="showFilter = !showFilter"
          >
            <v-icon>mdi-filter</v-icon>
          </v-btn>
        </h6>

        <!-- 筛选条件 -->
        <v-expand-transition>
          <div v-show="showFilter" class="px-5 pb-4">
            <v-row dense>
              <!-- 进程名筛选 -->
              <v-col cols="12" sm="6" md="3">
                <v-autocomplete
                  label="进程名"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.name"
                  :items="processList"
                  multiple
                  chips
                  clearable
                  hide-details
                />
              </v-col>

              <!-- 日志内容搜索 -->
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  label="日志内容"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.log"
                  clearable
                  hide-details
                />
              </v-col>

              <!-- 使用类型 -->
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  label="使用者"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.using"
                  clearable
                  hide-details
                />
              </v-col>

              <!-- 排序选择 -->
              <v-col cols="12" sm="6" md="3">
                <v-select
                  label="排序方式"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.sort"
                  :items="sortOptions"
                  item-title="label"
                  item-value="value"
                  hide-details
                />
              </v-col>

              <!-- 开始时间 -->
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  label="开始时间"
                  variant="outlined"
                  density="compact"
                  type="datetime-local"
                  v-model="searchForm.startTime"
                  clearable
                  hide-details
                />
              </v-col>

              <!-- 结束时间 -->
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  label="结束时间"
                  variant="outlined"
                  density="compact"
                  type="datetime-local"
                  v-model="searchForm.endTime"
                  clearable
                  hide-details
                />
              </v-col>

              <!-- 操作按钮 -->
              <v-col cols="12" sm="6" md="3" class="d-flex ga-2">
                <v-btn
                  color="primary"
                  size="small"
                  elevation="4"
                  variant="elevated"
                  @click="searchLogs"
                >
                  搜索
                </v-btn>
                <v-btn size="small" variant="tonal" @click="resetSearch">
                  重置
                </v-btn>
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>

        <!-- 日志列表 -->
        <v-table class="pa-3 log-table">
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
            <tr v-for="item in logData" :key="item.id">
              <td class="log-cell">
                <div
                  class="log-content"
                  v-html="convertAnsiToHtml(item.log)"
                ></div>
              </td>
              <td>
                <v-btn
                  elevation="4"
                  variant="elevated"
                  size="small"
                  @click="viewLogContext(item)"
                >
                  <v-icon>mdi-arrow-up-down</v-icon>
                </v-btn>
              </td>
              <td>
                <span class="text-caption">{{ formatTime(item.time) }}</span>
              </td>
              <td>
                <v-chip color="primary" size="small" class="font-weight-bold">
                  {{ item.name }}
                </v-chip>
              </td>
              <td>
                <v-chip color="secondary" size="small" class="font-weight-bold">
                  {{ item.using || "-" }}
                </v-chip>
              </td>
            </tr>
            <tr v-if="logData.length === 0">
              <td colspan="5" class="text-center text-secondary pa-8">
                暂无数据
              </td>
            </tr>
          </tbody>
        </v-table>

        <!-- 分页 -->
        <div class="text-center pa-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages > 500 ? 500 : totalPages"
            :total-visible="7"
            density="compact"
            @update:model-value="handlePageChange"
          ></v-pagination>
          <div class="mt-2 text-caption text-secondary">
            共 {{ totalLogs }} 条日志
          </div>
        </div>
      </div>
    </v-card>

    <!-- 终端日志查看器 -->
    <LogTerminal
      ref="logTerminalRef"
      :search-form="searchForm"
      :process-list="processList"
    />
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
  { title: "上下文", key: "actions", width: "100px", sortable: false },
  { title: "时间", key: "time", width: "150px" },
  { title: "进程名", key: "name", width: "30px" },
  { title: "使用者", key: "using", width: "30px" },
];

// 数据
const logData = ref<ProcessLog[]>([]);
const totalLogs = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const showFilter = ref(false);

// 搜索表单
const searchForm = ref({
  name: [] as string[],
  log: "",
  using: "",
  startTime: "",
  endTime: "",
  sort: "desc",
});

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(totalLogs.value / pageSize.value);
});

// 转换 ANSI 颜色代码为 HTML
const convertAnsiToHtml = (text: string) => {
  if (!text) return "";
  return ansiConverter
    .ansi_to_html(text)
    .replaceAll("color:rgb(255,255,255)", "color:rgb(160,160,160)");
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
  }
  // 添加匹配条件
  const match: any = {};
  if (searchForm.value.name && searchForm.value.name.length > 0) {
    // 支持多选，将进程名数组传递给后端
    query.filterName = searchForm.value.name;
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
    snackbarStore.showWarningMessage("未获取到日志");
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
    name: [],
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

// 查看日志上下文
const viewLogContext = (log: ProcessLog) => {
  if (logTerminalRef.value) {
    logTerminalRef.value.openWithContext(log.time, log.name);
  }
};

// 加载进程列表
const loadProcessList = async () => {
  try {
    const response = await getProcessList();
    if (response.code === 0 && response.data) {
      // 提取进程名，去重
      processList.value = Array.from(
        new Set(response.data.map((item) => item.name)),
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

<style lang="scss" scoped>
.log-content {
  font-size: 0.8rem;
  padding: 3px 10px;
  border-radius: 0;
  font-family: "Consolas", "Monaco", "Courier New", monospace;
  max-width: 100%;
  line-height: 1.2;
  display: block;
  margin: -1px 0;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: break-word;
}

/* ANSI 颜色样式 */
:deep(.log-content span) {
  white-space: pre-wrap;
  word-break: break-word;
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
        transition:
          box-shadow 0.2s,
          transform 0.2s;

        &:not(.v-data-table__selected):hover {
          box-shadow: 0 3px 15px -2px rgba(0, 0, 0, 0.12);
          transform: translateY(-4px);
        }
      }
    }
  }
}

/* 日志内容列样式 */
.log-cell {
  padding: 0 !important;
  height: auto !important;
  line-height: 1 !important;
}
</style>
