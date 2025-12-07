<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted, computed } from "vue";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { ProcessItem } from "~/src/types/process/process";

const snackbarStore = useSnackbarStore();
const dialog = ref(false);
const props = defineProps<{
  data: ProcessItem;
}>();

const outputLines = ref<string[]>([]);
const inputText = ref("");
const terminalOutput = ref<HTMLElement | null>(null);

const socket = ref<WebSocket | null>(null);


defineExpose({
  wsConnect: () => {
    dialog.value = true;
  },
});

// 使用 watch 监听 dialog 的状态变化
watch(dialog, (newValue) => {
  if (newValue) {
    nextTick(() => {
      initWebSocketStd();
    });
  }
});

const initWebSocketStd = () => {
  const baseUrl = `ws://${window.location.hostname}:8797/api/ws`;
  // const baseUrl = `ws://${window.location.hostname}:${window.location.port}/api/ws`;
  const url = `${baseUrl}?uuid=${props.data.uuid}&token=${localStorage.getItem(
    "token"
  )}`;

  initSocket(url);
};

const initSocket = (url: string) => {
  socket.value = new WebSocket(url);


  socket.value.onopen = () => {
    snackbarStore.showSuccessMessage("终端连接成功");
  };

  socket.value.onmessage = (event) => {
    // 每个消息都是一行
    const line = event.data;
    outputLines.value.push(line);

    // 自动滚动到底部
    nextTick(() => {
      if (terminalOutput.value) {
        terminalOutput.value.scrollTop = terminalOutput.value.scrollHeight;
      }
    });
  };

  socket.value.onclose = () => {
    snackbarStore.showErrorMessage("终端连接断开");
    dialog.value = false;
  };

  socket.value.onerror = (err) => {
    snackbarStore.showErrorMessage("终端连接发生错误");
    console.error("WebSocket Error:", err);
  };
};

const sendInput = () => {
  if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
    snackbarStore.showErrorMessage("终端未连接");
    return;
  }

  if (inputText.value.trim()) {
    // 发送输入到服务器
    socket.value.send(inputText.value );
    inputText.value = "";
  }
};

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Enter") {
    event.preventDefault();
    sendInput();
  }
};

const clearOutput = () => {
  outputLines.value = [];
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
  outputLines.value = [];
  inputText.value = "";
  if (socket.value) {
    socket.value.close();
    socket.value = null;
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
        background-color: #1e1e1e;
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
          <v-btn icon dense dark @click="clearOutput" title="清空输出">
            <v-icon>mdi-trash-can-outline</v-icon>
          </v-btn>
          <v-btn icon dense dark @click="wsClose">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <!-- 终端输出区域 -->
      <div
        ref="terminalOutput"
        class="terminal-output"
      >
        <div
          v-for="(line, index) in outputLines"
          :key="index"
          class="output-line"
        >
          {{ line }}
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="terminal-input">
        <div class="input-wrapper">
          <span class="prompt">$</span>
          <input
            v-model="inputText"
            @keydown="handleKeydown"
            type="text"
            class="input-field"
            placeholder="输入命令并按回车..."
          />
          <v-btn
            icon
            size="small"
            variant="text"
            @click="sendInput"
            class="send-btn"
          >
            <v-icon color="primary">mdi-send</v-icon>
          </v-btn>
        </div>
      </div>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.terminal-output {
  flex-grow: 1;
  overflow-y: auto;
  padding: 12px;
  font-family: "Courier New", Courier, monospace;
  font-size: 14px;
  color: #d4d4d4;
  background-color: #1e1e1e;
  line-height: 1.5;
}

.output-line {
  white-space: pre-wrap;
  word-wrap: break-word;
  margin-bottom: 2px;
}

.terminal-input {
  flex-shrink: 0;
  padding: 8px 12px;
  background-color: #2d2d2d;
  border-top: 1px solid #3e3e3e;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prompt {
  color: #4ec9b0;
  font-family: "Courier New", Courier, monospace;
  font-size: 14px;
  font-weight: bold;
  flex-shrink: 0;
}

.input-field {
  flex-grow: 1;
  background-color: transparent;
  border: none;
  outline: none;
  color: #d4d4d4;
  font-family: "Courier New", Courier, monospace;
  font-size: 14px;
  padding: 4px 8px;
}

.input-field:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.input-field::placeholder {
  color: #6a6a6a;
}

.send-btn {
  flex-shrink: 0;
}

/* 自定义滚动条样式 */
.terminal-output::-webkit-scrollbar {
  width: 10px;
}

.terminal-output::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.terminal-output::-webkit-scrollbar-thumb {
  background: #3e3e3e;
  border-radius: 5px;
}

.terminal-output::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>

