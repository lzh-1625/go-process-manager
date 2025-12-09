<!--
* @Component: LogMetricCard
* @Description: 日志统计折线图，支持日、周、月三个时间单位
-->
<script setup lang="ts">
import { ref, onMounted, computed, Ref } from "vue";
import type { EChartsOption } from "echarts";
import { useChart, RenderType, ThemeType } from "@/plugins/echarts";
import { getLogMetric, LogStatsticMetric } from "@/api/metric";

const loading = ref(true);
const logData = ref<LogStatsticMetric | null>(null);
const dateType = ref(1); // 1: 日, 2: 周, 3: 月

const dateTypes = [
  { value: 1, title: "日" },
  { value: 2, title: "周" },
  { value: 3, title: "月" },
];

const chartEl = ref<HTMLDivElement | null>(null);

const chartOption = computed<EChartsOption>(() => {
  if (!logData.value || !logData.value.items) return {};

  // 反转数组以按时间正序显示
  const items = [...logData.value.items].reverse();
  const dates = items.map((item) => item.date);
  const counts = items.map((item) => item.count);

  return {
    title: {
      text: `日志统计 (${
        dateTypes.find((t) => t.value === dateType.value)?.title
      })`,
      left: "center",
      top: 10,
      textStyle: {
        fontSize: 16,
        fontWeight: "bold",
      },
      subtext: `正在执行: ${logData.value.executing} 个任务`,
      subtextStyle: {
        fontSize: 12,
      },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "cross",
        label: {
          backgroundColor: "#6a7985",
        },
      },
      formatter: (params: any) => {
        const param = params[0];
        return `${param.name}<br/>${param.marker}日志数量: ${param.value}`;
      },
    },
    grid: {
      left: "3%",
      right: "4%",
      bottom: "10%",
      top: "20%",
      containLabel: true,
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: dates,
      axisLabel: {
        rotate: dateType.value === 1 ? 45 : 0,
      },
    },
    yAxis: [
      {
        type: "value",
        name: "日志数量",
        min: 0,
        position: "right",
        axisLine: {
          show: true,
        },
      }
    ],
    series: [
      {
        name: "日志数量",
        type: "line",
        smooth: true,
        symbol: "circle",
        symbolSize: 8,
        lineStyle: {
          width: 3,
          color: "#03a9f4",
        },
        itemStyle: {
          color: "#03a9f4",
          borderColor: "#fff",
          borderWidth: 2,
        },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              {
                offset: 0,
                color: "rgba(3, 169, 244, 0.3)",
              },
              {
                offset: 1,
                color: "rgba(3, 169, 244, 0.05)",
              },
            ],
          },
        },
        data: counts,
        emphasis: {
          focus: "series",
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: "rgba(3, 169, 244, 0.5)",
          },
        },
      },
    ],
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
  try {
    const response = await getLogMetric(dateType.value);
    logData.value = response.data as LogStatsticMetric;

    await nextTick();
    setOption(chartOption.value);
  } catch (error) {
    console.error("Failed to load log metric data:", error);
  }
};

const onDateTypeChange = () => {
  loadData();
};

const handleResize = () => {
  getInstance()?.resize();
};

onMounted(() => {
  setTimeout(async () => {
    loading.value = false;
    await nextTick();
    // 等待 DOM 渲染后再加载数据
    await loadData();
  }, 500);

  // 监听窗口大小变化
  window.addEventListener("resize", handleResize);

  onUnmounted(() => {
    window.removeEventListener("resize", handleResize);
  });
});

watch(
  () => chartOption.value,
  (newVal) => {
    if (logData.value) {
      setOption(newVal);
    }
  },
  { deep: true }
);
</script>

<template>
  <div>
    <v-card-title class="d-flex justify-space-between align-center pa-5">
      <span class="text-h6 font-weight-bold">日志统计趋势</span>
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
    </v-card-title>
    <v-card-text>
      <div
        v-if="loading"
        class="h-full d-flex align-center justify-center"
        style="min-height: 300px"
      >
        <v-progress-circular
          indeterminate
          color="primary"
        ></v-progress-circular>
      </div>
      <div v-else style="position: relative">
        <div
          v-if="logData"
          style="
            position: absolute;
            top: 10px;
            left: 20px;
            z-index: 10;
            background: rgba(255, 152, 0, 0.1);
            padding: 4px 12px;
            border-radius: 6px;
            border: 1.5px solid #ff9800;
          "
        >
          <div style="font-size: 10px; color: #ff9800; font-weight: bold">
            正在处理
          </div>
          <div
            style="
              font-size: 18px;
              color: #ff9800;
              font-weight: bold;
              text-align: center;
            "
          >
            {{ logData.executing }}
          </div>
        </div>
        <div ref="chartEl" style="width: 100%; height: 350px"></div>
      </div>
    </v-card-text>
  </div>
</template>

<style lang="scss" scoped></style>
