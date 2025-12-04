<template>
  <v-container fluid class="py-6 px-8">
    <!-- 页面标题和操作按钮 -->
    <v-card class="mb-6 rounded-2xl elevation-3">
      <div class="pa-4 d-flex align-center justify-space-between flex-wrap">
        <div class="d-flex align-center mb-2 mb-sm-0">
          <v-icon size="40" color="primary" class="mr-3">mdi-bell-ring</v-icon>
          <span class="text-h5 font-weight-bold text-primary">推送管理</span>
        </div>
        <v-btn
          color="primary"
          variant="flat"
          class="rounded-lg px-4"
          @click="openAddDialog"
        >
          <v-icon start>mdi-plus</v-icon>
          新增推送
        </v-btn>
      </div>
    </v-card>

    <!-- 推送列表 -->
    <v-card class="rounded-2xl elevation-2" :loading="loading">
      <v-data-table
        :headers="headers"
        :items="pushList"
        :loading="loading"
        item-key="id"
        class="text-body-2"
        density="comfortable"
      >
        <!-- HTTP方法列 -->
        <template #item.method="{ item }">
          <v-chip
            :color="getMethodColor(item.method)"
            size="small"
            variant="flat"
            class="font-weight-bold"
          >
            {{ item.method }}
          </v-chip>
        </template>

        <!-- URL列 -->
        <template #item.url="{ item }">
          <div class="text-truncate" style="max-width: 300px;" :title="item.url">
            {{ item.url }}
          </div>
        </template>

        <!-- Body列 -->
        <template #item.body="{ item }">
          <div class="text-truncate" style="max-width: 200px;" :title="item.body">
            {{ item.body || '-' }}
          </div>
        </template>

        <!-- 备注列 -->
        <template #item.remark="{ item }">
          <div class="text-truncate" style="max-width: 150px;" :title="item.remark">
            {{ item.remark || '-' }}
          </div>
        </template>

        <!-- 启用状态列 -->
        <template #item.enable="{ item }">
          <v-switch
            :model-value="item.enable"
            color="success"
            density="compact"
            hide-details
            @update:model-value="toggleEnable(item)"
          ></v-switch>
        </template>

        <!-- 操作列 -->
        <template #item.actions="{ item }">
          <div class="d-flex ga-1">
            <v-btn
              color="primary"
              size="small"
              variant="tonal"
              icon
              @click="openEditDialog(item)"
            >
              <v-icon size="small">mdi-pencil</v-icon>
              <v-tooltip activator="parent" location="top">编辑</v-tooltip>
            </v-btn>
            <v-btn
              color="error"
              size="small"
              variant="tonal"
              icon
              @click="confirmDelete(item)"
            >
              <v-icon size="small">mdi-delete</v-icon>
              <v-tooltip activator="parent" location="top">删除</v-tooltip>
            </v-btn>
          </div>
        </template>

        <!-- 无数据提示 -->
        <template #no-data>
          <div class="text-center py-8">
            <v-icon size="64" color="grey-lighten-1" class="mb-4">mdi-bell-off</v-icon>
            <div class="text-h6 text-grey">暂无推送配置</div>
            <div class="text-body-2 text-grey-lighten-1 mb-4">点击上方按钮添加新的推送配置</div>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- 新增/编辑对话框 -->
    <v-dialog v-model="dialog" max-width="600" persistent>
      <v-card class="rounded-xl">
        <v-card-title class="d-flex align-center pa-4">
          <v-icon color="primary" class="mr-2">
            {{ isEdit ? 'mdi-pencil' : 'mdi-plus' }}
          </v-icon>
          {{ isEdit ? '编辑推送' : '新增推送' }}
        </v-card-title>
        <v-divider></v-divider>
        <v-card-text class="pa-4">
          <v-form ref="formRef" v-model="formValid">
            <v-row dense>
              <v-col cols="12" sm="4">
                <v-select
                  v-model="form.method"
                  label="HTTP方法"
                  :items="methodOptions"
                  variant="outlined"
                  density="comfortable"
                  :rules="[rules.required]"
                  prepend-inner-icon="mdi-web"
                ></v-select>
              </v-col>
              <v-col cols="12" sm="8">
                <v-text-field
                  v-model="form.url"
                  label="推送URL"
                  variant="outlined"
                  density="comfortable"
                  :rules="[rules.required, rules.url]"
                  prepend-inner-icon="mdi-link"
                  placeholder="https://example.com/webhook"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-textarea
                  v-model="form.body"
                  label="请求体 (Body)"
                  variant="outlined"
                  density="comfortable"
                  rows="4"
                  prepend-inner-icon="mdi-code-json"
                  placeholder='{"message": "{{.Message}}", "time": "{{.Time}}"}'
                  hint="支持模板变量: {{.Message}}, {{.Time}}, {{.Level}} 等"
                  persistent-hint
                ></v-textarea>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="form.remark"
                  label="备注"
                  variant="outlined"
                  density="comfortable"
                  prepend-inner-icon="mdi-note-text"
                  placeholder="推送配置描述"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-switch
                  v-model="form.enable"
                  label="启用推送"
                  color="success"
                  hide-details
                ></v-switch>
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions class="pa-4">
          <v-spacer></v-spacer>
          <v-btn variant="tonal" @click="closeDialog">取消</v-btn>
          <v-btn
            color="primary"
            variant="flat"
            @click="submitForm"
            :loading="submitLoading"
            :disabled="!formValid"
          >
            {{ isEdit ? '保存' : '创建' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 删除确认对话框 -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card class="rounded-xl">
        <v-card-title class="d-flex align-center pa-4">
          <v-icon color="error" class="mr-2">mdi-alert-circle</v-icon>
          确认删除
        </v-card-title>
        <v-divider></v-divider>
        <v-card-text class="pa-4">
          <p>确定要删除这个推送配置吗？</p>
          <p class="text-caption text-grey mt-2">
            备注: {{ deleteItem?.remark || '无备注' }}
          </p>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions class="pa-4">
          <v-spacer></v-spacer>
          <v-btn variant="tonal" @click="deleteDialog = false">取消</v-btn>
          <v-btn
            color="error"
            variant="flat"
            @click="handleDelete"
            :loading="deleteLoading"
          >
            删除
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getPushList, createPush, editPush, deletePush } from "~/src/api/push";
import type { PushItem } from "~/src/types/push/push";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const snackbarStore = useSnackbarStore();

// 数据状态
const loading = ref(false);
const pushList = ref<PushItem[]>([]);

// 对话框状态
const dialog = ref(false);
const isEdit = ref(false);
const formRef = ref();
const formValid = ref(false);
const submitLoading = ref(false);

// 删除对话框状态
const deleteDialog = ref(false);
const deleteItem = ref<PushItem | null>(null);
const deleteLoading = ref(false);

// HTTP方法选项
const methodOptions = ["GET", "POST", "PUT", "DELETE"];

// 表单数据
const defaultForm = {
  id: 0,
  method: "POST",
  url: "",
  body: "",
  remark: "",
  enable: true,
};
const form = ref({ ...defaultForm });

// 表格列定义
const headers = [
  { title: "HTTP方法", key: "method", width: "100px" },
  { title: "推送URL", key: "url" },
  { title: "请求体", key: "body" },
  { title: "备注", key: "remark" },
  { title: "启用", key: "enable", width: "80px" },
  { title: "操作", key: "actions", width: "120px", sortable: false },
];

// 验证规则
const rules = {
  required: (v: string) => !!v || "必填项",
  url: (v: string) => {
    if (!v) return true;
    try {
      new URL(v);
      return true;
    } catch {
      return "请输入有效的URL";
    }
  },
};

// 获取HTTP方法颜色
const getMethodColor = (method: string) => {
  const colors: Record<string, string> = {
    GET: "success",
    POST: "primary",
    PUT: "warning",
    DELETE: "error",
  };
  return colors[method] || "grey";
};

// 加载推送列表
const loadPushList = async () => {
  loading.value = true;
  try {
    const res = await getPushList();
    if (res.code === 0 && res.data) {
      pushList.value = res.data;
    }
  } catch (error) {
    console.error("加载推送列表失败:", error);
    snackbarStore.showErrorMessage("加载推送列表失败");
  } finally {
    loading.value = false;
  }
};

// 打开新增对话框
const openAddDialog = () => {
  isEdit.value = false;
  form.value = { ...defaultForm };
  dialog.value = true;
};

// 打开编辑对话框
const openEditDialog = (item: PushItem) => {
  isEdit.value = true;
  form.value = { ...item };
  dialog.value = true;
};

// 关闭对话框
const closeDialog = () => {
  dialog.value = false;
  form.value = { ...defaultForm };
};

// 提交表单
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
      snackbarStore.showSuccessMessage(isEdit.value ? "保存成功" : "创建成功");
      closeDialog();
      loadPushList();
    }
  } catch (error) {
    console.error("提交失败:", error);
    snackbarStore.showErrorMessage(isEdit.value ? "保存失败" : "创建失败");
  } finally {
    submitLoading.value = false;
  }
};

// 切换启用状态
const toggleEnable = async (item: PushItem) => {
  try {
    const res = await editPush({
      ...item,
      enable: !item.enable,
    });
    if (res.code === 0) {
      snackbarStore.showSuccessMessage(item.enable ? "已禁用" : "已启用");
      loadPushList();
    }
  } catch (error) {
    console.error("切换状态失败:", error);
    snackbarStore.showErrorMessage("操作失败");
  }
};

// 确认删除
const confirmDelete = (item: PushItem) => {
  deleteItem.value = item;
  deleteDialog.value = true;
};

// 执行删除
const handleDelete = async () => {
  if (!deleteItem.value) return;

  deleteLoading.value = true;
  try {
    const res = await deletePush(deleteItem.value.id);
    if (res.code === 0) {
      snackbarStore.showSuccessMessage("删除成功");
      deleteDialog.value = false;
      deleteItem.value = null;
      loadPushList();
    }
  } catch (error) {
    console.error("删除失败:", error);
    snackbarStore.showErrorMessage("删除失败");
  } finally {
    deleteLoading.value = false;
  }
};

// 初始化
onMounted(() => {
  loadPushList();
});
</script>

<style scoped>
.v-data-table :deep(th) {
  font-weight: 600 !important;
}
</style>

