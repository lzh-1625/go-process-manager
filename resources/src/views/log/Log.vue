<template>
  <v-container fluid class="log-page py-4 py-sm-6 px-3 px-sm-6 px-md-8">
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
        <h6 class="log-page__title text-h6 font-weight-bold pa-4 pa-sm-5 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-text-box-search</v-icon>
          <span class="flex-fill">{{ $t("logPage.title") }}</span>
          <v-btn
            icon
            variant="text"
            size="small"
            :loading="loading"
            :disabled="loading"
            @click="refreshLogs"
          >
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
          <v-btn
            icon
            variant="text"
            size="small"
            :color="showOnlyContent ? 'primary' : undefined"
            @click="showOnlyContent = !showOnlyContent"
          >
            <v-icon>{{ showOnlyContent ? "mdi-text-box" : "mdi-text-box-outline" }}</v-icon>
          </v-btn>
        </h6>

        <!-- 筛选条件 -->
        <v-expand-transition>
          <div v-show="showFilter" class="px-4 px-sm-5 pb-4">
            <v-row dense>
              <!-- 进程名筛选 -->
              <v-col cols="12" sm="6" md="3">
                <v-autocomplete
                  :label="$t('logPage.processName')"
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
                  :label="$t('logPage.logContent')"
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
                  :label="$t('logPage.user')"
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
                  :label="$t('logPage.sortBy')"
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
                  :label="$t('common.startTime')"
                  variant="outlined"
                  density="compact"
                  type="datetime-local"
                  step="1"
                  v-model="searchForm.startTime"
                  clearable
                  hide-details
                />
              </v-col>

              <!-- 结束时间 -->
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  :label="$t('common.endTime')"
                  variant="outlined"
                  density="compact"
                  type="datetime-local"
                  step="1"
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
                  :loading="loading"
                  :disabled="loading"
                  @click="searchLogs"
                >
                  {{ $t("common.search") }}
                </v-btn>
                <v-btn
                  size="small"
                  variant="tonal"
                  :disabled="loading"
                  @click="resetSearch"
                >
                  {{ $t("common.reset") }}
                </v-btn>
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>

        <!-- 日志列表 -->
        <div
          class="log-stream px-2 px-sm-4 pb-2 pb-sm-4"
          :class="{ 'log-stream--content-only': showOnlyContent }"
        >
          <v-overlay
            :model-value="loading && logData.length > 0"
            contained
            class="align-center justify-center"
            persistent
          >
            <v-progress-circular
              indeterminate
              color="primary"
            ></v-progress-circular>
          </v-overlay>

          <div class="log-stream__header" v-if="!smAndDown && !showOnlyContent">
            <span>{{ $t("logPage.logContent") }}</span>
            <span>{{ $t("common.time") }} / {{ $t("logPage.processName") }} / {{ $t("logPage.user") }}</span>
          </div>

          <div
            v-for="item in logData"
            :key="item.id"
            class="log-stream__row"
            :class="{ 'log-stream__row--content-only': showOnlyContent }"
          >
            <div class="log-content" v-html="convertAnsiToHtml(item.log)"></div>
            <div v-if="!showOnlyContent" class="log-stream__meta">
              <div class="log-stream__labels">
                <v-chip color="info" size="x-small" variant="tonal">
                  {{ formatTime(item.time) }}
                </v-chip>
                <v-chip color="primary" size="x-small" variant="tonal">
                  {{ item.name }}
                </v-chip>
                <v-chip color="secondary" size="x-small" variant="tonal">
                  {{ item.using || "-" }}
                </v-chip>
              </div>
            </div>
          </div>

          <div
            v-if="logData.length === 0"
            class="text-center text-secondary py-10"
          >
            {{ $t("common.noData") }}
          </div>

          <div
            v-if="showOnlyContent"
            class="log-stream__next-page-action log-stream__next-page-action--bottom"
          >
            <v-btn
              icon
              size="small"
              variant="tonal"
              :loading="loading"
              :disabled="loading || !canContinueNextPage"
              @click="goToNextPage"
            >
              <v-icon>mdi-chevron-down</v-icon>
            </v-btn>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="!showOnlyContent" class="text-center pa-3 pa-sm-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages > 400 ? 400 : totalPages"
            :total-visible="paginationVisible"
            density="compact"
            :disabled="loading"
            @update:model-value="handlePageChange"
          ></v-pagination>
          <div class="mt-2 text-caption text-secondary">
            {{ $t("logPage.totalLogs", { n: totalLogs }) }}
          </div>
        </div>
      </div>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useDisplay } from "vuetify";
import { useI18n } from "vue-i18n";
import { getLog } from "~/src/api/log";
import type { GetLogReq, ProcessLog } from "~/src/types/log/log";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import Convert from "ansi-to-html";
import { getProcessList } from "~/src/api/process";

const { t } = useI18n();
const { smAndDown } = useDisplay();
const snackbarStore = useSnackbarStore();

const paginationVisible = computed(() => (smAndDown.value ? 3 : 7));
const ansiConverter = new Convert();

const processList = ref<string[]>([]);

const sortOptions = computed(() => [
  { label: t("logPage.auto"), value: "" },
  { label: t("logPage.timeAsc"), value: "asc" },
  { label: t("logPage.timeDesc"), value: "desc" },
]);

