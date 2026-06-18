<template>
  <v-container fluid class="log-page py-4 py-sm-6 px-3 px-sm-6 px-md-8">
    <v-card class="rounded-lg">
      <!-- loading spinner -->
      <div
        v-if="loading && logData.length === 0"
        class="h-full d-flex flex-grow-1 align-center justify-center"
        style="min-height: 400px"
      >
        <v-progress-circular indeterminate color="primary" />
      </div>

      <div v-else>
        <!-- 标题栏 -->
        <h6
          class="log-page__title text-h6 font-weight-bold pa-4 pa-sm-5 d-flex align-center"
        >
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
        </h6>

        <!-- 日志内容搜索框（始终显示） -->
        <div class="px-4 px-sm-5 pb-3 pt-1">
          <v-row dense class="align-center">
            <v-col cols="12" sm="7" md="8">
              <v-text-field
                :label="$t('logPage.logContent')"
                variant="outlined"
                density="compact"
                v-model="searchForm.log"
                clearable
                hide-details
                prepend-inner-icon="mdi-magnify"
                @keyup.enter="searchLogs"
              />
            </v-col>
            <v-col
              cols="12"
              sm="5"
              md="4"
              class="d-flex ga-2 align-center justify-end"
            >
              <v-btn
                color="primary"
                size="small"
                variant="elevated"
                elevation="2"
                :loading="loading"
                :disabled="loading"
                @click="searchLogs"
              >
                <v-icon start>mdi-magnify</v-icon>
                {{ $t("common.search") }}
              </v-btn>
              <v-btn
                size="small"
                variant="tonal"
                :disabled="loading"
                @click="resetSearch"
              >
                <v-icon start>mdi-refresh</v-icon>
                {{ $t("common.reset") }}
              </v-btn>
              <v-switch
                v-model="searchForm.hightLight"
                :label="$t('logPage.highLight')"
                hide-details
                color="primary"
              />
            </v-col>
          </v-row>
        </div>

        <!-- 更多筛选条件（可折叠） -->
        <v-expand-transition>
          <div v-show="showFilter" class="px-4 px-sm-5 pb-4">
            <v-row dense>
              <v-col cols="12" sm="6" md="4">
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
              <v-col cols="12" sm="6" md="4">
                <v-text-field
                  :label="$t('logPage.user')"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.using"
                  clearable
                  hide-details
                />
              </v-col>
              <v-col cols="12" sm="6" md="4">
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
              <v-col cols="12" sm="6" md="4">
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
              <v-col cols="12" sm="6" md="4">
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
            </v-row>
          </div>
        </v-expand-transition>

        <!-- 日志列表 -->
        <div class="log-stream px-2 px-sm-4 pb-2 pb-sm-4">
          <v-overlay
            :model-value="loading && logData.length > 0"
            contained
            class="align-center justify-center"
            persistent
          >
            <v-progress-circular indeterminate color="primary" />
          </v-overlay>

          <div class="log-stream__header" v-if="!smAndDown">
            <span>{{ $t("logPage.logContent") }}</span>
            <span
              >{{ $t("common.time") }} / {{ $t("logPage.processName") }} /
              {{ $t("logPage.user") }}</span
            >
          </div>

          <div
            v-for="item in logData"
            :key="item.id"
            class="log-stream__row"
            @mouseenter="hoveredRowId = item.id ?? null"
            @mouseleave="hoveredRowId = null"
          >
            <div class="log-content" v-html="convertAnsiToHtml(item.log)"></div>
            <div class="log-stream__meta">
              <div class="log-stream__labels">
                <v-chip color="info" size="x-small" variant="tonal">{{
                  formatTime(item.time)
                }}</v-chip>
                <v-chip color="primary" size="x-small" variant="tonal">{{
                  item.name
                }}</v-chip>
                <v-chip color="secondary" size="x-small" variant="tonal">{{
                  item.using || "-"
                }}</v-chip>
              </div>
            </div>

            <!-- 悬停浮层：上下文操作按钮 -->
            <div
              v-if="hoveredRowId === item.id"
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
                    @click.stop="contextViewer?.openAbove(item)"
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
                    @click.stop="contextViewer?.openBelow(item)"
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
        </div>

        <!-- 分页 -->
        <div class="text-center pa-3 pa-sm-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages > 400 ? 400 : totalPages"
            :total-visible="paginationVisible"
            density="compact"
            :disabled="loading"
            @update:model-value="handlePageChange"
          />
          <div class="mt-2 text-caption text-secondary">
            {{ $t("logPage.totalLogs", { n: totalLogs }) }}
          </div>
        </div>
      </div>
    </v-card>

    <!-- xterm 上下文查看器 -->
    <LogContextViewer ref="contextViewer" />
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useDisplay } from "vuetify";
import { useI18n } from "vue-i18n";
import { getLog } from "~/src/api/log";
import type { GetLogReq, ProcessLog } from "~/src/types/log/log";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { AnsiUp } from "ansi_up";
import { getProcessList } from "~/src/api/process";
import LogContextViewer from "~/src/components/log/LogContextViewer.vue";

