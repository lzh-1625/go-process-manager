<template>
  <v-dialog
    fullscreen
    hide-overlay
    transition="dialog-bottom-transition"
    v-model="dialog"
    @update:modelValue="handleDialogChange"
  >
    <v-card class="terminal-card" ref="cardRef">
      <v-toolbar dense color="grey-darken-4" dark class="terminal-toolbar">
        <v-toolbar-title>
          <v-icon start>mdi-console</v-icon>
          终端日志查看器
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <div class="d-flex align-center ga-2 mr-2">
          <v-chip size="small" color="info" variant="flat">
            {{ loadedLogsCount }} 条日志
          </v-chip>
          <v-chip v-if="loading" size="small" color="warning" variant="flat">
            加载中...
          </v-chip>
        </div>
        <v-toolbar-items>
          <v-btn icon @click="clearTerminal">
            <v-icon>mdi-delete-sweep</v-icon>
          </v-btn>
          <v-btn icon @click="scrollToTop">
            <v-icon>mdi-arrow-up-bold</v-icon>
          </v-btn>
          <v-btn icon @click="scrollToBottom">
            <v-icon>mdi-arrow-down-bold</v-icon>
          </v-btn>
          <v-btn icon @click="closeTerminal">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>
      <div ref="xtermEl" class="terminal-container"></div>
      <div class="terminal-status">
        <v-chip size="x-small" variant="text">
          滚动到顶部加载更早的日志 | 滚动到底部加载最新的日志
        </v-chip>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, nextTick, onUnmounted, watch } from "vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "@xterm/addon-canvas";
import "xterm/css/xterm.css";
import { getLog } from "~/src/api/log";
import type { GetLogReq, ProcessLog } from "~/src/types/log/log";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const snackbarStore = useSnackbarStore();

interface Props {
  searchForm?: {
    name?: string;
    log?: string;
    using?: string;
  };
  startTime?: number; // 开始时间戳（毫秒）
}

const props = defineProps<Props>();

const dialog = ref(false);
const xtermEl = ref<HTMLElement | null>(null);
const cardRef = ref<HTMLElement | null>(null);
const loading = ref(false);
const loadedLogsCount = ref(0);

let term: Terminal | null = null;
let fitAddon: FitAddon | null = null;

// 日志缓存
const logCache = ref<ProcessLog[]>([]);
const pageSize = 100;
let oldestTime: number | null = null;
let newestTime: number | null = null;
let isLoadingMore = false;

defineExpose({
  open: () => {
    dialog.value = true;
  },
});

// 监听对话框状态变化
watch(dialog, (newVal) => {
  if (newVal) {
    // 对话框打开时，延迟初始化以确保 DOM 已渲染
    nextTick(() => {
      setTimeout(() => {
        initTerminal();
        // 再次延迟以确保终端完全渲染
        setTimeout(() => {
          fitTerminal();
          loadInitialLogs();
        }, 100);
      }, 50);
    });
  }
});

// 处理对话框变化
const handleDialogChange = (val: boolean) => {
  if (!val) {
    cleanup();
  }
};

const initTerminal = () => {
  if (!xtermEl.value) {
    snackbarStore.showErrorMessage("终端容器初始化失败");
    return;
  }

  // 清理旧的终端实例
  if (term) {
    term.dispose();
  }

  fitAddon = new FitAddon();

  term = new Terminal({
    convertEol: true,
    disableStdin: true,
    cursorBlink: false,
    cursorStyle: "block",
    scrollback: 10000,
    fontSize: 14,
    fontFamily: '"Cascadia Code", "Fira Code", "Consolas", "Monaco", monospace',
    lineHeight: 1.2,
  });

  term.loadAddon(new CanvasAddon());
  term.loadAddon(fitAddon);
  term.open(xtermEl.value);

  // 监听滚动事件
  term.onScroll(() => {
    handleScroll();
  });

  window.addEventListener("resize", handleResize);

  // 显示欢迎信息
  term.writeln("\x1b[1;36m========================================\x1b[0m");
  term.writeln("\x1b[1;32m  欢迎使用终端日志查看器\x1b[0m");
  term.writeln("\x1b[1;36m========================================\x1b[0m");
  if (props.startTime) {
    const startTimeStr = formatTime(props.startTime);
    term.writeln(
      `\x1b[33m查询模式: 从 ${startTimeStr} 开始的日志\x1b[0m`
    );
  } else {
    term.writeln("\x1b[33m查询模式: 最新日志\x1b[0m");
  }
  term.writeln("\x1b[33m提示: 滚动到顶部加载更早的日志\x1b[0m");
  term.writeln("\x1b[33m      滚动到底部加载最新的日志\x1b[0m");
  term.writeln("\x1b[1;36m========================================\x1b[0m");
  term.writeln("");
};

