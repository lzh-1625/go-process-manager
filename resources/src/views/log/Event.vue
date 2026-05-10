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
          <v-icon color="primary" class="mr-2">mdi-bell-ring</v-icon>
          <span class="flex-fill">{{ $t('eventPage.title') }}</span>
          <v-btn
            icon
            variant="text"
            size="small"
            @click="refreshEvents"
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
          <div v-show="showFilter" class="px-5 pb-4">
            <v-row dense>
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  :label="$t('common.name')"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.name"
                  clearable
                  hide-details
                />
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-select
                  :label="$t('eventPage.eventType')"
                  variant="outlined"
                  density="compact"
                  v-model="searchForm.type"
                  :items="eventTypes"
                  item-title="label"
                  item-value="value"
                  clearable
                  hide-details
                />
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  :label="$t('common.startTime')"
                  variant="outlined"
                  density="compact"
                  type="datetime-local"
                  v-model="searchForm.startTime"
                  clearable
                  hide-details
                />
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-text-field
                  :label="$t('common.endTime')"
                  variant="outlined"
                  density="compact"
                  type="datetime-local"
                  v-model="searchForm.endTime"
                  clearable
                  hide-details
                />
              </v-col>
              <v-col cols="12" class="d-flex ga-2 mt-2">
                <v-btn
                  color="primary"
                  size="small"
                  elevation="4"
                  variant="elevated"
                  @click="searchEvents"
                >
                  {{ $t('common.search') }}
                </v-btn>
                <v-btn size="small" variant="tonal" @click="resetSearch">
                  {{ $t('common.reset') }}
                </v-btn>
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>

        <!-- 事件列表 -->
        <v-table class="pa-3">
          <thead>
            <tr>
              <th class="text-left" v-for="header in headers" :key="header.text">
                {{ header.text }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in eventData" :key="item.id">
              <td>
                <v-chip
                  :color="getEventTypeColor(item.type)"
                  size="small"
                  class="font-weight-bold"
                >
                  <v-icon start size="small">{{
                    getEventTypeIcon(item.type)
                  }}</v-icon>
                  {{ getEventTypeLabel(item.type) }}
                </v-chip>
              </td>
              <td class="font-weight-bold">
                <v-chip color="primary" size="small" class="font-weight-bold">
                  {{ item.name }}
                </v-chip>
              </td>
              <td>
                <div v-if="item.additional" class="additional-info">
                  <template
                    v-for="(value, key) in parseAdditional(item.additional)"
                    :key="key"
                  >
                    <v-chip size="x-small" variant="outlined" class="mr-1 mb-1">
                      {{ key }}: {{ value }}
                    </v-chip>
                  </template>
                </div>
                <span v-else class="text-secondary">-</span>
              </td>
              <td>{{ formatTime(item.createdTime) }}</td>
            </tr>
            <tr v-if="eventData.length === 0">
              <td colspan="4" class="text-center text-secondary pa-8">
                {{ $t('common.noData') }}
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
            {{ $t('eventPage.totalEvents', { n: totalEvents }) }}
          </div>
        </div>
      </div>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { getEventList } from "~/src/api/event";
import type { Event, EventListReq, EventType } from "~/src/types/event/event";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();

const headers = computed(() => [
  { text: t("eventPage.eventType"), align: "start", value: "type" },
  { text: t("common.name"), sortable: false, value: "name" },
  { text: t("eventPage.additionalInfo"), sortable: false, value: "additional" },
  { text: t("common.time"), value: "createdTime" },
]);

const eventTypes = computed(() => [
  { label: t("eventPage.processStart"), value: "ProcessStart" },
  { label: t("eventPage.processStop"), value: "ProcessStop" },
  { label: t("eventPage.apiRequest"), value: "ApiRequest" },
  { label: t("eventPage.processWarning"), value: "ProcessWarning" },
  { label: t("eventPage.taskStart"), value: "TaskStart" },
  { label: t("eventPage.taskStop"), value: "TaskStop" },
]);

// 数据
const eventData = ref<Event[]>([]);
const totalEvents = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const showFilter = ref(false);

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
    ProcessConnect: "info",
    TaskStart: "info",
    ApiRequest: "primary",
    TaskStop: "secondary",
  };
  return colorMap[type] || "grey";
};

// 获取事件类型图标
const getEventTypeIcon = (type: EventType) => {
  const iconMap: Record<EventType, string> = {
    ProcessStart: "mdi-play-circle",
    ProcessStop: "mdi-stop-circle",
    ApiRequest: "mdi-api",
    ProcessWarning: "mdi-alert-circle",
    ProcessConnect: "mdi-console",
    TaskStart: "mdi-clock-start",
    TaskStop: "mdi-clock-end",
  };
  return iconMap[type] || "mdi-information";
};

// 获取事件类型标签
const getEventTypeLabel = (type: EventType) => {
  const labelMap: Record<string, string> = {
    ProcessStart: t("eventPage.processStart"),
    ProcessStop: t("eventPage.processStop"),
    ProcessWarning: t("eventPage.processWarning"),
    ApiRequest: t("eventPage.apiRequest"),
    ProcessConnect: t("eventPage.processConnect"),
    TaskStart: t("eventPage.taskStart"),
    TaskStop: t("eventPage.taskStop"),
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
      snackbarStore.showErrorMessage(t("eventPage.loadEventsFailed"));
    }
  } catch (error) {
    console.error("加载事件错误:", error);
    snackbarStore.showErrorMessage(t("eventPage.loadEventsError"));
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

<style lang="scss" scoped>
.additional-info {
  max-width: 100%;
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