const formatDatetimeLocal = (date: Date) => {
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate(),
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

const getDefaultStartTime = () => {
  const date = new Date();
  date.setMonth(date.getMonth() - 1);
  return formatDatetimeLocal(date);
};

const getDefaultEndTime = () => {
  return formatDatetimeLocal(new Date());
};

// 数据
const logData = ref<ProcessLog[]>([]);
const totalLogs = ref(0);
const currentPage = ref(1);
const pageSize = ref(25);
const loading = ref(false);
const showFilter = ref(true);
const showOnlyContent = ref(false);

// 搜索表单
const searchForm = ref({
  name: [] as string[],
  log: "",
  using: "",
  startTime: getDefaultStartTime(),
  endTime: getDefaultEndTime(),
  sort: "desc",
});

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(totalLogs.value / pageSize.value);
});
const canContinueNextPage = computed(
  () => totalPages.value > 0 && currentPage.value < totalPages.value,
);

// 转换 ANSI 颜色代码为 HTML
const convertAnsiToHtml = (text: string) => {
  if (!text) return "";
  return ansiConverter
    .toHtml(text)
    .replaceAll("color:rgb(255,255,255)", "color:rgb(160,160,160)")
    .replaceAll("color:#ffffff", "color:#a0a0a0")
    .replaceAll("color:#FFF", "color:#a0a0a0");
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
const buildQuery = (page: number = currentPage.value): GetLogReq => {
  const query: GetLogReq = {
    page: {
      from: (page - 1) * pageSize.value,
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
const loadLogs = async (options?: { append?: boolean; page?: number }) => {
  const append = options?.append ?? false;
  const targetPage = options?.page ?? currentPage.value;
  loading.value = true;
  try {
    const query = buildQuery(targetPage);
    const response = await getLog(query);

    if (response.code === 0 && response.data) {
      const incomingLogs = response.data.data || [];
      logData.value = append ? [...logData.value, ...incomingLogs] : incomingLogs;
      currentPage.value = targetPage;
      totalLogs.value = response.data.total || 0;
    } else {
      snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
    }
  } catch (error) {
    console.error("加载日志错误:", error);
    snackbarStore.showWarningMessage(t("logPage.noLogsRetrieved"));
  } finally {
    loading.value = false;
  }
};

// 搜索日志
const searchLogs = () => {
  loadLogs({ page: 1 });
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
  loadLogs({ page: 1 });
};

// 刷新日志
const refreshLogs = () => {
  loadLogs();
};

// 处理页码变化
const handlePageChange = (page: number) => {
  loadLogs({ page });
};

const goToNextPage = () => {
  if (!canContinueNextPage.value) return;
  loadLogs({ append: true, page: currentPage.value + 1 });
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
  if (smAndDown.value) {
    showFilter.value = false;
  }
  loadProcessList();
  loadLogs();
});
</script>

<style lang="scss" scoped>
.log-content {
  font-size: 0.78rem;
  padding: 2px 0;
  border-radius: 0;
  font-family: "Consolas", "Monaco", "Courier New", monospace;
  max-width: 100%;
  line-height: 1.32;
  display: block;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

/* ANSI 颜色样式 */
:deep(.log-content span) {
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.log-page__title {
  font-size: 1.1rem !important;
}

@media (max-width: 600px) {
  .log-page__title {
    font-size: 1rem !important;
  }
}

.log-stream {
  position: relative;
  border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  border-radius: 8px;
  overflow: hidden;
  background: rgb(var(--v-theme-surface));
}

.log-stream__header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 10px;
  align-items: center;
  text-align: center;
  padding: 8px 10px;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgba(var(--v-theme-on-surface), 0.62);
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
  background: rgba(var(--v-theme-on-surface), 0.03);
}

.log-stream__row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 250px;
  gap: 10px;
  align-items: start;
  padding: 6px 10px;
  font-variant-numeric: tabular-nums;
}

.log-stream__row + .log-stream__row {
  border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.log-stream__row--content-only {
  grid-template-columns: minmax(0, 1fr);
  gap: 0;
}

.log-stream--content-only .log-stream__row {
  padding: 0;
}

.log-stream--content-only .log-stream__row:hover {
  background: transparent;
}

.log-stream--content-only .log-content {
  padding: 0;
  line-height: 1.28;
}

.log-stream__row:hover {
  background: rgba(var(--v-theme-on-surface), 0.03);
}

.log-stream__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.log-stream__time {
  white-space: nowrap;
  line-height: 1.4;
}

.log-stream__labels {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.log-stream__next-page-action {
  display: flex;
  justify-content: center;
}

.log-stream__next-page-action--top {
  padding: 8px 0 2px;
}

.log-stream__next-page-action--bottom {
  padding: 4px 0 8px;
}

@media (max-width: 960px) {
  .log-stream__header,
  .log-stream__row {
    grid-template-columns: minmax(0, 1fr) 240px;
  }
}

@media (max-width: 600px) {
  .log-stream {
    border-radius: 6px;
  }

  .log-stream__row {
    grid-template-columns: minmax(0, 1fr);
    gap: 6px;
    padding: 8px 10px;
  }

  .log-stream__time {
    font-size: 0.7rem !important;
  }
}
</style>
