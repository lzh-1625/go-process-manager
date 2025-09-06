<script setup lang="ts">
import { ref } from "vue";
import { getProcessConfig, putProcessConfig } from "~/src/api/process";
import { getPushList } from "~/src/api/push";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
import { ProcessConfig, ProcessItem } from "~/src/types/process/process";

const snackbarStore = useSnackbarStore();
const dialog = ref(false);
const configForm = ref<Partial<ProcessConfig>>({});
const pushItems = ref<{ value: any; label: string }[]>([]);
const pushSelectedValues = ref([]);

watch(
  pushSelectedValues,
  (newValues) => {
    configForm.value.pushIds = JSON.stringify(newValues);
  },
  { deep: true }
);

const props = defineProps<{
  data: ProcessItem;
}>();

defineExpose({
  openConfigDialog: () => {
    getConfig();
    initPushItem();
    dialog.value = true;
  },
});

const getConfig = () => {
  getProcessConfig(props.data.uuid).then((e) => {
    // 使用 Object.assign 来更新响应式对象，而不是替换它
    if (e.data) {
      Object.assign(configForm.value, e.data);
      pushSelectedValues.value = JSON.parse(
        (e.data!.pushIds as string) == "" ? "[]" : (e.data!.pushIds as string)
      );
    }
  });
};

const updateJsonString = () => {
  configForm.value.pushIds = JSON.stringify(pushSelectedValues);
};

const editConfig = () => {
  putProcessConfig(configForm.value).then((e) => {
    if (e.code === 0) {
      snackbarStore.showSuccessMessage("sucess");
      dialog.value = false; // 成功后通常会关闭对话框
    }
  });
};

const initPushItem = () => {
  getPushList().then((resp) => {
    // 3. 更新 ref 的 .value
    if (resp.data) {
      pushItems.value = resp.data.map((e) => ({
        value: e.id,
        label: `${e.remark} [${e.id}]`,
      }));
    }
  });
};
</script>

<template>
  <v-dialog v-model="dialog" width="700">
    <v-card>
      <v-card-title class="text-h5 grey lighten-2">
        <v-icon left>mdi-cog</v-icon>
        设置
      </v-card-title>

      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="12" md="6">
              <v-text-field
                label="进程名称"
                v-model="configForm.name"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                label="工作目录"
                v-model="configForm.cwd"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>

            <v-col cols="12" md="6">
              <v-text-field
                label="启动命令"
                v-model="configForm.cmd"
                variant="outlined"
                density="compact"
              ></v-text-field>
            </v-col>
          </v-row>

          <v-row>
            <v-col cols="12" md="6">
              <v-select
                label="终端类型"
                disabled
                v-model="configForm.termType"
                :items="['pty', 'std']"
                variant="outlined"
                density="compact"
              ></v-select>
            </v-col>
            <v-col cols="12" md="6">
              <v-select
                v-model="pushSelectedValues"
                @change="updateJsonString"
                :items="pushItems"
                item-title="label"
                item-value="value"
                chips
                label="状态推送"
                multiple
                variant="outlined"
                density="compact"
              ></v-select>
            </v-col>
          </v-row>

          <v-divider class="my-4"></v-divider>

          <v-row align="center">
            <v-col cols="12" sm="3">
              <v-switch
                :disabled="props.data.state?.state === 1"
                v-model="configForm.cgroupEnable"
                label="资源限制"
                color="primary"
                hide-details
              ></v-switch>
            </v-col>
            <v-col cols="12" sm="4">
              <v-text-field
                :disabled="
                  !configForm.cgroupEnable || props.data.state?.state === 3
                "
                label="CPU 限制 (%)"
                type="number"
                v-model.number="configForm.cpuLimit"
                variant="outlined"
                density="compact"
                hide-details="auto"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="4">
              <v-text-field
                :disabled="
                  !configForm.cgroupEnable || props.data.state?.state === 3
                "
                label="内存限制 (MB)"
                type="number"
                v-model.number="configForm.memoryLimit"
                variant="outlined"
                density="compact"
                hide-details="auto"
              ></v-text-field>
            </v-col>
          </v-row>

          <v-divider class="my-4"></v-divider>

          <v-row>
            <v-col cols="12" sm="4">
              <v-switch
                v-model="configForm.autoRestart"
                label="自动重启"
                color="primary"
                hide-details
              ></v-switch>
            </v-col>
            <v-col cols="12" sm="4">
              <v-switch
                :disabled="!configForm.autoRestart"
                v-model="configForm.compulsoryRestart"
                label="强制重启"
                color="primary"
                hide-details
              ></v-switch>
            </v-col>
            <v-col cols="12" sm="4">
              <v-switch
                v-model="configForm.logReport"
                label="日志上报"
                color="primary"
                hide-details
              ></v-switch>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn variant="text" color="grey-darken-1" @click="dialog = false">
          <v-icon left>mdi-close</v-icon>
          取消
        </v-btn>
        <v-btn variant="flat" color="primary" @click="editConfig">
          <v-icon left>mdi-check</v-icon>
          确认
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
