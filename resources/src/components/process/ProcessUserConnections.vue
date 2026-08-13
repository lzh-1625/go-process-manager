<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  users: Record<string, number>;
}>();

const userConnections = computed(() =>
  Object.entries(props.users).sort(([left], [right]) =>
    left.localeCompare(right)
  )
);
</script>

<template>
  <span v-if="userConnections.length" class="user-list">
    <v-chip
      v-for="[username, connectionCount] in userConnections"
      :key="username"
      color="grey"
      size="x-small"
      variant="tonal"
      class="user-chip"
    >
      <v-icon size="10" class="mr-1">mdi-account</v-icon>
      {{ username }}
      <span v-if="connectionCount > 1" class="user-connection-count">
        {{ connectionCount }}
      </span>
    </v-chip>
  </span>
</template>

<style scoped>
.user-list {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  vertical-align: middle;
}

.user-chip {
  position: relative;
  min-height: 16px !important;
  height: 16px !important;
  padding: 0 11px 0 4px !important;
  overflow: visible;
  font-size: 10px !important;
}

.user-connection-count {
  position: absolute;
  top: -5px;
  right: -4px;
  display: flex;
  min-width: 11px;
  height: 11px;
  align-items: center;
  justify-content: center;
  padding: 0 2px;
  border: 1px solid rgb(var(--v-theme-surface));
  border-radius: 6px;
  background: rgb(var(--v-theme-primary));
  color: rgb(var(--v-theme-on-primary));
  font-size: 8px;
  font-weight: 700;
  line-height: 1;
  pointer-events: none;
}
</style>