const { t } = useI18n();
const { smAndDown } = useDisplay();
const snackbarStore = useSnackbarStore();

const ansiConverter = new AnsiUp();

const paginationVisible = computed(() => (smAndDown.value ? 3 : 7));
const processList = ref<string[]>([]);
const contextViewer = ref<InstanceType<typeof LogContextViewer> | null>(null);
const hoveredRowId = ref<number | null>(null);

const sortOptions = computed(() => [
  { label: t("logPage.auto"), value: "" },
  { label: t("logPage.timeAsc"), value: "asc" },
  { label: t("logPage.timeDesc"), value: "desc" },
]);

const formatDatetimeLocal = (date: Date) => {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

const getDefaultStartTime = () => {
  const d = new Date();
  d.setMonth(d.getMonth() - 1);
  return formatDatetimeLocal(d);
};

// ── 数据 ──────────────────────────────────────────────────────────────────
const logData = ref<ProcessLog[]>([]);
const totalLogs = ref(0);
const currentPage = ref(1);
const pageSize = ref(25);
const loading = ref(false);
const showFilter = ref(false);

const searchForm = ref({
  name: [] as string[],
  log: "",
  using: "",
  startTime: getDefaultStartTime(),
  endTime: formatDatetimeLocal(new Date()),
  sort: "desc",
  hightLight: true,
});

const totalPages = computed(() => Math.ceil(totalLogs.value / pageSize.value));

// ── 工具函数 ──────────────────────────────────────────────────────────────
const convertAnsiToHtml = (text: string) => {
  if (!text) return "";
  return ansiConverter
    .ansi_to_html(text)
    .replaceAll("color:rgb(255,255,255)", "color:rgb(160,160,160)");
};

const formatTime = (ts: number) => {
  if (!ts) return "-";
  return new Date(ts).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

// ── 查询 ──────────────────────────────────────────────────────────────────
const buildQuery = (page: number = currentPage.value): GetLogReq => {
  const query: GetLogReq = {
    page: { from: (page - 1) * pageSize.value, size: pageSize.value },
  };
  if (searchForm.value.sort) query.sort = searchForm.value.sort;
  const match: any = {};
  if (searchForm.value.name.length > 0)
    query.filterName = searchForm.value.name;
  if (searchForm.value.log) match.log = searchForm.value.log;
  if (searchForm.value.using) match.using = searchForm.value.using;
  match.highLight = searchForm.value.hightLight;
  if (Object.keys(match).length > 0) query.match = match;
  if (searchForm.value.startTime || searchForm.value.endTime) {
    query.time = {};
    if (searchForm.value.startTime)
      query.time.startTime = new Date(searchForm.value.startTime).getTime();
    if (searchForm.value.endTime)
      query.time.endTime = new Date(searchForm.value.endTime).getTime();
  }
  return query;
};

const loadLogs = async (options?: { page?: number }) => {
  const targetPage = options?.page ?? currentPage.value;
  loading.value = true;
  try {
    const res = await getLog(buildQuery(targetPage));
    if (res.code === 0 && res.data) {
      logData.value = res.data.data || [];
      currentPage.value = targetPage;
      totalLogs.value = res.data.total || 0;
    } else {
      snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
    }
  } catch {
    snackbarStore.showWarningMessage(t("logPage.noLogsRetrieved"));
  } finally {
    loading.value = false;
  }
};

const searchLogs = () => loadLogs({ page: 1 });
const resetSearch = () => {
  searchForm.value = {
    name: [] as string[],
    log: "",
    using: "",
    startTime: getDefaultStartTime(),
    endTime: formatDatetimeLocal(new Date()),
    sort: "desc",
    hightLight: true,
  };
  loadLogs({ page: 1 });
};
const refreshLogs = () => loadLogs();
const handlePageChange = (page: number) => loadLogs({ page });

const loadProcessList = async () => {
  try {
    const res = await getProcessList();
    if (res.code === 0 && res.data) {
      processList.value = Array.from(
        new Set(res.data.map((i) => i.name)),
      ).sort();
    }
  } catch {
    // ignore
  }
};

onMounted(() => {
  if (smAndDown.value) showFilter.value = false;
  loadProcessList();
  loadLogs();
});
</script>

<style lang="scss" scoped>
.log-content {
  font-size: 0.78rem;
  padding: 2px 0;
  font-family: "Consolas", "Monaco", "Courier New", monospace;
  max-width: 100%;
  line-height: 1.32;
  display: block;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

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

/* 悬停浮层上下文按钮 */
.log-row__context-actions {
  position: absolute;
  top: 50%;
  right: 8px;
  transform: translateY(-50%);
  display: flex;
  gap: 4px;
  z-index: 2;
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
