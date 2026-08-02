<script setup lang="ts">
import SnackbarItem from "@/components/common/SnackbarItem.vue";
import { useSnackbarStore } from "@/stores/snackbarStore";

const snackbarStore = useSnackbarStore();
</script>

<template>
  <div class="notification-stack" aria-live="polite">
    <TransitionGroup name="notification-stack">
      <SnackbarItem
        v-for="notification in snackbarStore.notifications"
        :key="notification.id"
        :notification="notification"
        @dismiss="snackbarStore.removeMessage"
      />
    </TransitionGroup>
  </div>
</template>

<style scoped>
.notification-stack {
  position: fixed;
  z-index: 2500;
  top: 16px;
  right: 16px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
  pointer-events: none;
}

.notification-stack-enter-active,
.notification-stack-leave-active,
.notification-stack-move {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.notification-stack-enter-from,
.notification-stack-leave-to {
  opacity: 0;
  transform: translateY(-12px);
}

.notification-stack-leave-active {
  position: absolute;
}

@media (max-width: 600px) {
  .notification-stack {
    top: 12px;
    right: 12px;
    gap: 8px;
  }
}
</style>
