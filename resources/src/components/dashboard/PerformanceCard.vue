<!--
* @Component: PerformanceCard
* @Description: CPU和内存使用率饼图
-->
<script setup lang="ts">
import { ref, onMounted, computed, Ref } from "vue";
import { useI18n } from "vue-i18n";
import type { EChartsOption } from "echarts";
import { useChart, RenderType, ThemeType } from "@/plugins/echarts";
import { getPerformceUsage, PerformceUsage } from "@/api/metric";

const { t } = useI18n();

const loading = ref(true);
const performanceData = ref<PerformceUsage | null>(null);

const viewportWidth = ref(window.innerWidth);
const viewportHeight = ref(window.innerHeight);

const chartLayout = computed(() => {
  const w = viewportWidth.value;
  const h = viewportHeight.value;
  if (w > h && h <= 520) return "landscape";
  if (w <= 600) return "mobile";
  return "default";
});

const chartCols = computed(() => (chartLayout.value === "landscape" ? 6 : 12));

const chartHeight = computed(() => {
  const h = viewportHeight.value;
  switch (chartLayout.value) {
    case "landscape":
      return `${Math.min(240, Math.max(160, h - 88))}px`;
    case "mobile":
      return "260px";
    default:
      return "300px";
  }
});

const chartLegend = computed(() => {
  if (chartLayout.value === "landscape") {
    return {
      orient: "horizontal" as const,
      left: "center",
      bottom: 0,
      top: "auto",
      textStyle: { fontSize: 9 },
      type: "scroll" as const,
    };
  }
  if (chartLayout.value === "mobile") {
    return {
      orient: "horizontal" as const,
      left: "center",
      bottom: 0,
      top: "auto",
      textStyle: { fontSize: 10 },
      type: "scroll" as const,
    };
  }
  return {
    orient: "vertical" as const,
    left: "left",
    top: "middle",
    textStyle: { fontSize: 11 },
    type: "scroll" as const,
  };
});

const chartSeriesLayout = computed(() => {
  if (chartLayout.value === "landscape") {
    return {
      radius: ["28%", "50%"],
      center: ["50%", "42%"],
      showLabel: false,
    };
  }
  if (chartLayout.value === "mobile") {
    return {
      radius: ["35%", "62%"],
      center: ["50%", "45%"],
      showLabel: false,
    };
  }
  return {
    radius: ["40%", "70%"],
    center: ["60%", "55%"],
    showLabel: true,
  };
});

// CPU饼图配置
const cpuChartEl = ref<HTMLDivElement | null>(null);
const cpuOption = computed<EChartsOption>(() => {
  if (!performanceData.value) return {};

  // 按CPU使用率排序，取前10个进程
  const sortedItems = [...performanceData.value.items]
    .sort((a, b) => b.cpu - a.cpu)
    .slice(0, 10);

  // 计算其他进程的CPU占用
  const topProcessesCpu = sortedItems.reduce((sum, item) => sum + item.cpu, 0);
  const otherProcessesCpu = Math.max(
    0,
    100 - performanceData.value.cpuFree - topProcessesCpu
  );

  const data = [
    ...performanceData.value.items.map((item, index) => ({
      value: parseFloat(item.cpu.toFixed(2)),
      name: item.name,
    })),
  ];

  // 如果有其他进程，添加"其他"项
  if (otherProcessesCpu > 0.01) {
    data.push({
      value: parseFloat(otherProcessesCpu.toFixed(2)),
      name: t("dashboardPage.otherProcesses"),
    });
  }

  data.push({
    value: parseFloat(performanceData.value.cpuFree.toFixed(2)),
    name: t("dashboardPage.idle"),
  });

  return {
    tooltip: {
      trigger: "item",
      formatter: "{a} ---<br/>{b}: {c}% ({d}%)",
    },
    title: {
      text: t("dashboardPage.cpuUsage"),
      left: "center",
      top: chartLayout.value === "landscape" ? 4 : 10,
      textStyle: {
        fontSize: chartLayout.value === "landscape" ? 13 : 16,
        fontWeight: "bold",
      },
    },
    legend: chartLegend.value,
    series: [
        {
          name: t("dashboardPage.cpuUsage"),
        type: "pie",
        radius: chartSeriesLayout.value.radius,
        center: chartSeriesLayout.value.center,
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderWidth: 2,
        },
        label: {
          show: chartSeriesLayout.value.showLabel,
          position: "outside",
          formatter: (params: any) => {
            return `${params.name}: ${params.value}%`;
          },
          fontSize: 10,
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: "bold",
          },
        },
        labelLine: {
          show: chartSeriesLayout.value.showLabel,
          length: 15,
          length2: 10,
        },
        data,
      },
    ],
  };
});

