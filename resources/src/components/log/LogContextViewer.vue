<template>
  <v-dialog
    fullscreen
    v-model="dialog"
    transition="dialog-bottom-transition"
    @after-leave="cleanup"
  >
    <div class="lcv">
      <!-- 顶部工具栏 -->
      <div class="lcv__toolbar">
        <v-icon size="small" color="primary" class="mr-2">mdi-console</v-icon>
        <span class="lcv__anchor-info">{{ anchorInfo }}</span>
        <v-spacer />
        <v-btn icon size="small" variant="text" color="white" @click="dialog = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </div>

      <!-- 加载更多上文按钮 -->
      <div class="lcv__action-bar lcv__action-bar--top">
        <v-btn
          size="small"
          variant="tonal"
          color="primary"
          :loading="loadingAbove"
          :disabled="loadingAbove"
          prepend-icon="mdi-chevron-double-up"
          @click="loadMoreAbove"
        >
          {{ $t("logPage.loadMoreAbove") }}
        </v-btn>
      </div>

      <!-- 初始加载遮罩 -->
      <div v-if="initialLoading" class="lcv__init-mask">
        <v-progress-circular indeterminate color="primary" size="48" />
      </div>

      <!-- xterm 容器 -->
      <div
        ref="xtermEl"
        class="lcv__terminal"
        :style="{ opacity: initialLoading ? 0 : 1 }"
      ></div>

      <!-- 加载更多下文按钮 -->
      <div class="lcv__action-bar lcv__action-bar--bottom">
        <v-btn
          size="small"
          variant="tonal"
          color="primary"
          :loading="loadingBelow"
          :disabled="loadingBelow"
          prepend-icon="mdi-chevron-double-down"
          @click="loadMoreBelow"
        >
          {{  $t("logPage.loadMoreBelow") }}
        </v-btn>
      </div>
    </div>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, nextTick, onUnmounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "@xterm/addon-canvas";
import "xterm/css/xterm.css";
import { getLog } from "~/src/api/log";
import type { GetLogReq, ProcessLog } from "~/src/types/log/log";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();

// ── DOM & terminal ─────────────────────────────────────────────────────────
const dialog = ref(false);
const xtermEl = ref<HTMLElement | null>(null);
let term: Terminal | null = null;
let fitAddon: FitAddon | null = null;

// ── 状态 ───────────────────────────────────────────────────────────────────
const initialLoading = ref(false);
const loadingAbove = ref(false);
const loadingBelow = ref(false);
const allLogs = ref<ProcessLog[]>([]);
const contextTopID = ref<number>(0);
const contextBottomID = ref<number>(0);
const anchorLog = ref<ProcessLog | null>(null);

const INIT_SIZE = 50;
const CHUNK_SIZE = 25;

// ── 计算属性 ──────────────────────────────────────────────────────────────
const anchorInfo = computed(() => {
  if (!anchorLog.value) return t("logPage.contextMode");
  const { name, using } = anchorLog.value;
  return `${name}${using ? "  ·  " + using : ""}`;
});

// ── 公开接口 ───────────────────────────────────────────────────────────────
defineExpose({
  openAbove: (item: ProcessLog) => openContext(item, "above"),
  openBelow: (item: ProcessLog) => openContext(item, "below"),
});

// ── 打开上下文 ─────────────────────────────────────────────────────────────
const openContext = async (item: ProcessLog, direction: "above" | "below") => {
  allLogs.value = [];
  anchorLog.value = item;
  initialLoading.value = true;

  dialog.value = true;
  await nextTick();
  initTerminal();

  try {
    if (direction === "above") {
      await fetchInitialAbove(item);
    } else {
      await fetchInitialBelow(item);
    }
  } finally {
    initialLoading.value = false;
  }
};

const fetchInitialAbove = async (item: ProcessLog) => {
  const res = await getLog(buildCursorQuery(item.id!, "desc", INIT_SIZE, item.name));
  if (res.code !== 0 || !res.data) {
    snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
    return;
  }
  const fetched = res.data.data || [];
  const before = [...fetched].reverse(); // DESC → ASC

  allLogs.value = [...before, item];
  contextTopID.value = before.length > 0 ? before[0].id! : item.id!;
  contextBottomID.value = item.id!;

  writeAll();
  // 锚点在末尾，终端自然停在底部，正好显示锚点行
};

const fetchInitialBelow = async (item: ProcessLog) => {
  const res = await getLog(buildCursorQuery(item.id!, "asc", INIT_SIZE, item.name));
  if (res.code !== 0 || !res.data) {
    snackbarStore.showErrorMessage(t("logPage.loadLogsFailed"));
    return;
  }
  const fetched = res.data.data || [];

  allLogs.value = [item, ...fetched];
  contextTopID.value = item.id!;
  contextBottomID.value = fetched.length > 0 ? fetched[fetched.length - 1].id! : item.id!;

  // 锚点在头部，内容渲染完成后再滚到顶
  writeAll(() => term?.scrollToTop());
};

