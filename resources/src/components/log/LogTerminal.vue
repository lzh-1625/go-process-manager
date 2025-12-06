<template>
  <v-dialog
    fullscreen
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
        <!-- 进程名多选 -->
        <div class="process-select-container mr-3">
          <v-autocomplete
            v-model="selectedProcesses"
            :items="processList"
            label="选择进程"
            multiple
            chips
            closable-chips
            density="compact"
            variant="outlined"
            hide-details
            clearable
            style="min-width: 300px; max-width: 400px;"
            @update:model-value="onProcessChange"
          >
            <template #chip="{ props: chipProps, item }">
              <v-chip
                v-bind="chipProps"
                size="small"
                color="primary"
              >
                {{ item.title }}
              </v-chip>
            </template>
          </v-autocomplete>
        </div>
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
          <v-btn icon @click="loadOlderLogs">
            <v-icon>mdi-arrow-up-bold</v-icon>
          </v-btn>
          <v-btn icon @click="loadNewerLogs">
            <v-icon>mdi-arrow-down-bold</v-icon>
          </v-btn>
          <v-btn icon @click="closeTerminal">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <!-- 终端容器 -->
      <div class="terminal-wrapper">
        <!-- 顶部加载按钮 -->
        <v-btn
          v-show="showLoadOlderBtn"
          class="load-older-btn"
          color="primary"
          size="small"
          :loading="loading"
          @click="loadOlderLogs"
        >
          <v-icon start>mdi-arrow-up</v-icon>
          加载更早的日志
        </v-btn>

        <div ref="xtermEl" class="terminal-container"></div>

        <!-- 底部加载按钮 -->
        <v-btn
          v-show="showLoadNewerBtn"
          class="load-newer-btn"
          color="primary"
          size="small"
          :loading="loading"
          @click="loadNewerLogs"
        >
          <v-icon start>mdi-arrow-down</v-icon>
          加载更新的日志
        </v-btn>
      </div>

      <div class="terminal-status">
        <v-chip size="x-small" variant="text">
          滚动到顶部或底部显示加载按钮
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
  processList?: string[]; // 进程列表
  startTime?: number; // 开始时间戳（毫秒）
}

const props = defineProps<Props>();

const dialog = ref(false);
const xtermEl = ref<HTMLElement | null>(null);
const cardRef = ref<HTMLElement | null>(null);
const loading = ref(false);
const loadedLogsCount = ref(0);

// 进程选择
const selectedProcesses = ref<string[]>([]);

// 上下文模式
const contextTime = ref<number | null>(null);
const contextProcessName = ref<string | null>(null);

// 按钮显示控制
const showLoadOlderBtn = ref(false);
const showLoadNewerBtn = ref(false);

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
    contextTime.value = null;
    contextProcessName.value = null;
    dialog.value = true;
  },
  openWithContext: (time: number, processName: string) => {
    contextTime.value = time;
    contextProcessName.value = processName;
    selectedProcesses.value = [processName];
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
          // 延迟加载日志，确保终端已完全适配
          setTimeout(() => {
            loadInitialLogs();
          }, 50);
        }, 150);
      }, 100);
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

};

