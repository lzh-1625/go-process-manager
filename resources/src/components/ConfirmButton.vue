<template>
  <div>
    <!-- 触发按钮 -->
    <v-btn :color="color" size="small" variant="tonal" @click="dialog = true">
      <slot>{{ label }}</slot>
    </v-btn>

    <!-- 确认弹窗 -->
    <v-dialog v-model="dialog" max-width="400">
      <v-card>
        <v-card-title class="text-h6">{{ title }}</v-card-title>

        <v-card-text>
          {{ message }}
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" @click="dialog = false">取消</v-btn>
          <v-btn :color="color" variant="flat" @click="confirm"> 确认 </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref } from "vue";

const props = defineProps({
  label: { type: String, default: "操作" },
  title: { type: String, default: "确认" },
  message: { type: String, default: "确定要执行此操作吗？" },
  color: { type: String, default: "primary" },
});

const emits = defineEmits(["confirm"]);

const dialog = ref(false);

const confirm = () => {
  dialog.value = false;
  emits("confirm");
};
</script>
