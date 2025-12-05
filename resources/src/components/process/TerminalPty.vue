<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted } from "vue"; // 引入 watch, nextTick 和 onUnmounted
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { ProcessItem } from "~/src/types/process/process";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { AttachAddon } from "xterm-addon-attach";
import { CanvasAddon } from "@xterm/addon-canvas";
import "xterm/css/xterm.css";

const snackbarStore = useSnackbarStore();
const dialog = ref(false);
const props = defineProps<{
  data: ProcessItem;
}>();

const xtermEl = ref<HTMLElement | null>(null);

let socket: WebSocket | null = null;
let term: Terminal | null = null;
const fitAddon = new FitAddon();

defineExpose({
  wsConnect: () => {
    dialog.value = true;
  },
});

// 使用 watch 监听 dialog 的状态变化
watch(dialog, (newValue) => {
  if (newValue) {
    nextTick(() => {
      initWebSocketPty();
    });
  }
});

const initWebSocketPty = () => {
  if (!xtermEl.value) {
    snackbarStore.showErrorMessage("终端容器初始化失败");
    return;
  }
  // 在这里计算初始尺寸更准确
  const initialCols = Math.floor(xtermEl.value.clientWidth / 9);
  const initialRows = Math.floor(xtermEl.value.clientHeight / 19);

  const baseUrl = `ws://${window.location.hostname}:${window.location.port}/api/ws`;
  const url = `${baseUrl}?uuid=${props.data.uuid}&token=${localStorage.getItem(
    "token"
  )}&cols=${initialCols}&rows=${initialRows}`;

  initSocket(url);
};

const initSocket = (url: string) => {
  socket = new WebSocket(url);

  socket.onopen = () => {
    // WebSocket 连接成功后，初始化 Terminal
    initTerm();
  };

  socket.onclose = () => {
    snackbarStore.showErrorMessage("终端连接断开");
    dialog.value = false;
  };

  socket.onerror = (err) => {
    snackbarStore.showErrorMessage("终端连接发生错误");
    console.error("WebSocket Error:", err);
  };
};

const initTerm = () => {
  if (!socket || !xtermEl.value) return;
  const showCursor = props.data.state.state === 3;
  term = new Terminal({
    convertEol: true,
    disableStdin: false,
    cursorBlink: showCursor,
    cursorStyle: "block",
    theme: {
      foreground: "#ECECEC",
      cursor:  "help"
    },
  });

  const attachAddon = new AttachAddon(socket);

  term.loadAddon(new CanvasAddon()); // 推荐先加载渲染器
  term.loadAddon(attachAddon);
  term.loadAddon(fitAddon);

  term.open(xtermEl.value);

  // 在打开后执行 fit() 来适配尺寸
  fitAddon.fit();
  term.focus();
  window.addEventListener("resize", handleResize);
};

const handleResize = () => {
  fitAddon.fit();
};

const wsClose = () => {
  dialog.value = false;
  cleanup();
};

const toolbarColor = computed(() => {
  if (props.data.state.state == 3) {
    return;
  }
  return "red";
});

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

<template>
  <v-dialog
    fullscreen
    hide-overlay
    transition="dialog-bottom-transition"
    v-model="dialog"
    @update:modelValue="(val) => !val && cleanup()"
  >
    <v-card
      style="
        height: 100%;
        background-color: black;
        display: flex;
        flex-direction: column;
      "
    >
      <v-toolbar
        dense
        :color="toolbarColor"
        dark
        style="height: 35px; flex-grow: 0"
      >
        <v-toolbar-title style="height: 100%"
          >{{ props.data.name }} ({{ props.data.termType }})</v-toolbar-title
        >
        <v-spacer></v-spacer>
        <v-toolbar-items style="height: 35px">
          <v-btn icon dense dark @click="wsClose">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>
      <div
        id="xterm"
        ref="xtermEl"
        style="flex-grow: 1; height: 100%; width: 100%"
      ></div>
    </v-card>
  </v-dialog>
</template>

<style>
#xterm .terminal {
  height: 100%;
}
</style>