// 适应终端大小
const fitTerminal = () => {
  if (fitAddon && term) {
    try {
      fitAddon.fit();
      term.refresh(0, term.rows - 1);
    } catch (error) {
      console.error("终端适应大小失败:", error);
    }
  }
};

const handleResize = () => {
  // 延迟调用以确保窗口调整完成
  setTimeout(() => {
    fitTerminal();
  }, 100);
};

const handleScroll = () => {
  if (!term || isLoadingMore) return;

  const buffer = term.buffer.active;
  const viewportY = buffer.viewportY;
  const baseY = buffer.baseY;

  // 滚动到顶部，加载更早的日志
  if (viewportY <= 5 && oldestTime) {
    loadOlderLogs();
  }
  // 滚动到底部，加载最新的日志
  else if (viewportY >= baseY - 5 && newestTime) {
    loadNewerLogs();
  }
};

// 构建查询参数
const buildQuery = (beforeTime?: number, afterTime?: number): GetLogReq => {
  const query: GetLogReq = {
    page: {
      from: 0,
      size: pageSize,
    },
  };

  // 添加搜索条件
  const match: any = {};
  if (props.searchForm?.name) {
    match.name = props.searchForm.name;
  }
  if (props.searchForm?.log) {
    match.log = props.searchForm.log;
  }
  if (props.searchForm?.using) {
    match.using = props.searchForm.using;
  }
  if (Object.keys(match).length > 0) {
    query.match = match;
  }

  // 添加时间范围
  query.time = {};
  if (beforeTime) {
    // 加载更早的日志（向上滚动）
    query.time.endTime = beforeTime;
    query.sort = "desc"; // 按时间倒序
  } else if (afterTime) {
    // 加载更新的日志（向下滚动）
    query.time.startTime = afterTime;
    query.sort = "asc"; // 按时间正序
  } else {
    // 初始加载
    if (props.startTime) {
      // 如果指定了开始时间，从该时间开始往下查询
      query.time.startTime = props.startTime;
      query.sort = "asc"; // 按时间正序
    } else {
      // 没有指定开始时间，按时间倒序加载最新的日志
      query.sort = "desc";
    }
  }

  return query;
};

// 加载初始日志
const loadInitialLogs = async () => {
  if (!term) return;

  loading.value = true;
  try {
    const query = buildQuery();
    const response = await getLog(query);

    if (response.code === 0 && response.data?.data) {
      const logs = response.data.data;
      logCache.value = logs;

      if (logs.length > 0) {
        // 根据是否指定开始时间决定显示顺序
        if (props.startTime) {
          // 指定了开始时间，按时间正序显示（从旧到新）
          logs.forEach((log) => {
            writeLogToTerminal(log);
          });
        } else {
          // 未指定开始时间，倒序显示（最新的在下面）
          logs.reverse().forEach((log) => {
            writeLogToTerminal(log);
          });
        }

        oldestTime = Math.min(...logs.map((l) => l.time));
        newestTime = Math.max(...logs.map((l) => l.time));
        loadedLogsCount.value = logs.length;
      } else {
        term.writeln("\x1b[33m未找到日志记录\x1b[0m");
      }
    }
  } catch (error) {
    console.error("加载日志错误:", error);
    term.writeln("\x1b[31m加载日志失败\x1b[0m");
  } finally {
    loading.value = false;
  }
};

