<template>
  <div style="height: 100vh; width: 100vw; background: black; overflow: hidden">
    <div ref="xtermEl" style="height: 100%; width: 100%"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { AttachAddon } from "xterm-addon-attach";
import { CanvasAddon } from "@xterm/addon-canvas";
import "xterm/css/xterm.css";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();
const xtermEl = ref(null);

let socket = null;
let term = null;
let fitAddon = null;

onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get("token");

  if (!token) {
    snackbarStore.showErrorMessage(t("shareTerminal.missingToken"));
    return;
  }

  initWebSocketPty(token);
});

const initWebSocketPty = (token) => {
  if (!xtermEl.value) {
    snackbarStore.showErrorMessage(t("shareTerminal.terminalInitFailed"));
    return;
  }

  const initialCols = Math.floor(window.innerWidth / 9);
  const initialRows = Math.floor(window.innerHeight / 19);

  const baseUrl = `ws://${window.location.hostname}:8797/api/ws/share`;
  const url = `${baseUrl}?token=${token}&cols=${initialCols}&rows=${initialRows}`;

  initSocket(url);
};

const initSocket = (url) => {
  socket = new WebSocket(url);

  socket.onopen = () => {
    initTerm();
  };

  socket.onclose = () => {
    snackbarStore.showErrorMessage(t("shareTerminal.connectionClosed"));
    cleanup();
  };

  socket.onerror = (err) => {
    snackbarStore.showErrorMessage(t("shareTerminal.connectionError"));
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
:deep(.xterm) {
  height: 100% !important;
  width: 100% !important;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
