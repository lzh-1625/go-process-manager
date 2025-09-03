<script setup lang="ts">
import { ProcessItem } from "~/src/types/process/process";
import { init } from "echarts";
const initEChart = () => {
  props.data.usage.cpu = (props.data.usage.cpu ?? [0, 0]).map((num) =>
    parseFloat(num.toFixed(2))
  );
  props.data.usage.mem = (props.data.usage.mem ?? [0, 0]).map((num) =>
    parseFloat(num.toFixed(2))
  );
  const cpu = props.data.usage.cpu[props.data.usage.cpu.length - 1] ?? "-";
  const mem = props.data.usage.mem[props.data.usage.mem.length - 1] ?? "-";
  var myChart = init(document.getElementById("echarts" + props.index));
  var option = {
    tooltip: {
      trigger: "axis",
    },
    legend: {
      data: ["CPU", "内存"],
    },
    animationDuration: 2000,
    grid: {
      left: "3%",
      right: "4%",
      bottom: "3%",
      containLabel: true,
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      show: false,
      data: props.data.usage.time,
    },
    yAxis: [
      {
        type: "value",
        name: " CPU(" + cpu + "%)",
        min: 0, // 设置CPU的y轴最小值为10
        max: props.data.usage.cpuCapacity,
        minInterval: 0.1,
        splitLine: {
          show: false,
        },
        axisLine: { show: false },
        axisTick: { show: false },
      },
      {
        type: "value",
        name: " 内存(" + mem + "MB)",
        max: parseFloat((props.data.usage.memCapacity / 1024).toFixed(2)),
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { show: false },
      },
    ],
    series: [
      {
        name: "CPU",
        type: "line",
        data: props.data.usage.cpu,
        yAxisIndex: 0,
        showSymbol: false,
        lineStyle: {
          color: "#4ee5b9",
        },
        itemStyle: {
          color: "#4ee5b9",
        },
      },
      {
        name: "内存",
        type: "line",
        data: props.data.usage.mem,
        yAxisIndex: 1,
        showSymbol: false,
        lineStyle: {
          color: "#ffe17e",
        },
        itemStyle: {
          color: "#ffe17e",
        },
      },
    ],
  };
  if (props.data.cgroupEnable) {
    if (props.data.cpuLimit) {
      (option.series as any).push({
        name: "CPU限制（%）",
        type: "line",
        yAxisIndex: 0,
        data: new Array(props.data.usage.time!.length).fill(
          props.data.cpuLimit
        ),
        lineStyle: {
          type: "dashed",
          color: "#4ee5b9",
        },
        showSymbol: false,
      });
    }
    if (props.data.memoryLimit) {
      (option.series as any).push({
        name: "内存限制（MB）",
        type: "line",
        yAxisIndex: 1,
        data: new Array(props.data.usage.time!.length).fill(
          props.data.memoryLimit
        ),
        lineStyle: {
          type: "dashed",
          color: "#ffe17e",
        },
        showSymbol: false,
      });
    }
  }
  console.log(option);
  myChart.setOption(option);
};

const buttons = [
  { label: "按钮1", action: () => console.log("按钮1点击") },
  { label: "按钮2", action: () => console.log("按钮2点击") },
  { label: "按钮3", action: () => console.log("按钮3点击") },
];

onMounted(() => {
  initEChart();
});

const props = defineProps<{
  data: ProcessItem;
  index: Number;
}>();
</script>
<template>
  <div class="chart-container">
    <!-- 顶部：进程名字 + 菜单 -->
    <div class="header">
      <div class="top-left">{{ props.data.name }}</div>
      <div class="top-right">
        <button @click="">菜单</button>
      </div>
    </div>

    <!-- 中间：ECharts -->
    <div :id="'echarts' + props.index" class="chart"></div>

    <!-- 底部：按钮组 + 时间 -->
    <div class="footer">
      <div class="bottom-left">
        <button v-for="(btn, idx) in buttons" :key="idx" @click="">
          {{ btn.label }}
        </button>
      </div>
      <div class="bottom-right">{{ props.data.startTime }}</div>
    </div>
  </div>
</template>

<style scoped>
.chart-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 260px; /* 可根据实际容器调整 */
  background: #fff;
  border: 1px solid #ccc;
  border-radius: 8px;
  overflow: hidden;
}

/* 顶部 header */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 10px;
  font-weight: bold;
  height: 30px; /* 顶部固定高度 */
}

/* 中间图表自适应 */
.chart {
  flex: 1; /* 占满剩余空间 */
  width: 100%;
}

/* 底部 footer */
.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 10px;
  height: 40px; /* 底部固定高度 */
}

.bottom-left {
  display: flex;
  gap: 5px;
}

.bottom-right {
  text-align: right;
}
</style>
