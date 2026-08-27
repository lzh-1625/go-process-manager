<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted, computed } from "vue"; // 引入 watch, nextTick 和 onUnmounted
import { useI18n } from "vue-i18n";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { ProcessItem } from "~/src/types/process/process";
import ProcessUserConnections from "./ProcessUserConnections.vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "@xterm/addon-canvas";
import * as Zmodem from "zmodem.js-ex/src/zmodem_browser";
import "xterm/css/xterm.css";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();
const dialog = ref(false);
const terminalDisconnected = ref(false);
const zmodemTransferActive = ref(false);
const zmodemUploadDialog = ref(false);
const zmodemUploadFiles = ref<File[]>([]);
const props = defineProps<{
  data: ProcessItem;
}>();

const xtermEl = ref<HTMLElement | null>(null);

let socket: WebSocket | null = null;
let term: Terminal | null = null;
let zmodemSession: any = null;
let zmodemUploadSession: any = null;
let zmodemSentry: any = null;
const fitAddon = new FitAddon();

defineExpose({
  wsConnect: () => {
    terminalDisconnected.value = false;
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
    snackbarStore.showErrorMessage(t("processCardPage.terminalInitFailed"));
    return;
  }
  // 在这里计算初始尺寸更准确
  const initialCols = Math.floor(xtermEl.value.clientWidth / 9);
  const initialRows = Math.floor(xtermEl.value.clientHeight / 19);
  document.cookie = `Authorization=bearer ${localStorage.getItem("token")!}; path=/api/ws; SameSite=Lax`;
  const scheme = window.location.protocol === "https:" ? "wss" : "ws";
  const baseUrl = `${scheme}://${window.location.hostname}:${window.location.port}/api/ws`;
  // const baseUrl = `ws://${window.location.hostname}:8797/api/ws`;
  const url = `${baseUrl}?uuid=${props.data.uuid}&cols=${initialCols}&rows=${initialRows}`;

  initSocket(url);
};

const initSocket = (url: string) => {
  socket = new WebSocket(url);
  socket.binaryType = "arraybuffer";
  socket.onopen = () => {
    // WebSocket 连接成功后，初始化 Terminal
    initTerm();
  };

  socket.onclose = () => {
    terminalDisconnected.value = true;
    if (term) {
      term.options.disableStdin = true;
      term.options.cursorBlink = false;
    }
    snackbarStore.showInfoMessage(t("processCardPage.terminalDisconnected"));
  };

  socket.onerror = (err) => {
    snackbarStore.showErrorMessage(t("processCardPage.terminalError"));
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

  term.loadAddon(new CanvasAddon()); // 推荐先加载渲染器
  term.loadAddon(fitAddon);

  term.open(xtermEl.value);

  zmodemSentry = new Zmodem.Sentry({
    to_terminal: (octets: number[] | Uint8Array) =>
      term?.write(new Uint8Array(octets)),
    sender: (octets: number[] | Uint8Array) => {
      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(new Uint8Array(octets));
      }
    },
    on_detect: onZmodemDetected,
    on_retract: resetZmodemTransfer,
  });
  socket.onmessage = (event) => {
    try {
      zmodemSentry.consume(event.data);
    } catch (err) {
      console.error("ZMODEM Error:", err);
      abortZmodemTransfer();
    }
  };
  term.onData((data) => {
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(data);
    }
  });

  // 在打开后执行 fit() 来适配尺寸
  fitAddon.fit();
  term.focus();
  window.addEventListener("resize", handleResize);
};

const onZmodemDetected = (detection: any) => {
  const session = detection.confirm();
  zmodemSession = session;
  zmodemTransferActive.value = true;
  if (term) {
    term.options.disableStdin = true;
  }

  session.on("session_end", resetZmodemTransfer);

  if (session.type === "send") {
    zmodemUploadSession = session;
    zmodemUploadDialog.value = true;
    return;
  }

  session.on("offer", (offer: any) => {
    const fileName = offer.get_details().name;
    offer.accept().then((payloads) => {
      Zmodem.Browser.save_to_disk(payloads, fileName);
    }).catch(abortZmodemTransfer);
  });
  session.start();
};

const sendZmodemFiles = () => {
  if (!zmodemUploadSession || zmodemUploadFiles.value.length === 0) return;

  zmodemUploadDialog.value = false;
  Zmodem.Browser.send_files(zmodemUploadSession, zmodemUploadFiles.value)
    .then(() => zmodemUploadSession?.close())
    .catch(abortZmodemTransfer);
};

const cancelZmodemUpload = () => {
  abortZmodemTransfer();
};

const abortZmodemTransfer = () => {
  zmodemSession?.abort();
  resetZmodemTransfer();
};

const resetZmodemTransfer = () => {
  zmodemTransferActive.value = false;
  zmodemUploadDialog.value = false;
  zmodemUploadFiles.value = [];
  zmodemSession = null;
  zmodemUploadSession = null;
  if (term && !terminalDisconnected.value) {
    term.options.disableStdin = false;
  }
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
  abortZmodemTransfer();
  zmodemSentry = null;
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
        <v-toolbar-title style="height: 100%">
          {{ props.data.name }}
          <ProcessUserConnections :users="props.data.user" class="ml-2" />
          <span v-if="terminalDisconnected" class="terminal-disconnected">
            {{ $t("processCardPage.terminalDisconnected") }}
          </span>
          <span v-if="zmodemTransferActive" class="zmodem-transfer-active">
            {{ $t("processCardPage.zmodemTransferActive") }}
          </span>
          <v-chip
            v-if="props.data.controller"
            size="x-small"
            color="warning"
            variant="tonal"
            class="ml-2 process-user-chip"
          >
            <v-icon size="10" class="mr-1">mdi-account-lock</v-icon>
            {{ $t("processCardPage.terminalControlledBy", { controller: props.data.controller }) }}
          </v-chip>
        </v-toolbar-title>
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

  <v-dialog v-model="zmodemUploadDialog" persistent max-width="500">
    <v-card>
      <v-card-title>{{ $t("processCardPage.zmodemUploadTitle") }}</v-card-title>
      <v-card-text>
        {{ $t("processCardPage.zmodemUploadDescription") }}
        <v-file-input
          v-model="zmodemUploadFiles"
          class="mt-4"
          multiple
          show-size
          :label="$t('processCardPage.zmodemUploadFiles')"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="cancelZmodemUpload">{{ $t("common.cancel") }}</v-btn>
        <v-btn
          color="primary"
          :disabled="zmodemUploadFiles.length === 0"
          @click="sendZmodemFiles"
        >
          {{ $t("processCardPage.zmodemUpload") }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style>
#xterm .terminal {
  height: 100%;
}

.terminal-disconnected {
  margin-left: 8px;
  font-size: 12px;
  opacity: 0.8;
}

.zmodem-transfer-active {
  margin-left: 8px;
  font-size: 12px;
  opacity: 0.8;
}

#xterm .xterm-viewport {
  scrollbar-width: thin;
  scrollbar-color: rgba(var(--v-theme-primary), 0.7) transparent;
}

#xterm .xterm-viewport::-webkit-scrollbar {
  width: 7px;
  height: 7px;
  display: block;
}

#xterm .xterm-viewport::-webkit-scrollbar-track {
  background: transparent;
}

#xterm .xterm-viewport::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(var(--v-theme-primary), 0.52);
}

#xterm .xterm-viewport::-webkit-scrollbar-thumb:hover {
  background: rgba(var(--v-theme-primary), 0.72);
}
</style>
