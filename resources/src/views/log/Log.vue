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
          <!-- 上下文模式指示器 -->
          <v-chip
            v-if="isContextMode"
            size="x-small"
            color="primary"
            variant="tonal"
            class="mr-2"
            closable
            @click:close="exitContextMode"
          >
            {{ $t("logPage.contextMode") }}
          </v-chip>
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
          :class="{ 'log-stream--context': isContextMode }"
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

          <div class="log-stream__header" v-if="!smAndDown && !isContextMode">
            <span>{{ $t("logPage.logContent") }}</span>
            <span>{{ $t("common.time") }} / {{ $t("logPage.processName") }} / {{ $t("logPage.user") }}</span>
          </div>

          <!-- 上文加载按钮：仅在上下文模式下显示 -->
          <div
            v-if="isContextMode"
            class="log-stream__next-page-action log-stream__next-page-action--top"
          >
            <v-btn
              size="small"
              variant="tonal"
              :loading="loadingAbove"
              :disabled="loadingAbove || !hasMoreAbove"
              prepend-icon="mdi-chevron-double-up"
              @click="loadMoreAbove"
            >
              {{ hasMoreAbove ? $t("logPage.loadMoreAbove") : $t("logPage.noMoreAbove") }}
            </v-btn>
          </div>

          <div
            v-for="item in logData"
            :key="item.id"
            class="log-stream__row"
            :class="{
              'log-stream__row--context': isContextMode,
              'log-stream__row--context-anchor': isContextMode && item.id === contextAnchorID,
            }"
            @mouseenter="hoveredRowId = item.id ?? null"
            @mouseleave="hoveredRowId = null"
          >
            <div class="log-content" v-html="convertAnsiToHtml(item.log)"></div>
            <div v-if="!isContextMode" class="log-stream__meta">
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

            <!-- 悬停时浮层显示上下文操作图标，仅普通模式下可见 -->
            <div
              v-if="!isContextMode && hoveredRowId === item.id"
              class="log-row__context-actions"
            >
              <v-tooltip :text="$t('logPage.viewContextAbove')" location="top">
                <template #activator="{ props }">
                  <v-btn
                    v-bind="props"
                    icon
                    size="x-small"
                    variant="elevated"
                    density="compact"
                    @click.stop="viewContextAbove(item)"
                  >
                    <v-icon size="14">mdi-arrow-up-circle-outline</v-icon>
                  </v-btn>
                </template>
              </v-tooltip>
              <v-tooltip :text="$t('logPage.viewContextBelow')" location="top">
                <template #activator="{ props }">
                  <v-btn
                    v-bind="props"
                    icon
                    size="x-small"
                    variant="elevated"
                    density="compact"
                    @click.stop="viewContextBelow(item)"
                  >
                    <v-icon size="14">mdi-arrow-down-circle-outline</v-icon>
                  </v-btn>
                </template>
              </v-tooltip>
            </div>
          </div>

          <div
            v-if="logData.length === 0"
            class="text-center text-secondary py-10"
          >
            {{ $t("common.noData") }}
          </div>

          <!-- 下文加载按钮：仅在上下文模式下显示 -->
          <div
            v-if="isContextMode"
            class="log-stream__next-page-action log-stream__next-page-action--bottom"
          >
            <v-btn
              size="small"
              variant="tonal"
              :loading="loadingBelow"
              :disabled="loadingBelow || !hasMoreBelow"
              prepend-icon="mdi-chevron-double-down"
              @click="loadMoreBelow"
            >
              {{ hasMoreBelow ? $t("logPage.loadMoreBelow") : $t("logPage.noMoreBelow") }}
            </v-btn>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="!isContextMode" class="text-center pa-3 pa-sm-4">
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

