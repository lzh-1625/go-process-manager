<template>
  <v-container fluid class="py-6 px-8">
    <v-card class="rounded-lg">
      <div
        v-if="loading"
        class="h-full d-flex flex-grow-1 align-center justify-center"
        style="min-height: 400px"
      >
        <v-progress-circular
          indeterminate
          color="primary"
        ></v-progress-circular>
      </div>

      <div v-else>
        <h6 class="text-h6 font-weight-bold pa-5 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-bell-ring</v-icon>
          <span class="flex-fill">{{ $t('pushPage.title') }}</span>
          <v-btn
            icon
            variant="text"
            size="small"
            @click="refreshPushList"
          >
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
          <v-btn
            color="primary"
            variant="tonal"
            size="small"
            @click="openAddDialog"
          >
            <v-icon left>mdi-plus</v-icon>
            {{ $t('pushPage.addPush') }}
          </v-btn>
        </h6>

        <v-table class="pa-3">
          <thead>
            <tr>
              <th class="text-left" v-for="header in headers" :key="header.title">
                {{ header.title }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in paginatedPushList" :key="item.id">
              <td>
                <v-chip
                  :color="getMethodColor(item.method)"
                  size="small"
                  class="font-weight-bold"
                >
                  {{ item.method }}
                </v-chip>
              </td>
              <td>
                <div class="text-truncate" style="max-width: 300px;" :title="item.url">
                  {{ item.url }}
                </div>
              </td>
              <td>
                <div class="text-truncate" style="max-width: 200px;" :title="item.body">
                  {{ item.body || '-' }}
                </div>
              </td>
              <td>
                <div class="text-truncate" style="max-width: 150px;" :title="item.remark">
                  {{ item.remark || '-' }}
                </div>
              </td>
              <td>
                <v-switch
                  :model-value="item.enable"
                  color="primary"
                  density="compact"
                  hide-details
                  @update:model-value="toggleEnable(item)"
                ></v-switch>
              </td>
              <td>
                <v-icon class="mr-2" @click="openEditDialog(item)" size="small">
                  mdi-pencil
                </v-icon>
                <v-icon @click="confirmDelete(item)" size="small">
                  mdi-delete
                </v-icon>
              </td>
            </tr>
            <tr v-if="pushList.length === 0">
              <td colspan="6" class="text-center text-secondary pa-8">
                {{ $t('common.noData') }}
              </td>
            </tr>
          </tbody>
        </v-table>

        <div class="text-center pa-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            :total-visible="7"
            density="compact"
            @update:model-value="handlePageChange"
          ></v-pagination>
          <div class="mt-2 text-caption text-secondary">
            {{ $t('pushPage.totalPushes', { n: pushList.length }) }}
          </div>
        </div>
      </div>
    </v-card>

    <v-dialog v-model="dialog" max-width="600px">
      <v-card class="rounded-xl">
        <v-card-title class="text-h6 font-weight-medium">
          {{ isEdit ? $t('pushPage.editPush') : $t('pushPage.addPush') }}
        </v-card-title>

        <v-divider></v-divider>

        <v-card-text class="pt-6">
          <v-form ref="formRef" v-model="formValid">
            <v-container fluid>
              <v-row dense>
                <v-col cols="12" sm="4">
                  <v-select
                    v-model="form.method"
                    :label="$t('pushPage.httpMethod')"
                    :items="methodOptions"
                    variant="outlined"
                    density="comfortable"
                    :rules="[rules.required]"
                  ></v-select>
                </v-col>
                <v-col cols="12" sm="8">
                  <v-text-field
                    v-model="form.url"
                    :label="$t('pushPage.pushUrl')"
                    variant="outlined"
                    density="comfortable"
                    :rules="[rules.required, rules.url]"
                    placeholder="https://example.com/webhook"
                  ></v-text-field>
                </v-col>
                <v-col cols="12">
                  <v-textarea
                    v-model="form.body"
                    :label="$t('pushPage.requestBody')"
                    variant="outlined"
                    density="comfortable"
                    rows="4"
                    placeholder='{"message": "{$message}", "user": "{$user}"}'
                    :hint="$t('pushPage.placeholderHint')"
                    persistent-hint
                  ></v-textarea>
                </v-col>
                <v-col cols="12">
                  <v-text-field
                    v-model="form.remark"
                    :label="$t('common.remark')"
                    variant="outlined"
                    density="comfortable"
                    :placeholder="$t('pushPage.pushConfigDesc')"
                  ></v-text-field>
                </v-col>
                <v-col cols="12">
                  <v-switch
                    v-model="form.enable"
                    :label="$t('pushPage.enablePush')"
                    color="primary"
                    hide-details
                  ></v-switch>
                </v-col>
              </v-row>
            </v-container>
          </v-form>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions class="justify-end pa-4">
          <v-btn text @click="closeDialog">{{ $t('common.cancel') }}</v-btn>
          <v-btn
            color="primary"
            @click="submitForm"
            :loading="submitLoading"
            :disabled="!formValid"
          >
            {{ isEdit ? $t('common.save') : $t('common.create') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="deleteDialog" max-width="480">
      <v-card class="rounded-xl">
        <v-card-title class="text-h6 font-weight-medium">{{ $t('pushPage.confirmDelete') }}</v-card-title>

        <v-divider></v-divider>

        <v-card-text class="pt-6">
          {{ $t('pushPage.confirmDeleteMsg') }}
          <div class="text-caption text-secondary mt-2">
            {{ $t('common.remark') }}: {{ deleteItem?.remark || $t('common.none') }}
          </div>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions class="justify-end pa-4">
          <v-btn text @click="deleteDialog = false">{{ $t('common.cancel') }}</v-btn>
          <v-btn
            color="error"
            @click="handleDelete"
            :loading="deleteLoading"
          >
            {{ $t('common.delete') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { getPushList, createPush, editPush, deletePush } from "~/src/api/push";
import type { PushItem } from "~/src/types/push/push";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();

const loading = ref(false);
const pushList = ref<PushItem[]>([]);

const dialog = ref(false);
const isEdit = ref(false);
const formRef = ref();
const formValid = ref(false);
const submitLoading = ref(false);

const deleteDialog = ref(false);
const deleteItem = ref<PushItem | null>(null);
const deleteLoading = ref(false);

const currentPage = ref(1);
const pageSize = ref(10);

const methodOptions = ["GET", "POST"];

const defaultForm = {
  id: 0,
  method: "POST",
  url: "",
  body: "",
  remark: "",
  enable: true,
};
const form = ref({ ...defaultForm });

const headers = computed(() => [
  { title: t("pushPage.httpMethod"), key: "method" },
  { title: t("pushPage.pushUrl"), key: "url" },
  { title: t("pushPage.requestBody"), key: "body" },
  { title: t("common.remark"), key: "remark" },
  { title: t("common.enable"), key: "enable" },
  { title: t("common.operation"), key: "actions" },
]);

const rules = {
  required: (v: string) => !!v || t("common.required"),
  url: (v: string) => {
    if (!v) return true;
    try {
      new URL(v);
      return true;
    } catch {
      return t("pushPage.invalidUrl");
    }
  },
};

const totalPages = computed(() => {
  return Math.ceil(pushList.value.length / pageSize.value);
});

const paginatedPushList = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return pushList.value.slice(start, end);
});

const handlePageChange = (page: number) => {
  currentPage.value = page;
};

const getMethodColor = (method: string) => {
  const colors: Record<string, string> = {
    GET: "success",
    POST: "primary",
    PUT: "warning",
    DELETE: "error",
  };
  return colors[method] || "grey";
};

const loadPushList = async () => {
  loading.value = true;
  try {
    const res = await getPushList();
    if (res.code === 0 && res.data) {
      pushList.value = res.data;
    }
  } catch (error) {
    console.error("加载推送列表失败:", error);
    snackbarStore.showErrorMessage(t("pushPage.loadPushListFailed"));
  } finally {
    loading.value = false;
  }
};

const refreshPushList = () => {
  loadPushList();
};

const openAddDialog = () => {
  isEdit.value = false;
  form.value = { ...defaultForm };
  dialog.value = true;
};

const openEditDialog = (item: PushItem) => {
  isEdit.value = true;
  form.value = { ...item };
  dialog.value = true;
};

const closeDialog = () => {
  dialog.value = false;
  form.value = { ...defaultForm };
};

const submitForm = async () => {
  if (!formValid.value) return;

  submitLoading.value = true;
  try {
    const data = {
      id: form.value.id,
      method: form.value.method,
      url: form.value.url,
      body: form.value.body,
      remark: form.value.remark,
      enable: form.value.enable,
    };

    let res;
    if (isEdit.value) {
      res = await editPush(data);
    } else {
      res = await createPush(data);
    }

    if (res.code === 0) {
      snackbarStore.showSuccessMessage(isEdit.value ? t("common.saveSuccess") : t("common.createSuccess"));
      closeDialog();
      loadPushList();
    }
  } catch (error) {
    console.error("提交失败:", error);
    snackbarStore.showErrorMessage(isEdit.value ? t("pushPage.saveFailed") : t("pushPage.createFailed"));
  } finally {
    submitLoading.value = false;
  }
};

const toggleEnable = async (item: PushItem) => {
  try {
    const res = await editPush({
      ...item,
      enable: !item.enable,
    });
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(item.enable ? t("pushPage.disabled") : t("pushPage.enabled"));
      loadPushList();
    }
  } catch (error) {
    console.error("切换状态失败:", error);
    snackbarStore.showErrorMessage(t("pushPage.operationFailed"));
  }
};

const confirmDelete = (item: PushItem) => {
  deleteItem.value = item;
  deleteDialog.value = true;
};

const handleDelete = async () => {
  if (!deleteItem.value) return;

  deleteLoading.value = true;
  try {
    const res = await deletePush(deleteItem.value.id);
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(t("common.deleteSuccess"));
      deleteDialog.value = false;
      deleteItem.value = null;
      loadPushList();
    }
  } catch (error) {
    console.error("删除失败:", error);
    snackbarStore.showErrorMessage(t("pushPage.deleteFailed"));
  } finally {
    deleteLoading.value = false;
  }
};

onMounted(() => {
  loadPushList();
});
</script>

<style lang="scss" scoped>
.v-table {
  table {
    padding: 4px;
    padding-bottom: 8px;

    th {
      text-transform: uppercase;
      white-space: nowrap;
    }

    td {
      border-bottom: 0 !important;
    }

    tbody {
      tr {
        transition: box-shadow 0.2s, transform 0.2s;

        &:not(.v-data-table__selected):hover {
          box-shadow: 0 3px 15px -2px rgba(0, 0, 0, 0.12);
          transform: translateY(-4px);
        }
      }
    }
  }
}
</style>