// 适应终端大小
const fitTerminal = () => {
  if (fitAddon && term && xtermEl.value) {
    try {
      // 确保容器有尺寸
      const container = xtermEl.value;
      if (container.offsetWidth === 0 || container.offsetHeight === 0) {
        console.warn("终端容器尺寸为0，延迟适配");
        setTimeout(fitTerminal, 100);
        return;
      }

      fitAddon.fit();
      term.refresh(0, term.rows - 1);

      // 强制刷新滚动条
      term.scrollToBottom();
      term.scrollLines(-1);
      term.scrollLines(1);
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
  if (!term) return;

  const buffer = term.buffer.active;
  const viewportY = buffer.viewportY;
  const baseY = buffer.baseY;

  // 滚动到顶部，显示加载更早日志的按钮
  showLoadOlderBtn.value = viewportY <= 5 && oldestTime !== null;

  // 滚动到底部，显示加载最新日志的按钮
  showLoadNewerBtn.value = viewportY >= baseY - 5 && newestTime !== null;
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

  // 使用选中的进程名（多选）
  if (selectedProcesses.value && selectedProcesses.value.length > 0) {
    // 如果只选了一个进程，使用单个值
    if (selectedProcesses.value.length === 1) {
      match.name = selectedProcesses.value[0];
    } else {
      // 多个进程，使用数组（需要后端支持）
      match.name = selectedProcesses.value;
    }
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
    // 上下文模式：加载指定时间点前后各100条日志
    if (contextTime.value) {
      await loadContextLogs();
      return;
    }

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

        // 加载完成后重新适配终端，确保滚动条正确
        nextTick(() => {
          setTimeout(() => {
            fitTerminal();
          }, 100);
        });
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

// 加载上下文日志（前后各100条）
const loadContextLogs = async () => {
  if (!term || !contextTime.value) return;

  try {
    // 加载之前的100条日志
    const beforeQuery: GetLogReq = {
      page: { from: 0, size: 5 },
      time: { endTime: contextTime.value },
      sort: "desc",
    };

    if (contextProcessName.value) {
      beforeQuery.match = { name: contextProcessName.value };
    }

    const beforeResponse = await getLog(beforeQuery);
    const beforeLogs = beforeResponse.code === 0 && beforeResponse.data?.data
      ? beforeResponse.data.data.reverse()
      : [];

    // 加载之后的100条日志
    const afterQuery: GetLogReq = {
      page: { from: 0, size: 5 },
      time: { startTime: contextTime.value },
      sort: "asc",
    };

    if (contextProcessName.value) {
      afterQuery.match = { name: contextProcessName.value };
    }

    const afterResponse = await getLog(afterQuery);
    const afterLogs = afterResponse.code === 0 && afterResponse.data?.data
      ? afterResponse.data.data
      : [];

    // 合并日志
    const allLogs = [...beforeLogs, ...afterLogs];
    logCache.value = allLogs;

    if (allLogs.length > 0) {
      // 计算目标日志之前的行数
      let linesBeforeTarget = 0;
      let targetFound = false;

      // 写入日志并计算目标位置
      allLogs.forEach((log) => {
        const isTargetLog = log.time === contextTime.value;

        if (!targetFound && !isTargetLog) {
          // 计算这条日志占用的行数
          const logLines = (log.log.match(/\n/g) || []).length + 1;
          linesBeforeTarget += logLines;
        }

        if (isTargetLog) {
          targetFound = true;
          linesBeforeTarget += 1; // 标记行
          writeLogToTerminal(log);;
        } else {
          writeLogToTerminal(log);
        }
      });

      oldestTime = Math.min(...allLogs.map((l) => l.time));
      newestTime = Math.max(...allLogs.map((l) => l.time));
      loadedLogsCount.value = allLogs.length;

      term.writeln("");
      term.writeln(`\x1b[32m加载完成: 共 ${allLogs.length} 条日志\x1b[0m`);
      term.writeln(`\x1b[32m前: ${beforeLogs.length} 条，后: ${afterLogs.length} 条\x1b[0m`);

      // 加载完成后滚动到目标日志位置
      if (targetFound) {
        nextTick(() => {
          setTimeout(() => {
            fitTerminal();
            setTimeout(() => {
              if (term) {
                // 滚动到目标位置，将目标日志置于屏幕中央偏上的位置
                const viewportRows = term.rows;
                const scrollToLine = linesBeforeTarget - Math.floor(viewportRows / 3);

                // 先滚动到顶部
                term.scrollToTop();
                // 然后滚动到目标位置
                if (scrollToLine > 0) {
                  term.scrollLines(scrollToLine);
                }
              }
            }, 100);
          }, 150);
        });
      }
    } else {
      term.writeln("\x1b[33m未找到上下文日志\x1b[0m");
    }
  } catch (error) {
    console.error("加载上下文日志错误:", error);
    term.writeln("\x1b[31m加载上下文日志失败\x1b[0m");
  }
};

// 加载更早的日志
const loadOlderLogs = async () => {
  if (!term || !oldestTime || isLoadingMore) return;

  isLoadingMore = true;
  loading.value = true;
  showLoadOlderBtn.value = false; // 隐藏按钮

  try {
    const query = buildQuery(oldestTime);
    const response = await getLog(query);

    if (response.code === 0 && response.data?.data) {
      const logs = response.data.data;

      if (logs.length > 0) {
        // 保存当前缓存的日志
        const currentLogs = [...logCache.value];

        // 将新日志添加到缓存前面（倒序，因为查询返回的是desc）
        const newLogs = logs.reverse();
        logCache.value = [...newLogs, ...currentLogs];

        // 清空终端并重新写入所有日志
        term.clear();

        // 写入所有日志
        logCache.value.forEach((log) => {
          term!.write(log.log);
        });

        // 更新最旧时间
        oldestTime = newLogs[0].time;
        loadedLogsCount.value = logCache.value.length;

        // 滚动到之前的位置（大约）
        nextTick(() => {
          setTimeout(() => {
            if (term) {
              // 计算新内容的行数
              const newLineCount = logs.reduce((count, log) => {
                return count + (log.log.match(/\n/g) || []).length + 1;
              }, 0);

              // 滚动到之前查看的内容位置
              term.scrollLines(newLineCount);
            }
          }, 100);
        });

        snackbarStore.showSuccessMessage(`已加载 ${logs.length} 条更早的日志`);
      } else {
        snackbarStore.showInfoMessage("没有更早的日志了");
      }
    }
  } catch (error) {
    console.error("加载更早日志错误:", error);
    snackbarStore.showErrorMessage("加载更早日志失败");
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
  showLoadNewerBtn.value = false; // 隐藏按钮

  try {
    const query = buildQuery(undefined, newestTime);
    const response = await getLog(query);

    if (response.code === 0 && response.data?.data) {
      const logs = response.data.data;

      if (logs.length > 0) {
        // 追加日志到缓存和终端
        logs.forEach((log) => {
          writeLogToTerminal(log);
          logCache.value.push(log);
        });

        newestTime = logs[logs.length - 1].time;
        loadedLogsCount.value = logCache.value.length;

        // 加载完成后重新适配终端，确保滚动条正确
        nextTick(() => {
          setTimeout(() => {
            fitTerminal();
          }, 100);
        });

        snackbarStore.showSuccessMessage(`已加载 ${logs.length} 条新日志`);
      } else {
        snackbarStore.showInfoMessage("没有更新的日志了");
      }
    }
  } catch (error) {
    console.error("加载最新日志错误:", error);
    snackbarStore.showErrorMessage("加载最新日志失败");
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

// 进程选择变化
const onProcessChange = () => {
  // 重新加载日志
  clearTerminal();
};

// 清空终端
const clearTerminal = () => {
  if (term) {
    term.clear();
    loadedLogsCount.value = 0;
    logCache.value = [];
    oldestTime = null;
    newestTime = null;
    showLoadOlderBtn.value = false;
    showLoadNewerBtn.value = false;

    // 显示欢迎信息
    term.writeln("\x1b[1;36m========================================\x1b[0m");
    term.writeln("\x1b[1;32m  重新加载日志...\x1b[0m");
    term.writeln("\x1b[1;36m========================================\x1b[0m");

    if (selectedProcesses.value && selectedProcesses.value.length > 0) {
      term.writeln(`\x1b[33m选择的进程: ${selectedProcesses.value.join(', ')}\x1b[0m`);
    }

    term.writeln("");

    loadInitialLogs();
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
  showLoadOlderBtn.value = false;
  showLoadNewerBtn.value = false;
  selectedProcesses.value = [];
  contextTime.value = null;
  contextProcessName.value = null;
};

onUnmounted(() => {
  cleanup();
});
</script>

<style scoped>
.terminal-card {
  height: 100%;
  width: 100%;
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

.process-select-container {
  display: flex;
  align-items: center;
}

.terminal-wrapper {
  flex: 1 1 auto;
  position: relative;
  overflow: hidden;
  min-height: 0;
}

.terminal-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
  padding: 0;
  position: relative;
}

.load-older-btn {
  position: absolute;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}

.load-newer-btn {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
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
  padding: 8px;
}

:deep(.xterm-viewport) {
  width: 100% !important;
  overflow-y: auto !important;
}

:deep(.xterm-screen) {
  width: 100% !important;
  height: 100% !important;
}

:deep(.xterm canvas) {
  width: 100% !important;
  height: 100% !important;
}

/* 多选框样式调整 */
:deep(.v-autocomplete .v-field) {
  background-color: rgba(255, 255, 255, 0.1);
}

:deep(.v-autocomplete .v-field__input) {
  color: white;
}
</style>

