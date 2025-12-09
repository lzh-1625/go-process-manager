<script setup lang="ts">
import { ProcessItem } from "~/src/types/process/process";
import { init } from "echarts";
import TerminalPty from "./TerminalPty.vue";
import TerminalStd from "./TerminalStd.vue";
import {
  deleteProcessConfig,
  getContorl,
  killProcess,
  startProcess,
  createProcessShare,
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
        name: "内存(" + (mem / 1024).toFixed(2) + "MB)",
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
        data: props.data.usage.mem.map((num) =>
          parseFloat((num / 1024).toFixed(2))
        ),
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

const handleStart = () => {
  startProcess(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("success");
    }
  });
};

const handleStop = () => {
  killProcess(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("success");
    }
  });
};

const handleResize = () => {
  chartInstance.resize();
};

onMounted(() => {
  initEChart();
  window.addEventListener("resize", handleResize);
});
onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
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

// 分享链接相关
const shareDialog = ref(false);
const shareForm = ref({
  minutes: 60,
  write: false,
});
const shareToken = ref("");
const shareUrl = ref("");

const openShareDialog = () => {
  shareDialog.value = true;
  shareToken.value = "";
  shareUrl.value = "";
  shareForm.value = {
    minutes: 60,
    write: false,
  };
};

const generateShareLink = () => {
  createProcessShare({
    pid: props.data.uuid,
    minutes: shareForm.value.minutes,
    write: shareForm.value.write,
  }).then((e) => {
    if (e.code === 0 && e.data) {
      shareToken.value = e.data.token;
      shareUrl.value = `${window.location.origin}/share?token=${e.data.token}`;
      snackbarStore.showSuccessMessage("分享链接创建成功");
    }
  });
};

const copyShareLink = () => {
  navigator.clipboard.writeText(shareUrl.value).then(() => {
    snackbarStore.showSuccessMessage("链接已复制到剪贴板");
  });
};

const copyToken = () => {
  navigator.clipboard.writeText(shareToken.value).then(() => {
    snackbarStore.showSuccessMessage("Token已复制");
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
      <div class="top-right" v-permission="1">
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
            <v-list-item v-if="data.termType == 'pty'" @click="openShareDialog"> 创建分享链接 </v-list-item>
          </v-list>
        </v-menu>
      </div>
    </div>

    <!-- 中间：ECharts -->
    <div :id="'echarts' + props.data.uuid" class="chart"></div>

    <!-- 底部：按钮组 + 时间 -->
    <div class="footer">
      <div class="bottom-left">
        <v-chip size="small" variant="outlined" class="d-flex align-center">
          <!-- 终端按钮 -->
          <v-btn
            @click="terminalComponent?.wsConnect()"
            size="small"
            icon="mdi-console"
            variant="text"
            density="comfortable"
          />
          <!-- 启动按钮 -->
          <v-btn
            @click="handleStart"
            size="small"
            icon="mdi-play"
            variant="text"
            density="comfortable"
          />
          <!-- 停止按钮 -->
          <v-btn
            @click="handleStop"
            size="small"
            icon="mdi-stop"
            variant="text"
            density="comfortable"
          />
          <!-- 编辑按钮 -->
          <v-btn
            v-permission="1"
            @click="processConfigComponent?.openConfigDialog()"
            size="small"
            icon="mdi-pencil"
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
    <TerminalStd
      v-else-if="props.data.termType == 'std'"
      :data="props.data"
      ref="terminalComponent"
    ></TerminalStd>
    <TerminalPty
      v-else
      :data="props.data"
      ref="terminalComponent"
    ></TerminalPty>
    <ProcessConfig
      :data="props.data"
      ref="processConfigComponent"
    ></ProcessConfig>

    <!-- 分享链接对话框 -->
    <v-dialog v-model="shareDialog" max-width="600">
      <v-card class="rounded-xl">
        <v-card-title class="text-h6 font-weight-medium">
          <v-icon color="primary" class="mr-2">mdi-share-variant</v-icon>
          <span>创建分享链接</span>
        </v-card-title>

        <v-divider></v-divider>

        <v-card-text class="pt-6">
          <v-container fluid v-if="!shareToken">
            <v-row dense>
              <v-col cols="12">
                <v-text-field
                  v-model.number="shareForm.minutes"
                  type="number"
                  label="有效时长（分钟）"
                  variant="outlined"
                  density="comfortable"
                  hint="链接的有效期，单位为分钟"
                  persistent-hint
                  :rules="[(v) => v > 0 || '时长必须大于0']"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <div class="d-flex align-center justify-space-between">
                  <div>
                    <div class="font-weight-medium">允许写入权限</div>
                    <div class="text-caption text-grey">
                      是否允许通过分享链接进行终端输入操作
                    </div>
                  </div>
                  <v-switch
                    v-model="shareForm.write"
                    color="primary"
                    inset
                    hide-details
                  ></v-switch>
                </div>
              </v-col>
            </v-row>
          </v-container>

          <v-container fluid v-else>
            <v-row dense>
              <v-col cols="12">
                <v-alert type="success" variant="tonal" class="mb-4">
                  <div class="text-subtitle-2 mb-2">分享链接已生成</div>
                  <div class="text-caption">
                    有效期：{{ shareForm.minutes }} 分钟
                    <span v-if="shareForm.write" class="ml-2">• 可写入</span>
                    <span v-else class="ml-2">• 只读</span>
                  </div>
                </v-alert>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  :model-value="shareUrl"
                  label="分享链接"
                  variant="outlined"
                  density="comfortable"
                  readonly
                  append-inner-icon="mdi-content-copy"
                  @click:append-inner="copyShareLink"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  :model-value="shareToken"
                  label="Token"
                  variant="outlined"
                  density="comfortable"
                  readonly
                  append-inner-icon="mdi-content-copy"
                  @click:append-inner="copyToken"
                ></v-text-field>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions class="justify-end pa-4">
          <v-btn text @click="shareDialog = false">
            {{ shareToken ? "关闭" : "取消" }}
          </v-btn>
          <v-btn
            v-if="!shareToken"
            color="primary"
            @click="generateShareLink"
            :disabled="shareForm.minutes <= 0"
          >
            生成链接
          </v-btn>
          <v-btn v-else color="primary" @click="copyShareLink">
            <v-icon left>mdi-content-copy</v-icon>
            复制链接
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>
.chart-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 250px; /* 可根据实际容器调整 */
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
