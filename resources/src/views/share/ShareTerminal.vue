<template>
  <div style="height: 100vh; width: 100vw; background: black; overflow: hidden">
    <div ref="xtermEl" style="height: 100%; width: 100%"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { AttachAddon } from "xterm-addon-attach";
import { CanvasAddon } from "@xterm/addon-canvas";
import "xterm/css/xterm.css";

const snackbarStore = useSnackbarStore();
const xtermEl = ref(null);

let socket = null;
let term = null;
let fitAddon = null;

onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get("token");

  if (!token) {
    snackbarStore.showErrorMessage("缺少访问令牌");
    return;
  }

  initWebSocketPty(token);
});

const initWebSocketPty = (token) => {
  if (!xtermEl.value) {
    snackbarStore.showErrorMessage("终端容器初始化失败");
    return;
  }

  // 计算初始尺寸
  const initialCols = Math.floor(window.innerWidth / 9);
  const initialRows = Math.floor(window.innerHeight / 19);

  const baseUrl = `ws://${window.location.hostname}:8797/api/ws/share`;
  const url = `${baseUrl}?token=${token}&cols=${initialCols}&rows=${initialRows}`;

  initSocket(url);
};

const initSocket = (url) => {
  socket = new WebSocket(url);

  socket.onopen = () => {
    // WebSocket 连接成功后，初始化 Terminal
    initTerm();
  };

  socket.onclose = () => {
    snackbarStore.showErrorMessage("终端连接已断开");
    cleanup();
  };

  socket.onerror = (err) => {
    snackbarStore.showErrorMessage("终端连接发生错误");
    console.error("WebSocket Error:", err);
  };
};

const initTerm = () => {
  if (!socket || !xtermEl.value) return;

  term = new Terminal({
    convertEol: true,
    disableStdin: false,
    cursorBlink: true,
    cursorStyle: "block",
    lineHeight: 1.12,
    theme: {
      foreground: "#ECECEC",
      cursor: "help",
    },
  });

  const attachAddon = new AttachAddon(socket);
  fitAddon = new FitAddon();

  term.loadAddon(new CanvasAddon());
  term.loadAddon(attachAddon);
  term.loadAddon(fitAddon);

  term.open(xtermEl.value);

  // 在打开后执行 fit() 来适配尺寸
  fitAddon.fit();
  term.focus();

  window.addEventListener("resize", handleResize);
};

const handleResize = () => {
  if (fitAddon) {
    fitAddon.fit();
  }
};

const cleanup = () => {
  window.removeEventListener("resize", handleResize);

  if (term) {
    term.dispose();
    term = null;
  }

  if (socket) {
    socket.close();
    socket = null;
  }
};

onUnmounted(() => {
  cleanup();
});
</script>

<style scoped>
/* 确保终端填充整个容器 */
:deep(.xterm) {
  height: 100% !important;
  width: 100% !important;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
