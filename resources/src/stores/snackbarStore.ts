import { defineStore } from "pinia";

export type MessageType = "" | "info" | "success" | "error" | "warning";

export interface SnackbarMessage {
  id: number;
  message: string;
  type: MessageType;
}

const maxVisibleMessages = 5;

export const useSnackbarStore = defineStore({
  id: "snackbarStore",
  state: () => ({
    nextId: 1,
    notifications: [] as SnackbarMessage[],
  }),

  persist: {
    enabled: true,
    strategies: [{ storage: localStorage, paths: [""] }],
  },

  getters: {},
  actions: {
    enqueueMessage(message: string, type: MessageType) {
      this.notifications.push({ id: this.nextId++, message, type });
      const overflow = this.notifications.length - maxVisibleMessages;
      if (overflow > 0) {
        this.notifications.splice(0, overflow);
      }
    },

    removeMessage(id: number) {
      const index = this.notifications.findIndex(
        (notification) => notification.id === id,
      );
      if (index >= 0) {
        this.notifications.splice(index, 1);
      }
    },

    showMessage(message: string) {
      this.enqueueMessage(message, "");
    },

    showErrorMessage(message: string) {
      this.enqueueMessage(message, "error");
    },
    showSuccessMessage(message: string) {
      this.enqueueMessage(message, "success");
    },
    showInfoMessage(message: string) {
      this.enqueueMessage(message, "info");
    },
    showWarningMessage(message: string) {
      this.enqueueMessage(message, "warning");
    },
  },
});
