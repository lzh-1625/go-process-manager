<script setup lang="ts">
import { ProcessItem } from "~/src/types/process/process";
import { init } from "echarts";
import TerminalPty from "./TerminalPty.vue";
import {
  deleteProcessConfig,
  getContorl,
  killProcess,
  startProcess,
} from "~/src/api/process";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import ProcessConfig from "./ProcessConfig.vue";
let chartInstance;

const snackbarStore = useSnackbarStore();
const initEChart = () => {
  props.data.usage.cpu = (props.data.usage.cpu ?? [0, 0]).map((num) =>
    parseFloat(num.toFixed(2))
  );
  props.data.usage.mem = (props.data.usage.mem ?? [0, 0]).map((num) =>
    parseFloat(num.toFixed(2))
  );
  const cpu = props.data.usage.cpu[props.data.usage.cpu.length - 1] ?? "-";
  const mem = props.data.usage.mem[props.data.usage.mem.length - 1] ?? "-";
  var myChart = init(document.getElementById("echarts" + props.data.uuid));
  var option = {
    tooltip: {
      trigger: "axis",
    },
    legend: {
      data: ["CPU", "内存"],
    },
    animationDuration: 2000,
    grid: {
      left: "0%",
      right: "0%", // 原来 4%，改成更大
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
        name: "CPU(" + cpu + "%)",
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
        name: "内存(" + mem + "MB)",
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
        data: new Array(props.data.usage.time?.length ?? 0).fill(
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
        data: new Array(props.data.usage.time?.length ?? 0).fill(
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
  myChart.setOption(option);
  chartInstance = myChart;
};

type WsHandle = {
  wsConnect: () => void;
};
const terminalComponent = ref<WsHandle | null>(null);

type ConfigHandle = {
  openConfigDialog: () => void;
};
const processConfigComponent = ref<ConfigHandle | null>(null);

const buttons = [
  {
    icon: "mdi-console",
    action: () => {
      terminalComponent.value?.wsConnect();
    },
  },
  {
    icon: "mdi-play",
    action: () => {
      startProcess(props.data.uuid).then((e) => {
        if (e.code === 0) {
          snackbarStore.showSuccessMessage("success");
        }
      });
    },
  },
  {
    icon: "mdi-stop",
    action: () => {
      killProcess(props.data.uuid).then((e) => {
        if (e.code === 0) {
          snackbarStore.showSuccessMessage("success");
        }
      });
    },
  },
  {
    icon: "mdi-pencil",
    action: () => {
      processConfigComponent.value?.openConfigDialog();
    },
  },
];

const handleResize = () => {
  chartInstance.resize();
};

onMounted(() => {
  initEChart();
  window.addEventListener("resize", handleResize);
});

const props = defineProps<{
  data: ProcessItem;
}>();

const control = () => {
  getContorl(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("sucess");
    }
  });
};

const del = () => {
  deleteProcessConfig(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("sucess");
    }
  });
};
</script>
<template>
  <div class="chart-container">
    <!-- 顶部：进程名字 + 菜单 -->
    <div class="header">
      <div class="top-left">
        <v-icon
          color="green"
          v-if="props.data.state.state == 3 || props.data.state.state == 1"
          x-large
          style="float: left"
        >
          mdi-checkbox-marked-circle</v-icon
        >
        <v-icon
          color="red"
          v-if="props.data.state.state == 0"
          x-large
          style="float: left"
        >
          mdi-stop-circle</v-icon
        >
        <div v-if="props.data.state.state == 2" style="float: left">
          <v-tooltip top color="warning">
            <template>
              <v-icon color="yellow accent-4" x-large> mdi-alert-circle</v-icon>
            </template>
            <span>{{ props.data.state.info }}</span>
          </v-tooltip>
        </div>
        {{ props.data.name }}
      </div>
      <div class="top-right">
        <v-menu bottom left>
          <template v-slot:activator="{ props }">
            <v-btn
              variant="text"
              @click=""
              density="compact"
              class="px-1 min-w-0"
              v-bind="props"
            >
              <v-icon>mdi-dots-vertical</v-icon>
            </v-btn>
          </template>

          <v-list nav dense>
            <v-list-item @click="control"> 获取控制权 </v-list-item>
            <v-list-item @click="del"> 删除进程 </v-list-item>
            <v-list-item @click=""> 创建分享链接 </v-list-item>
          </v-list>
        </v-menu>
      </div>
    </div>

    <!-- 中间：ECharts -->
    <div :id="'echarts' + props.data.uuid" class="chart"></div>

    <!-- 底部：按钮组 + 时间 -->
    <div class="footer">
      <div class="bottom-left">
        <v-chip
          size="small"
          variant="outlined"
          style="border-color: grey; color: black"
          class="d-flex align-center"
        >
          <v-btn
            v-for="(btn, idx) in buttons"
            :key="idx"
            @click="btn.action"
            size="small"
            :icon="btn.icon"
            variant="text"
            density="comfortable"
          />
        </v-chip>
      </div>

      <div class="bottom-right text-caption">
        <span>{{ props.data.startTime }}</span>
      </div>
    </div>

    <TerminalPty
      v-if="props.data.termType == 'pty'"
      :data="props.data"
      ref="terminalComponent"
    ></TerminalPty>
    <TerminalPty v-else :data="props.data"></TerminalPty>
    <ProcessConfig
      :data="props.data"
      ref="processConfigComponent"
    ></ProcessConfig>
  </div>
</template>

<style scoped>
.chart-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 250px; /* 可根据实际容器调整 */
  background: #fff;
  overflow: hidden;
}

/* 顶部 header */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 5px;
  font-weight: bold;
  height: 30px; /* 顶部固定高度 */
}

/* 中间图表自适应 */
.chart {
  flex: 1; /* 占满剩余空间 */
  width: 90%;
  margin-left: 5%;
  margin-right: 5%;
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
