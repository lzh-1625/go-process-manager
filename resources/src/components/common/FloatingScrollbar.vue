<template>
  <div
    v-show="canScroll"
    class="floating-scrollbar"
    :class="{ 'floating-scrollbar--visible': visible || dragging }"
    @pointerdown="onTrackPointerDown"
  >
    <div
      class="floating-scrollbar__thumb"
      :style="thumbStyle"
      @pointerdown.stop="onThumbPointerDown"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";

const canScroll = ref(false);
const visible = ref(false);
const dragging = ref(false);
const thumbTop = ref(0);
const thumbHeight = ref(36);
const viewportHeight = ref(0);

let hideTimer: number | undefined;
let animationFrame: number | undefined;
let resizeObserver: ResizeObserver | undefined;
let mutationObserver: MutationObserver | undefined;
let dragStartY = 0;
let dragStartScrollTop = 0;

const trackPadding = 8;
const minThumbHeight = 36;

const scrollingElement = () => document.scrollingElement || document.documentElement;

const maxScrollTop = () => {
  const el = scrollingElement();
  return Math.max(0, el.scrollHeight - el.clientHeight);
};

const trackHeight = computed(() => Math.max(0, viewportHeight.value - trackPadding * 2));

const thumbStyle = computed(() => ({
  height: `${thumbHeight.value}px`,
  transform: `translateY(${thumbTop.value}px)`,
}));

const showTemporarily = () => {
  visible.value = true;
  window.clearTimeout(hideTimer);

  if (!dragging.value) {
    hideTimer = window.setTimeout(() => {
      visible.value = false;
    }, 900);
  }
};

const updateScrollbar = () => {
  animationFrame = undefined;

  const el = scrollingElement();
  const maxScroll = maxScrollTop();
  viewportHeight.value = window.innerHeight;
  canScroll.value = maxScroll > 1;

  if (!canScroll.value || trackHeight.value <= 0) {
    thumbTop.value = 0;
    return;
  }

  const nextThumbHeight = Math.max(
    minThumbHeight,
    (el.clientHeight / el.scrollHeight) * trackHeight.value
  );
  const maxThumbTop = Math.max(0, trackHeight.value - nextThumbHeight);

  thumbHeight.value = nextThumbHeight;
  thumbTop.value = maxScroll === 0 ? 0 : (el.scrollTop / maxScroll) * maxThumbTop;
};

const scheduleUpdate = () => {
  if (animationFrame !== undefined) return;
  animationFrame = window.requestAnimationFrame(updateScrollbar);
};

const onScroll = () => {
  scheduleUpdate();
  showTemporarily();
};

const scrollToPointer = (clientY: number) => {
  const maxScroll = maxScrollTop();
  const maxThumbTop = Math.max(0, trackHeight.value - thumbHeight.value);
  const nextThumbTop = Math.min(
    maxThumbTop,
    Math.max(0, clientY - trackPadding - thumbHeight.value / 2)
  );

  scrollingElement().scrollTop = maxThumbTop === 0
    ? 0
    : (nextThumbTop / maxThumbTop) * maxScroll;
};

const onTrackPointerDown = (event: PointerEvent) => {
  if (!canScroll.value) return;
  scrollToPointer(event.clientY);
  showTemporarily();
};

const onThumbPointerDown = (event: PointerEvent) => {
  if (!canScroll.value) return;

  dragging.value = true;
  visible.value = true;
  dragStartY = event.clientY;
  dragStartScrollTop = scrollingElement().scrollTop;

  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", onPointerUp, { once: true });
};

const onPointerMove = (event: PointerEvent) => {
  if (!dragging.value) return;

  const maxScroll = maxScrollTop();
  const maxThumbTop = Math.max(1, trackHeight.value - thumbHeight.value);
  const scrollRatio = maxScroll / maxThumbTop;

  scrollingElement().scrollTop = dragStartScrollTop + (event.clientY - dragStartY) * scrollRatio;
};

const onPointerUp = () => {
  dragging.value = false;
  window.removeEventListener("pointermove", onPointerMove);
  showTemporarily();
};

onMounted(() => {
  window.addEventListener("resize", scheduleUpdate);
  document.addEventListener("scroll", onScroll, true);

  resizeObserver = new ResizeObserver(scheduleUpdate);
  resizeObserver.observe(document.body);

  mutationObserver = new MutationObserver(scheduleUpdate);
  mutationObserver.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
  });

  nextTick(scheduleUpdate);
});

onUnmounted(() => {
  window.clearTimeout(hideTimer);

  if (animationFrame !== undefined) {
    window.cancelAnimationFrame(animationFrame);
  }

  window.removeEventListener("resize", scheduleUpdate);
  document.removeEventListener("scroll", onScroll, true);
  window.removeEventListener("pointermove", onPointerMove);
  resizeObserver?.disconnect();
  mutationObserver?.disconnect();
});
</script>

<style scoped>
.floating-scrollbar {
  position: fixed;
  top: 8px;
  right: 2px;
  bottom: 8px;
  z-index: 3000;
  width: 10px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.18s ease;
}

.floating-scrollbar--visible {
  opacity: 1;
  pointer-events: auto;
}

.floating-scrollbar__thumb {
  position: absolute;
  top: 0;
  right: 2px;
  width: 5px;
  min-height: 36px;
  border-radius: 999px;
  background: rgba(var(--v-theme-primary), 0.32);
  box-shadow: 0 0 0 1px rgba(var(--v-theme-surface), 0.18);
  cursor: pointer;
  transition: width 0.14s ease, background-color 0.14s ease;
}

.floating-scrollbar:hover .floating-scrollbar__thumb,
.floating-scrollbar__thumb:hover {
  width: 7px;
  background: rgba(var(--v-theme-primary), 0.52);
}

.floating-scrollbar__thumb:active {
  width: 7px;
  background: rgba(var(--v-theme-primary), 0.68);
}
</style>