// 上下文模式状态
const isContextMode = ref(false);
const contextTopID = ref<number>(0);
const contextBottomID = ref<number>(0);
const contextAnchorID = ref<number>(0);
const hasMoreAbove = ref(false);
const hasMoreBelow = ref(false);
const hoveredRowId = ref<number | null>(null);
const loadingAbove = ref(false);
const loadingBelow = ref(false);
const CONTEXT_INIT_SIZE = 50;
const CONTEXT_CHUNK_SIZE = 25;

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

  if (searchForm.value.sort) {
    query.sort = searchForm.value.sort;
  }
  const match: any = {};
  if (searchForm.value.name && searchForm.value.name.length > 0) {
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

// 构建上下文查询（仅基于游标，不带筛选条件）
const buildContextQuery = (cursorID: number, sort: string, size: number): GetLogReq => {
  return {
    cursorID,
    sort,
    page: { from: 0, size },
  };
};

// 加载日志
const loadLogs = async (options?: { page?: number }) => {
  const targetPage = options?.page ?? currentPage.value;
  loading.value = true;
  try {
    const query = buildQuery(targetPage);
    const response = await getLog(query);
    if (response.code === 0 && response.data) {
      logData.value = response.data.data || [];
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

// 进入上下文模式：查看指定日志的上文
const viewContextAbove = async (item: ProcessLog) => {
  isContextMode.value = true;
  contextAnchorID.value = item.id!;
  loading.value = true;
  try {
    const query = buildContextQuery(item.id!, "desc", CONTEXT_INIT_SIZE);
    const response = await getLog(query);
    if (response.code === 0 && response.data) {
      const fetched = response.data.data || [];
      // DESC 返回（ID 从大到小），翻转为正序展示，锚点行附在末尾
      const logs = [...fetched].reverse();
      logData.value = [...logs, item];
      contextTopID.value = logs.length > 0 ? logs[0].id! : item.id!;
      contextBottomID.value = item.id!;
      hasMoreAbove.value = fetched.length >= CONTEXT_INIT_SIZE;
      hasMoreBelow.value = true;
    } else {
      snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
    }
  } catch (error) {
    console.error("加载上文错误:", error);
    snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
  } finally {
    loading.value = false;
  }
};

// 进入上下文模式：查看指定日志的下文
const viewContextBelow = async (item: ProcessLog) => {
  isContextMode.value = true;
  contextAnchorID.value = item.id!;
  loading.value = true;
  try {
    const query = buildContextQuery(item.id!, "asc", CONTEXT_INIT_SIZE);
    const response = await getLog(query);
    if (response.code === 0 && response.data) {
      const fetched = response.data.data || [];
      // ASC 返回（ID 从小到大），锚点行置顶
      logData.value = [item, ...fetched];
      contextTopID.value = item.id!;
      contextBottomID.value = fetched.length > 0 ? fetched[fetched.length - 1].id! : item.id!;
      hasMoreAbove.value = true;
      hasMoreBelow.value = fetched.length >= CONTEXT_INIT_SIZE;
    } else {
      snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
    }
  } catch (error) {
    console.error("加载下文错误:", error);
    snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
  } finally {
    loading.value = false;
  }
};

// 在上下文模式下加载更多上文（向上追加）
const loadMoreAbove = async () => {
  if (loadingAbove.value || !hasMoreAbove.value) return;
  loadingAbove.value = true;
  try {
    const query = buildContextQuery(contextTopID.value, "desc", CONTEXT_CHUNK_SIZE);
    const response = await getLog(query);
    if (response.code === 0 && response.data) {
      const fetched = response.data.data || [];
      const logs = [...fetched].reverse();
      logData.value = [...logs, ...logData.value];
      if (logs.length > 0) {
        contextTopID.value = logs[0].id!;
      }
      hasMoreAbove.value = fetched.length >= CONTEXT_CHUNK_SIZE;
    }
  } catch (error) {
    console.error("加载更多上文错误:", error);
  } finally {
    loadingAbove.value = false;
  }
};

// 在上下文模式下加载更多下文（向下追加）
const loadMoreBelow = async () => {
  if (loadingBelow.value || !hasMoreBelow.value) return;
  loadingBelow.value = true;
  try {
    const query = buildContextQuery(contextBottomID.value, "asc", CONTEXT_CHUNK_SIZE);
    const response = await getLog(query);
    if (response.code === 0 && response.data) {
      const fetched = response.data.data || [];
      logData.value = [...logData.value, ...fetched];
      if (fetched.length > 0) {
        contextBottomID.value = fetched[fetched.length - 1].id!;
      }
      hasMoreBelow.value = fetched.length >= CONTEXT_CHUNK_SIZE;
    }
  } catch (error) {
    console.error("加载更多下文错误:", error);
  } finally {
    loadingBelow.value = false;
  }
};

// 退出上下文模式，回到普通分页视图
const exitContextMode = () => {
  isContextMode.value = false;
  loadLogs();
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

// 加载进程列表
const loadProcessList = async () => {
  try {
    const response = await getProcessList();
    if (response.code === 0 && response.data) {
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
  position: relative;
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

/* 上下文模式：单列布局，隐藏 meta */
.log-stream__row--context {
  grid-template-columns: minmax(0, 1fr);
  gap: 0;
  padding: 0;
}

/* 上下文模式下 hover 无背景变化 */
.log-stream--context .log-stream__row:hover {
  background: transparent;
}

.log-stream--context .log-content {
  padding: 0;
  line-height: 1.28;
}

/* 锚点行高亮 */
.log-stream__row--context-anchor {
  background: rgba(var(--v-theme-primary), 0.06) !important;
  border-left: 2px solid rgb(var(--v-theme-primary));
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

.log-stream__labels {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

/* 悬停时浮层显示的上下文操作按钮 */
.log-row__context-actions {
  position: absolute;
  top: 50%;
  right: 8px;
  transform: translateY(-50%);
  display: flex;
  gap: 4px;
  z-index: 2;
  pointer-events: auto;
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
}
</style>
