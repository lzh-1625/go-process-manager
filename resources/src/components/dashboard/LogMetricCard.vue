<!--
* @Component: LogMetricCard
* @Description: 日志统计折线图，支持小时、日、周、月四个时间单位
-->
<script setup lang="ts">
import { ref, onMounted, computed, Ref } from "vue";
import { useI18n } from "vue-i18n";
import { useTheme } from "vuetify";
import type { EChartsOption } from "echarts";
import { useChart, RenderType, ThemeType } from "@/plugins/echarts";
import { getLogMetric, LogStatsticMetric, LogStatsticMetricItem } from "@/api/metric";
import {
  getChartTooltipStyle,
  getChartTooltipOverflowBehavior,
  logMetricProcessingBadgeStyle,
} from "./chartAppearance";

const { t } = useI18n();
const theme = useTheme();

const loading = ref(true);
const logData = ref<LogStatsticMetric | null>(null);
const dateType = ref(0); // 0: 小时, 1: 日, 2: 周, 3: 月
const showTotal = ref(false);

const dateTypes = computed(() => [
  { value: 0, title: t("dashboardPage.hour") },
  { value: 1, title: t("dashboardPage.day") },
  { value: 2, title: t("dashboardPage.week") },
  { value: 3, title: t("dashboardPage.month") },
]);

const chartEl = ref<HTMLDivElement | null>(null);

const processEntries = computed(() =>
  Object.entries(logData.value?.items ?? {}).filter(([, items]) => items.length > 0)
);

const hasLogData = computed(() => processEntries.value.length > 0);
const chartTextColor = computed(() =>
  theme.global.current.value.dark ? "#E7E9F6" : "#2F2B3D"
);

const totalLogItems = computed<LogStatsticMetricItem[]>(() => {
  const [, firstProcessItems] = processEntries.value[0] ?? [];
  if (!firstProcessItems) return [];

  return firstProcessItems.map((item, index) => ({
    date: item.date,
    count: processEntries.value.reduce(
      (total, [, items]) => total + (items[index]?.count ?? 0),
      0
    ),
  }));
});

const chartEntries = computed<[string, LogStatsticMetricItem[]][]>(() => {
  if (showTotal.value) {
    return [[t("dashboardPage.totalLogCount"), totalLogItems.value]];
  }
  return processEntries.value;
});

const chartOption = computed<EChartsOption>(() => {
  if (!hasLogData.value) return {};

  const [, firstProcessItems] = processEntries.value[0];
  const dates = [...firstProcessItems].reverse().map((item) => item.date);

  return {
    title: {
      text: `${t("dashboardPage.logStatistics")} (${
        dateTypes.value.find((type) => type.value === dateType.value)?.title
      })`,
      left: "center",
      top: 10,
      textStyle: {
        fontSize: 16,
        fontWeight: "bold",
      },
      subtextStyle: {
        fontSize: 12,
      },
    },
    legend: {
      type: "scroll",
      top: 42,
      textStyle: {
        color: chartTextColor.value,
      },
    },
    tooltip: {
      trigger: "axis",
      ...getChartTooltipStyle(theme.global.current.value.dark),
      ...getChartTooltipOverflowBehavior(),
      axisPointer: {
        type: "cross",
        label: {
          backgroundColor: "#6a7985",
        },
      },
      formatter: (params: any) => {
        const [firstParam] = params;
        return `${firstParam.name}<br/>${params
          .map(
            (param: any) =>
              `${param.marker}${param.seriesName}: ${param.value}`
          )
          .join("<br/>")}`;
      },
    },
    grid: {
      left: "3%",
      right: "4%",
      bottom: "10%",
      top: "28%",
      containLabel: true,
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: dates,
      axisLabel: {
        rotate: dateType.value <= 1 ? 45 : 0,
      },
    },
    yAxis: [
      {
        type: "value",
        name: t("dashboardPage.logCount"),
        min: 0,
        minInterval: 1,
        position: "right",
        axisLine: {
          show: true,
        },
      }
    ],
    series: chartEntries.value.map(([processName, items]) => ({
        name: processName,
        type: "line",
        smooth: true,
        symbol: "circle",
        symbolSize: 8,
        lineStyle: {
          width: 3,
        },
        itemStyle: {
          borderColor: "#fff",
          borderWidth: 2,
        },
        data: [...items].reverse().map((item) => item.count),
        emphasis: {
          focus: "series",
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: "rgba(0, 0, 0, 0.5)",
          },
        },
      })),
  };
});

