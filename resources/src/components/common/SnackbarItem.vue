<script setup lang="ts">
import type { SnackbarMessage } from "@/stores/snackbarStore";

const props = defineProps<{
  notification: SnackbarMessage;
}>();

const emit = defineEmits<{
  dismiss: [id: number];
}>();

const icons = {
  info: "mdi-information-outline",
  success: "mdi-check-circle-outline",
  error: "mdi-alert-circle-outline",
  warning: "mdi-alert-outline",
};

const type = computed(() => props.notification.type || "info");
const icon = computed(() => icons[type.value]);

let dismissTimer: ReturnType<typeof setTimeout> | null = null;

onMounted(() => {
  dismissTimer = setTimeout(() => {
    emit("dismiss", props.notification.id);
  }, 2000);
});

onBeforeUnmount(() => {
  if (dismissTimer) {
    clearTimeout(dismissTimer);
  }
});
</script>

<template>
  <v-card
    :class="[`notification-card--${type}`]"
    class="notification-card"
    color="surface"
    elevation="8"
    role="status"
  >
    <v-avatar
      :color="type"
      class="notification-icon"
      size="36"
      variant="tonal"
    >
      <v-icon size="21">{{ icon }}</v-icon>
    </v-avatar>

    <span class="notification-message">{{ notification.message }}</span>

    <v-btn
      :aria-label="$t('common.close')"
      class="notification-close"
      icon="mdi-close"
      size="small"
      variant="text"
      @click="emit('dismiss', notification.id)"
    />
  </v-card>
</template>

<style scoped>
.notification-card {
  display: flex;
  width: min(520px, calc(100vw - 32px));
  min-height: 64px;
  align-items: center;
  gap: 12px;
  padding: 12px 10px 12px 16px;
  border-inline-start: 4px solid rgb(var(--v-theme-info));
  border-radius: 12px;
  pointer-events: auto;
}

.notification-card--success {
  border-inline-start-color: rgb(var(--v-theme-success));
}

.notification-card--error {
  border-inline-start-color: rgb(var(--v-theme-error));
}

.notification-card--warning {
  border-inline-start-color: rgb(var(--v-theme-warning));
}

.notification-icon,
.notification-close {
  flex: 0 0 auto;
}

.notification-message {
  flex: 1;
  overflow-wrap: anywhere;
  color: rgb(var(--v-theme-on-surface));
  font-size: 0.875rem;
  font-weight: 500;
  line-height: 1.45;
}

.notification-close {
  color: rgb(var(--v-theme-on-surface-variant));
}

@media (max-width: 600px) {
  .notification-card {
    width: calc(100vw - 24px);
  }
}
</style>
