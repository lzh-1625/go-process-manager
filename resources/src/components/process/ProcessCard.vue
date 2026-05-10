<script setup lang="ts">
import { ProcessItem } from "~/src/types/process/process";
import { useI18n } from "vue-i18n";
import echarts from "@/plugins/echarts";
import TerminalPty from "./TerminalPty.vue";
import {
  deleteProcessConfig,
  getContorl,
  killProcess,
  startProcess,
  createProcessShare,
} from "~/src/api/process";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import ProcessConfig from "./ProcessConfig.vue";

const { t } = useI18n();
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
  var myChart = echarts.init(document.getElementById("echarts" + props.data.uuid));
  var option = {
    tooltip: {
      trigger: "axis",
    },
    legend: {
      data: ["CPU", "Memery"],
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
        name: "Memery(" + (mem / 1024).toFixed(2) + "MB)",
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
        name: "Memery",
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
        name: t("processCardPage.cpuLimit"),
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
        name: t("processCardPage.memoryLimit"),
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
      snackbarStore.showSuccessMessage(t("processCardPage.startSuccess"));
    }
  });
};

const handleStop = () => {
  killProcess(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage(t("processCardPage.stopSuccess"));
    }
  });
};

const handleResize = () => {
  chartInstance?.resize();
};

// 监听图表容器大小变化
let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
  initEChart();

  // 监听窗口大小变化
  window.addEventListener("resize", handleResize);

  // 监听图表容器大小变化
  const chartEl = document.getElementById("echarts" + props.data.uuid);
  if (chartEl) {
    resizeObserver = new ResizeObserver(() => {
      handleResize();
    });
    resizeObserver.observe(chartEl);
  }
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
  if (resizeObserver) {
    resizeObserver.disconnect();
  }
  // 销毁图表实例
  chartInstance?.dispose();
});

const props = defineProps<{
  data: ProcessItem;
}>();

const control = () => {
  getContorl(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage(t("processCardPage.controlSuccess"));
    }
  });
};

const del = () => {
  deleteProcessConfig(props.data.uuid).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage(t("processCardPage.deleteSuccess"));
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
      snackbarStore.showSuccessMessage(t("processCardPage.createSuccess"));
    }
  });
};

const copyShareLink = () => {
  navigator.clipboard.writeText(shareUrl.value).then(() => {
    snackbarStore.showSuccessMessage(t("processCardPage.copySuccess"));
  });
};

const copyToken = () => {
  navigator.clipboard.writeText(shareToken.value).then(() => {
    snackbarStore.showSuccessMessage(t("processCardPage.tokenCopySuccess"));
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
        <v-tooltip
          v-if="props.data.state.state == 2"
          location="top"
          :text="props.data.state.info"
        >
          <template v-slot:activator="{ props: tooltipProps }">
            <v-icon
              v-bind="tooltipProps"
              color="yellow-accent-4"
              x-large
              style="float: left"
            >
              mdi-alert-circle
            </v-icon>
          </template>
        </v-tooltip>
        <span class="process-name">{{ props.data.name }}</span>
        <span v-if="props.data.user" class="user-info">
          <v-icon size="small" class="mr-1">mdi-account</v-icon>
          {{ props.data.user }}
        </span>
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
            <v-list-item @click="control"> {{ $t("processCardPage.control") }} </v-list-item>
            <v-list-item @click="del"> {{ $t("processCardPage.delete") }} </v-list-item>
            <v-list-item @click="openShareDialog"> {{ $t("processCardPage.createShareLink") }} </v-list-item>
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

    <TerminalPty :data="props.data" ref="terminalComponent"></TerminalPty>
    <ProcessConfig
      :data="props.data"
      ref="processConfigComponent"
    ></ProcessConfig>

    <!-- 分享链接对话框 -->
    <v-dialog v-model="shareDialog" max-width="600">
      <v-card class="rounded-xl">
        <v-card-title class="text-h6 font-weight-medium">
          <v-icon color="primary" class="mr-2">mdi-share-variant</v-icon>
          <span>{{ $t("processCardPage.shareDialogTitle") }}</span>
        </v-card-title>

        <v-divider></v-divider>

        <v-card-text class="pt-6">
          <v-container fluid v-if="!shareToken">
            <v-row dense>
              <v-col cols="12">
                <v-text-field
                  v-model.number="shareForm.minutes"
                  type="number"
                  :label="$t('processCardPage.validDuration')"
                  variant="outlined"
                  density="comfortable"
                  :hint="$t('processCardPage.validDurationHint')"
                  persistent-hint
                  :rules="[(v) => v > 0 || $t('processCardPage.validDurationRule')]"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <div class="d-flex align-center justify-space-between">
                  <div>
                    <div class="font-weight-medium">{{ $t("processCardPage.allowWrite") }}</div>
                    <div class="text-caption text-grey">
                      {{ $t("processCardPage.allowWriteHint") }}
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
                  <div class="text-subtitle-2 mb-2">{{ $t("processCardPage.shareLinkGenerated") }}</div>
                  <div class="text-caption">
                    {{ $t("processCardPage.validity") }}：{{ shareForm.minutes }} {{ $t("processCardPage.minutes") }}
                    <span v-if="shareForm.write" class="ml-2">• {{ $t("processCardPage.writable") }}</span>
                    <span v-else class="ml-2">• {{ $t("processCardPage.readonly") }}</span>
                  </div>
                </v-alert>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  :model-value="shareUrl"
                  :label="$t('processCardPage.shareLinkLabel')"
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
                  :label="$t('processCardPage.tokenLabel')"
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
            {{ shareToken ? $t("processCardPage.close") : $t("processCardPage.cancel") }}
          </v-btn>
          <v-btn
            v-if="!shareToken"
            color="primary"
            @click="generateShareLink"
            :disabled="shareForm.minutes <= 0"
          >
            {{ $t("processCardPage.generateLink") }}
          </v-btn>
          <v-btn v-else color="primary" @click="copyShareLink">
            <v-icon left>mdi-content-copy</v-icon>
            {{ $t("processCardPage.copyLink") }}
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

.top-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.process-name {
  font-weight: bold;
}

.user-info {
  display: inline-flex;
  align-items: center;
  font-size: 0.85em;
  opacity: 0.7;
  font-weight: normal;
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