// 内存饼图配置
const memChartEl = ref<HTMLDivElement | null>(null);
const memOption = computed<EChartsOption>(() => {
  if (!performanceData.value) return {};

  // 按内存使用量排序，显示所有进程
  const sortedItems = [...performanceData.value.items].sort(
    (a, b) => b.mem - a.mem
  );

  // 只统计给定的进程数据和空闲内存（KB转换为MB）
  const data = [
    ...sortedItems.map((item, index) => ({
      value: parseFloat((item.mem / 1024).toFixed(2)),
      name: item.name,
    })),
    // 添加空闲项
    {
      value: parseFloat((performanceData.value.memFree / 1024).toFixed(2)),
      name: t("dashboardPage.idle"),
      itemStyle: { color: "#2196f3" },
    },
  ];

  return {
    tooltip: {
      trigger: "item",
      formatter: "{a} <br/>{b}: {c}MB ({d}%)",
    },
    title: {
      text: t("dashboardPage.memoryUsage"),
      left: "center",
      top: chartLayout.value === "landscape" ? 4 : 10,
      textStyle: {
        fontSize: chartLayout.value === "landscape" ? 13 : 16,
        fontWeight: "bold",
      },
    },
    legend: chartLegend.value,
    series: [
      {
        name: t("dashboardPage.memoryUsage"),
        type: "pie",
        radius: chartSeriesLayout.value.radius,
        center: chartSeriesLayout.value.center,
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderWidth: 2,
        },
        label: {
          show: chartSeriesLayout.value.showLabel,
          position: "outside",
          formatter: (params: any) => {
            return `${params.name}: ${params.value}MB`;
          },
          fontSize: 10,
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: "bold",
          },
        },
        labelLine: {
          show: chartSeriesLayout.value.showLabel,
          length: 15,
          length2: 10,
        },
        data,
      },
    ],
  };
});

const { setOption: setCpuOption, getInstance: getCpuInstance } = useChart(
  cpuChartEl as Ref<HTMLDivElement>,
  false,
  false,
  RenderType.SVGRenderer,
  ThemeType.Default
);

const { setOption: setMemOption, getInstance: getMemInstance } = useChart(
  memChartEl as Ref<HTMLDivElement>,
  false,
  false,
  RenderType.SVGRenderer,
  ThemeType.Default
);

const loadData = async () => {
  try {
    const response = await getPerformceUsage();
    performanceData.value = response.data as PerformceUsage;

    await nextTick();
    setCpuOption(cpuOption.value);
    setMemOption(memOption.value);
  } catch (error) {
    console.error("Failed to load performance data:", error);
  }
};

const handleResize = () => {
  viewportWidth.value = window.innerWidth;
  viewportHeight.value = window.innerHeight;
  getCpuInstance()?.resize();
  getMemInstance()?.resize();
};

// 监听图表容器大小变化
let cpuResizeObserver: ResizeObserver | null = null;
let memResizeObserver: ResizeObserver | null = null;

onMounted(() => {
  setTimeout(async () => {
    loading.value = false;
    await nextTick();
    // 等待 DOM 渲染后再加载数据
    await loadData();

    // 数据加载完成后，监听图表容器大小变化
    await nextTick();
    if (cpuChartEl.value) {
      cpuResizeObserver = new ResizeObserver(() => {
        getCpuInstance()?.resize();
      });
      cpuResizeObserver.observe(cpuChartEl.value);
    }

    if (memChartEl.value) {
      memResizeObserver = new ResizeObserver(() => {
        getMemInstance()?.resize();
      });
      memResizeObserver.observe(memChartEl.value);
    }

    // 每30秒刷新一次数据
    const interval = setInterval(loadData, 30000);

    onUnmounted(() => {
      clearInterval(interval);
    });
  }, 500);

  // 监听窗口大小变化
  window.addEventListener("resize", handleResize);

  onUnmounted(() => {
    window.removeEventListener("resize", handleResize);
    if (cpuResizeObserver) {
      cpuResizeObserver.disconnect();
    }
    if (memResizeObserver) {
      memResizeObserver.disconnect();
    }
  });
});

watch(
  () => [cpuOption.value, memOption.value, chartHeight.value],
  () => {
    if (performanceData.value) {
      setCpuOption(cpuOption.value);
      setMemOption(memOption.value);
      nextTick(() => {
        getCpuInstance()?.resize();
        getMemInstance()?.resize();
      });
    }
  },
  { deep: true }
);
</script>

<template>
  <div>
    <v-card-title
      class="text-h6 font-weight-bold"
      :class="chartLayout === 'landscape' ? 'pa-3 pb-1' : 'pa-5'"
    >
      {{ $t("dashboardPage.systemPerformance") }}
    </v-card-title>
    <v-card-text :class="chartLayout === 'landscape' ? 'pa-3 pt-0' : ''">
      <div
        v-if="loading"
        class="h-full d-flex align-center justify-center"
        :style="{ minHeight: chartHeight }"
      >
        <v-progress-circular
          indeterminate
          color="primary"
        ></v-progress-circular>
      </div>
      <v-row v-else dense class="performance-charts-row">
        <v-col :cols="chartCols" md="6">
          <div
            ref="cpuChartEl"
            class="performance-chart"
            :style="{ height: chartHeight }"
          ></div>
        </v-col>
        <v-col :cols="chartCols" md="6">
          <div
            ref="memChartEl"
            class="performance-chart"
            :style="{ height: chartHeight }"
          ></div>
        </v-col>
      </v-row>
    </v-card-text>
  </div>
</template>

<style lang="scss" scoped>
.performance-chart {
  width: 100%;
  min-height: 160px;
}

.performance-charts-row {
  margin-bottom: 0;
}
</style>