// 加载更早的日志
const loadOlderLogs = async () => {
  if (!term || !oldestTime || isLoadingMore) return;

  isLoadingMore = true;
  loading.value = true;

  try {
    const query = buildQuery(oldestTime);
    const response = await getLog(query);

    if (response.code === 0 && response.data?.data) {
      const logs = response.data.data;

      if (logs.length > 0) {
        // 保存当前滚动位置
        const buffer = term.buffer.active;
        const oldBaseY = buffer.baseY;

        // 在顶部插入日志
        const tempLines: string[] = [];
        logs.reverse().forEach((log) => {
          tempLines.push(log.log);
        });

        // 将新日志写入缓冲区顶部
        term.write("\x1b[1L"); // 插入一行
        term.write("\x1b[H"); // 移动到顶部
        tempLines.forEach((line) => {
          term!.write(line);
        });

        oldestTime = logs[0].time;
        loadedLogsCount.value += logs.length;

        // 调整滚动位置
        setTimeout(() => {
          const buffer = term!.buffer.active;
          const newBaseY = buffer.baseY;
          term!.scrollLines(newBaseY - oldBaseY);
        }, 100);
      }
    }
  } catch (error) {
    console.error("加载更早日志错误:", error);
  } finally {
    loading.value = false;
    isLoadingMore = false;
  }
};

// 加载更新的日志
const loadNewerLogs = async () => {
  if (!term || !newestTime || isLoadingMore) return;

  isLoadingMore = true;
  loading.value = true;

  try {
    const query = buildQuery(undefined, newestTime);
    const response = await getLog(query);

    if (response.code === 0 && response.data?.data) {
      const logs = response.data.data;

      if (logs.length > 0) {
        logs.forEach((log) => {
          writeLogToTerminal(log);
        });

        newestTime = logs[logs.length - 1].time;
        loadedLogsCount.value += logs.length;
      }
    }
  } catch (error) {
    console.error("加载最新日志错误:", error);
  } finally {
    loading.value = false;
    isLoadingMore = false;
  }
};

// 写入日志到终端
const writeLogToTerminal = (log: ProcessLog) => {
  if (!term) return;
  // const line = formatLogLine(log);
  term.write(log.log);
};

// 格式化日志行
// const formatLogLine = (log: ProcessLog): string => {
//   const time = formatTime(log.time);
//   const name = log.name || "-";
//   const using = log.using || "-";
//   const logContent = log.log || "";

//   // 使用 ANSI 颜色代码格式化输出
//   return `\x1b[90m[${time}]\x1b[0m \x1b[36m[${name}]\x1b[0m \x1b[35m[${using}]\x1b[0m ${logContent}`;
// };

// 格式化时间
const formatTime = (timestamp: number): string => {
  if (!timestamp) return "-";
  const date = new Date(timestamp);
  return date.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

// 清空终端
const clearTerminal = () => {
  if (term) {
    term.clear();
    loadedLogsCount.value = 0;
    logCache.value = [];
    oldestTime = null;
    newestTime = null;

    // 显示欢迎信息
    term.writeln("\x1b[1;36m========================================\x1b[0m");
    term.writeln("\x1b[1;32m  重新加载日志...\x1b[0m");
    term.writeln("\x1b[1;36m========================================\x1b[0m");
    term.writeln("");

    loadInitialLogs();
  }
};

// 滚动到顶部
const scrollToTop = () => {
  if (term) {
    term.scrollToTop();
  }
};

// 滚动到底部
const scrollToBottom = () => {
  if (term) {
    term.scrollToBottom();
  }
};

// 关闭终端
const closeTerminal = () => {
  dialog.value = false;
  cleanup();
};

const cleanup = () => {
  window.removeEventListener("resize", handleResize);
  if (term) {
    term.dispose();
    term = null;
  }
  logCache.value = [];
  oldestTime = null;
  newestTime = null;
  loadedLogsCount.value = 0;
};

onUnmounted(() => {
  cleanup();
});
</script>

<style scoped>
.terminal-card {
  height: 100vh;
  width: 100vw;
  background-color: #1e1e1e;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.terminal-toolbar {
  height: 48px;
  flex-shrink: 0;
  flex-grow: 0;
}

.terminal-container {
  flex: 1 1 auto;
  overflow: hidden;
  padding: 8px;
  min-height: 0;
  position: relative;
}

.terminal-status {
  background: rgba(0, 0, 0, 0.3);
  padding: 4px 8px;
  text-align: center;
  flex-shrink: 0;
  flex-grow: 0;
  height: 32px;
}

:deep(.terminal) {
  width: 100% !important;
  height: 100% !important;
}

:deep(.xterm) {
  width: 100% !important;
  height: 100% !important;
}

:deep(.xterm-viewport) {
  width: 100% !important;
}
</style>

