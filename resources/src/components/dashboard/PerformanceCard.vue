<!--
* @Component: PerformanceCard
* @Description: CPU和内存使用率饼图
-->
<script setup lang="ts">
import { ref, onMounted, computed, Ref } from "vue";
import type { EChartsOption } from "echarts";
import { useChart, RenderType, ThemeType } from "@/plugins/echarts";
import { useTheme } from "vuetify";
import { getPerformceUsage, PerformceUsage } from "@/api/metric";

const { current } = useTheme();
const loading = ref(true);
const performanceData = ref<PerformceUsage | null>(null);

// 预定义的颜色集合
const colors = [
  "#f44336",
  "#e91e63",
  "#9c27b0",
  "#673ab7",
  "#3f51b5",
  "#2196f3",
  "#03a9f4",
  "#00bcd4",
  "#009688",
  "#4caf50",
  "#8bc34a",
  "#cddc39",
  "#ffeb3b",
  "#ffc107",
  "#ff9800",
  "#ff5722",
  "#795548",
  "#607d8b",
];

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
    ...sortedItems.map((item, index) => ({
      value: parseFloat(item.cpu.toFixed(2)),
      name: item.name,
      itemStyle: { color: colors[index % colors.length] },
    })),
  ];

  // 如果有其他进程，添加"其他"项
  if (otherProcessesCpu > 0.01) {
    data.push({
      value: parseFloat(otherProcessesCpu.toFixed(2)),
      name: "其他进程",
      itemStyle: { color: "#9e9e9e" },
    });
  }

  // 添加空闲项
  data.push({
    value: parseFloat(performanceData.value.cpuFree.toFixed(2)),
    name: "空闲",
    itemStyle: { color: "#4caf50" },
  });

  return {
    backgroundColor: current.value.colors.surface,
    tooltip: {
      trigger: "item",
      formatter: "{a} <br/>{b}: {c}% ({d}%)",
    },
    title: {
      text: "CPU使用率",
      left: "center",
      top: 10,
      textStyle: {
        color: current.value.colors.onSurface,
        fontSize: 16,
        fontWeight: "bold",
      },
    },
    legend: {
      orient: "vertical",
      left: "left",
      top: "middle",
      textStyle: {
        color: current.value.colors.onSurface,
        fontSize: 11,
      },
      type: "scroll",
      pageIconColor: current.value.colors.primary,
      pageIconInactiveColor: current.value.colors.onSurface,
      pageTextStyle: {
        color: current.value.colors.onSurface,
      },
    },
    series: [
      {
        name: "CPU",
        type: "pie",
        radius: ["40%", "70%"],
        center: ["60%", "55%"],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: current.value.colors.surface,
          borderWidth: 2,
        },
        label: {
          show: true,
          position: "outside",
          formatter: (params: any) => {
            if (params.percent < 3) return "";
            return `${params.name}: ${params.value}%`;
          },
          color: current.value.colors.onSurface,
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
          show: true,
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
      itemStyle: { color: colors[index % colors.length] },
    })),
    // 添加空闲项
    {
      value: parseFloat((performanceData.value.memFree / 1024).toFixed(2)),
      name: "空闲",
      itemStyle: { color: "#2196f3" },
    },
  ];

  return {
    backgroundColor: current.value.colors.surface,
    tooltip: {
      trigger: "item",
      formatter: "{a} <br/>{b}: {c}MB ({d}%)",
    },
    title: {
      text: "内存使用率",
      left: "center",
      top: 10,
      textStyle: {
        color: current.value.colors.onSurface,
        fontSize: 16,
        fontWeight: "bold",
      },
    },
    legend: {
      orient: "vertical",
      left: "left",
      top: "middle",
      textStyle: {
        color: current.value.colors.onSurface,
        fontSize: 11,
      },
      type: "scroll",
      pageIconColor: current.value.colors.primary,
      pageIconInactiveColor: current.value.colors.onSurface,
      pageTextStyle: {
        color: current.value.colors.onSurface,
      },
    },
    series: [
      {
        name: "内存",
        type: "pie",
        radius: ["40%", "70%"],
        center: ["60%", "55%"],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: current.value.colors.surface,
          borderWidth: 2,
        },
        label: {
          show: true,
          position: "outside",
          formatter: (params: any) => {
            if (params.percent < 3) return "";
            return `${params.name}: ${params.value}MB`;
          },
          color: current.value.colors.onSurface,
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
          show: true,
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
  getCpuInstance()?.resize();
  getMemInstance()?.resize();
};

onMounted(() => {
  setTimeout(async () => {
    loading.value = false;
    await nextTick();
    // 等待 DOM 渲染后再加载数据
    await loadData();

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
  });
});

watch(
  () => [cpuOption.value, memOption.value],
  () => {
    if (performanceData.value) {
      setCpuOption(cpuOption.value);
      setMemOption(memOption.value);
    }
  },
  { deep: true }
);
</script>

<template>
  <div>
    <v-card-title class="text-h6 font-weight-bold pa-5">
      系统性能监控
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
      <v-row v-else>
        <v-col cols="12" md="6">
          <div ref="cpuChartEl" style="width: 100%; height: 300px"></div>
        </v-col>
        <v-col cols="12" md="6">
          <div ref="memChartEl" style="width: 100%; height: 300px"></div>
        </v-col>
      </v-row>
    </v-card-text>
  </div>
</template>

<style lang="scss" scoped></style>
