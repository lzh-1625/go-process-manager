<template>
  <v-container fluid class="py-6 px-8">
    <v-card class="rounded-lg">
      <!-- loading spinner -->
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
        <!-- 标题栏 -->
        <h6 class="text-h6 font-weight-bold pa-5 d-flex align-center">
          <v-icon color="primary" class="mr-2">mdi-bell-ring</v-icon>
          <span class="flex-fill">推送管理</span>
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
            新增推送
          </v-btn>
        </h6>

        <!-- 推送列表 -->
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
                暂无数据
              </td>
            </tr>
          </tbody>
        </v-table>

        <!-- 分页 -->
        <div class="text-center pa-4">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            :total-visible="7"
            density="compact"
            @update:model-value="handlePageChange"
          ></v-pagination>
          <div class="mt-2 text-caption text-secondary">
            共 {{ pushList.length }} 条推送配置
          </div>
        </div>
      </div>
    </v-card>

    <!-- 新增/编辑对话框 -->
    <v-dialog v-model="dialog" max-width="600px">
      <v-card class="rounded-xl">
        <v-card-title class="text-h6 font-weight-medium">
          {{ isEdit ? '编辑推送' : '新增推送' }}
        </v-card-title>

        <v-divider></v-divider>

        <v-card-text class="pt-6">
          <v-form ref="formRef" v-model="formValid">
            <v-container fluid>
              <v-row dense>
                <v-col cols="12" sm="4">
                  <v-select
                    v-model="form.method"
                    label="HTTP方法"
                    :items="methodOptions"
                    variant="outlined"
                    density="comfortable"
                    :rules="[rules.required]"
                  ></v-select>
                </v-col>
                <v-col cols="12" sm="8">
                  <v-text-field
                    v-model="form.url"
                    label="推送URL"
                    variant="outlined"
                    density="comfortable"
                    :rules="[rules.required, rules.url]"
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
                    placeholder='{"message": "{$message}", "user": "{$user}"}'
                    hint="占位符：{$name}进程名称 {$user}使用者 {$message}消息内容 {$status}进程状态"
                    persistent-hint
                  ></v-textarea>
                </v-col>
                <v-col cols="12">
                  <v-text-field
                    v-model="form.remark"
                    label="备注"
                    variant="outlined"
                    density="comfortable"
                    placeholder="推送配置描述"
                  ></v-text-field>
                </v-col>
                <v-col cols="12">
                  <v-switch
                    v-model="form.enable"
                    label="启用推送"
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
          <v-btn text @click="closeDialog">取消</v-btn>
          <v-btn
            color="primary"
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
    <v-dialog v-model="deleteDialog" max-width="480">
      <v-card class="rounded-xl">
        <v-card-title class="text-h6 font-weight-medium">确认删除</v-card-title>

        <v-divider></v-divider>

        <v-card-text class="pt-6">
          确定要删除这个推送配置吗？
          <div class="text-caption text-secondary mt-2">
            备注: {{ deleteItem?.remark || '无备注' }}
          </div>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions class="justify-end pa-4">
          <v-btn text @click="deleteDialog = false">取消</v-btn>
          <v-btn
            color="error"
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
import { ref, onMounted, computed } from "vue";
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

// 分页
const currentPage = ref(1);
const pageSize = ref(10);

// HTTP方法选项
const methodOptions = ["GET", "POST"];

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
  { title: "HTTP方法", key: "method" },
  { title: "推送URL", key: "url" },
  { title: "请求体", key: "body" },
  { title: "备注", key: "remark" },
  { title: "启用", key: "enable" },
  { title: "操作", key: "actions" },
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

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(pushList.value.length / pageSize.value);
});

// 计算当前页数据
const paginatedPushList = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return pushList.value.slice(start, end);
});

// 处理页码变化
const handlePageChange = (page: number) => {
  currentPage.value = page;
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

// 刷新推送列表
const refreshPushList = () => {
  loadPushList();
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