const { setOption, getInstance } = useChart(
  chartEl as Ref<HTMLDivElement>,
  false,
  false,
  RenderType.SVGRenderer,
  ThemeType.Default
);

const loadData = async () => {
  loading.value = true;
  let loaded = false;
  try {
    const response = await getLogMetric(dateType.value);
    logData.value = response.data as LogStatsticMetric;
    loaded = true;
  } catch (error) {
    console.error("Failed to load log metric data:", error);
  } finally {
    loading.value = false;
  }

  if (loaded && hasLogData.value) {
    await nextTick();
    setOption(chartOption.value, true);
  }
};

const onDateTypeChange = () => {
  loadData();
};

const handleResize = () => {
  getInstance()?.resize();
};

// 监听图表容器大小变化
let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
  setTimeout(async () => {
    await loadData();

    // 数据加载完成后，监听图表容器大小变化
    await nextTick();
    if (chartEl.value) {
      resizeObserver = new ResizeObserver(() => {
        handleResize();
      });
      resizeObserver.observe(chartEl.value);
    }
  }, 500);

  // 监听窗口大小变化
  window.addEventListener("resize", handleResize);

  onUnmounted(() => {
    window.removeEventListener("resize", handleResize);
    if (resizeObserver) {
      resizeObserver.disconnect();
    }
  });
});

watch(
  () => chartOption.value,
  (newVal) => {
    if (logData.value) {
      setOption(newVal, true);
    }
  },
  { deep: true }
);
</script>

<template>
  <div>
    <v-card-title class="d-flex justify-space-between align-center pa-5">
      <span class="text-h6 font-weight-bold">{{ $t("dashboardPage.logStatistics") }}</span>
      <div class="d-flex align-center justify-end flex-wrap ga-3">
        <v-switch
          v-model="showTotal"
          :label="$t('dashboardPage.totalLogCount')"
          color="primary"
          density="compact"
          hide-details
        ></v-switch>
        <v-btn-toggle
          v-model="dateType"
          color="primary"
          mandatory
          density="compact"
          @update:model-value="onDateTypeChange"
        >
          <v-btn
            v-for="type in dateTypes"
            :key="type.value"
            :value="type.value"
            size="small"
          >
            {{ type.title }}
          </v-btn>
        </v-btn-toggle>
        <v-chip
          v-if="hasLogData"
          color="warning"
          variant="tonal"
          size="small"
          :style="logMetricProcessingBadgeStyle"
        >
          <v-icon size="16">mdi-progress-clock</v-icon>
          <span class="text-caption">{{ $t("dashboardPage.processing") }}</span>
          <span class="font-weight-bold">{{ logData?.executing }}</span>
        </v-chip>
      </div>
    </v-card-title>
    <v-card-text>
      <div style="position: relative">
        <template v-if="hasLogData">
          <div ref="chartEl" style="width: 100%; height: 350px"></div>
        </template>
        <div
          v-else
          class="d-flex align-center justify-center text-secondary"
          style="height: 350px"
        >
          {{ $t("common.noData") }}
        </div>
        <div
          v-if="loading"
          class="log-metric-loading d-flex align-center justify-center"
        >
          <v-progress-circular indeterminate color="primary"></v-progress-circular>
        </div>
      </div>
    </v-card-text>
  </div>
</template>

<style lang="scss" scoped>
.log-metric-loading {
  position: absolute;
  inset: 0;
  min-height: 350px;
  z-index: 20;
  background: rgba(255, 255, 255, 0.7);
}
</style>