// ── 终端初始化 ─────────────────────────────────────────────────────────────
const initTerminal = () => {
  if (!xtermEl.value) return;

  if (term) {
    term.dispose();
    term = null;
  }

  term = new Terminal({
    convertEol: true,
    disableStdin: true,
    cursorBlink: false,
    scrollback: 50000,
    theme: {
      background: "#1a1a1a",
      foreground: "#ececec",
      black: "#1a1a1a",
    },
    fontFamily: '"Consolas", "Monaco", "Courier New", monospace',
    fontSize: 13,
    lineHeight: 1.3,
  });

  fitAddon = new FitAddon();
  try {
    term.loadAddon(new CanvasAddon());
  } catch {
    // canvas 不可用时降级到 DOM 渲染
  }
  term.loadAddon(fitAddon);
  term.open(xtermEl.value);
  fitAddon.fit();

  window.addEventListener("resize", handleResize);
};

const handleResize = () => fitAddon?.fit();

// ── 加载更多上文 ────────────────────────────────────────────────────────────
const loadMoreAbove = async () => {
  if (loadingAbove.value) return;
  loadingAbove.value = true;
  try {
    const res = await getLog(buildCursorQuery(contextTopID.value, "desc", CHUNK_SIZE, anchorLog.value!.name));
    if (res.code !== 0 || !res.data) return;

    const fetched = res.data.data || [];
    if (fetched.length === 0) {
      snackbarStore.showMessage(t("logPage.noMoreAbove"));
      return;
    }
    const newLogs = [...fetched].reverse();

    allLogs.value = [...newLogs, ...allLogs.value];
    contextTopID.value = newLogs[0].id!;

    // 重写后滚到顶部，让用户直接看到新加载的上文
    rewriteTerminal();
  } catch (e) {
    console.error("loadMoreAbove error:", e);
  } finally {
    loadingAbove.value = false;
  }
};

// ── 加载更多下文 ────────────────────────────────────────────────────────────
const loadMoreBelow = async () => {
  if (loadingBelow.value) return;
  loadingBelow.value = true;
  try {
    const res = await getLog(buildCursorQuery(contextBottomID.value, "asc", CHUNK_SIZE, anchorLog.value!.name));
    if (res.code !== 0 || !res.data) return;

    const fetched = res.data.data || [];
    if (fetched.length === 0) {
      snackbarStore.showMessage(t("logPage.noMoreBelow"));
      return;
    }
    allLogs.value = [...allLogs.value, ...fetched];
    contextBottomID.value = fetched[fetched.length - 1].id!;

    appendLogs(fetched);
  } catch (e) {
    console.error("loadMoreBelow error:", e);
  } finally {
    loadingBelow.value = false;
  }
};

// ── 写入 / 重写终端 ────────────────────────────────────────────────────────
/**
 * 将 allLogs 拼成一段完整字符串，通过单次 write 写入 xterm。
 * 通过 callback 确保在内容真正渲染到屏幕后再执行后续操作（如滚动）。
 */
const writeAll = (callback?: () => void) => {
  if (!term) return;
  const lines: string[] = [];

  for (const log of allLogs.value) {
    lines.push(log.log ?? "");
  }

  const content = lines.join("\r\n") + (lines.length ? "\r\n" : "");
  if (callback) {
    term.write(content, callback);
  } else {
    term.write(content);
  }
};

const appendLogs = (logs: ProcessLog[]) => {
  if (!term) return;
  for (const log of logs) {
    term.writeln(log.log ?? "");
  }
};

/**
 * 清空终端后重写全部内容，并将视口滚动到 targetScrollY 行。
 * 用于"加载更多上文"后保持用户视角不发生跳变。
 */
const rewriteTerminal = () => {
  if (!term) return;
  term.reset();
  // 内容渲染完成后再滚到顶部
  writeAll(() => term?.scrollToTop());
};

// ── 查询构建（仅查目标进程的日志） ────────────────────────────────────────
const buildCursorQuery = (
  cursorID: number,
  sort: string,
  size: number,
  processName: string,
): GetLogReq => ({
  cursorID,
  sort,
  page: { from: 0, size },
  filterName: [processName],
});

// ── 清理 ────────────────────────────────────────────────────────────────────
const cleanup = () => {
  window.removeEventListener("resize", handleResize);
  if (term) {
    term.dispose();
    term = null;
  }
  fitAddon = null;
  allLogs.value = [];
  anchorLog.value = null;
};

onUnmounted(cleanup);
</script>

<style scoped>
.lcv {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1a1a1a;
  color: #ececec;
}

.lcv__toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  height: 40px;
  flex-shrink: 0;
  background: #111;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  user-select: none;
}

.lcv__anchor-info {
  font-size: 0.82rem;
  font-weight: 600;
  color: #ececec;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

.lcv__action-bar {
  display: flex;
  justify-content: center;
  flex-shrink: 0;
  background: #111;
}

.lcv__action-bar--top {
  padding: 6px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.lcv__action-bar--bottom {
  padding: 6px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.lcv__terminal {
  flex: 1;
  overflow: hidden;
  transition: opacity 0.15s;
}

/* xterm 撑满容器 */
.lcv__terminal :deep(.xterm) {
  height: 100%;
  padding: 6px 10px;
}

.lcv__terminal :deep(.xterm-screen) {
  height: 100% !important;
}

.lcv__terminal :deep(.xterm-viewport) {
  border-radius: 0;
}

.lcv__init-mask {
  position: absolute;
  inset: 40px 0 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1a1a1a;
  z-index: 10;
}
</style>
