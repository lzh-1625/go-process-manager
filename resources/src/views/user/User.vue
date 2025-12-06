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
          <v-icon color="primary" class="mr-2">mdi-account-multiple</v-icon>
          <span class="flex-fill">用户管理</span>
          <v-btn icon variant="text" size="small" @click="refreshUsers">
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
          <v-btn
            color="primary"
            variant="tonal"
            size="small"
            @click="addDialog = true"
          >
            <v-icon left>mdi-plus</v-icon>
            添加用户
          </v-btn>
        </h6>

        <!-- 用户列表 -->
        <v-table class="pa-3">
          <thead>
            <tr>
              <th
                class="text-left"
                v-for="header in headers"
                :key="header.title"
              >
                {{ header.title }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in paginatedUsers" :key="item.account">
              <td class="font-weight-bold">{{ item.account }}</td>
              <td>
                <v-chip
                  :color="
                    item.role === 1
                      ? 'primary'
                      : item.role === 2
                      ? 'success'
                      : 'grey'
                  "
                  size="small"
                  class="font-weight-bold"
                >
                  {{ getRole(item.role) }}
                </v-chip>
              </td>
              <td>{{ timeHanlder(item.createTime) }}</td>
              <td>{{ item.remark || "-" }}</td>
              <td>
                <v-btn
                  icon
                  variant="text"
                  size="small"
                  :disabled="item.role != 2"
                  @click="oprEdit(item)"
                >
                  <v-icon color="primary">mdi-application-edit</v-icon>
                </v-btn>
                <v-btn
                  icon
                  variant="text"
                  size="small"
                  @click="editItem(item)"
                  :disabled="item.role == 0"
                >
                  <v-icon color="warning">mdi-pencil</v-icon>
                </v-btn>
                <v-btn
                  icon
                  variant="text"
                  size="small"
                  @click="deleteItem(item)"
                  :disabled="item.role == 0"
                >
                  <v-icon color="error">mdi-delete</v-icon>
                </v-btn>
              </td>
            </tr>
            <tr v-if="desserts.length === 0">
              <td colspan="5" class="text-center text-secondary pa-8">
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
            共 {{ desserts.length }} 个用户
          </div>
        </div>
      </div>
    </v-card>
  </v-container>

  <!-- 修改密码对话框 -->
  <v-dialog v-model="dialog" max-width="520">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">修改用户</v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-container fluid>
          <v-row dense>
            <v-col cols="12">
              <v-select
                v-model="userForm.role"
                :items="items"
                item-title="label"
                label="选择用户角色"
                variant="outlined"
                density="comfortable"
              ></v-select>
            </v-col>
            <v-col cols="12">
              <v-text-field
                v-model="userForm.password"
                :append-inner-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
                :type="show1 ? 'text' : 'password'"
                label="新密码"
                hint="长度不能小于4"
                variant="outlined"
                density="comfortable"
                @click:append-inner="show1 = !show1"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                v-model="userForm.remark"
                label="备注"
                variant="outlined"
                density="comfortable"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="close">取消</v-btn>
        <v-btn color="primary" @click="save">确认</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- 删除确认对话框 -->
  <v-dialog v-model="dialogDelete" max-width="480">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">确认删除</v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        确认删除该用户吗？此操作无法撤销。
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="closeDelete">取消</v-btn>
        <v-btn color="error" @click="deleteItemConfirm">确认删除</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- 添加用户对话框 -->
  <v-dialog v-model="addDialog" max-width="520">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">添加新用户</v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-container fluid>
          <v-row dense>
            <v-col cols="12">
              <v-text-field
                v-model="addUserForm.account"
                label="用户名"
                variant="outlined"
                density="comfortable"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                v-model="addUserForm.password"
                :append-inner-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
                :type="show1 ? 'text' : 'password'"
                label="密码"
                variant="outlined"
                density="comfortable"
                @click:append-inner="show1 = !show1"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-select
                v-model="addUserForm.role"
                :items="items"
                item-title="label"
                label="选择用户角色"
                variant="outlined"
                density="comfortable"
              ></v-select>
            </v-col>
            <v-col cols="12">
              <v-text-field
                v-model="addUserForm.remark"
                label="备注"
                variant="outlined"
                density="comfortable"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="addDialog = false">取消</v-btn>
        <v-btn color="primary" @click="add">确认</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ✅ 操作权限对话框 -->
  <v-dialog v-model="oprEditdialog" max-width="900">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">
        <v-icon color="primary" class="mr-2">mdi-shield-edit-outline</v-icon>
        <span>修改权限</span>
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-data-table
          :headers="permissionHeaders"
          :items="oprList"
          class="elevation-1 rounded-lg"
          density="compact"
          hover
        >
          <template
            v-for="field in [
              'owned',
              'start',
              'stop',
              'terminal',
              'write',
              'log',
            ]"
            #[`item.${field}`]="{ item }"
          >
            <v-switch
              color="primary"
              inset
              density="compact"
              hide-details
              :model-value="item[field]"
              @update:model-value="updatePermission(item, field, $event)"
            ></v-switch>
          </template>
        </v-data-table>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="oprEditdialog = false">关闭</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ✅ 修改权限对话框 -->
  <v-dialog
    v-model="oprEditdiaFormDialog"
    max-width="520"
    transition="dialog-fade-transition"
  >
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">
        <v-icon color="primary" class="mr-2">mdi-shield-edit-outline</v-icon>
        <span>修改权限</span>
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-container fluid>
          <v-row dense>
            <v-col cols="12" v-for="(label, key) in switchLabels" :key="key">
              <div class="d-flex align-center justify-space-between py-2">
                <span class="font-weight-medium text-body-1">{{ label }}</span>
                <v-switch
                  color="primary"
                  inset
                  density="compact"
                  hide-details
                  v-model="permissionEditForm[key]"
                ></v-switch>
              </div>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="oprEditdiaFormDialog = false">取消</v-btn>
        <v-btn color="primary" @click="submit">确认</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed } from "vue";
import {
  login,
  createUser,
  deleteUser,
  editUser,
  registerAdmin,
  getUserList,
  getPermission,
  editPermission,
} from "~/src/api/user";
import { useSnackbarStore } from "~/src/stores/snackbarStore";
const snackbarStore = useSnackbarStore();

const loading = ref(false);
const dialog = ref(false);
const addDialog = ref(false);
const dialogDelete = ref(false);
const show1 = ref(false);
const account = ref("");
const uuid = ref("");
const oprEditdialog = ref(false);
const permissionEditForm = ref({});
const oprEditdiaFormDialog = ref(false);

// 分页
const currentPage = ref(1);
const pageSize = ref(10);

const headers = [
  { title: "用户名", key: "account" },
  { title: "角色", key: "role" },
  { title: "创建时间", key: "createTime" },
  { title: "备注", key: "remark" },
  { title: "操作", key: "actions", sortable: false },
];
const permissionHeaders = [
  { title: "进程id", key: "pid", sortable: true },
  {
    title: "进程名",
    sortable: true,
    key: "name",
  },
  { title: "拥有", key: "owned", sortable: false },
  { title: "启动", key: "start", sortable: false },
  { title: "停止", key: "stop", sortable: false },
  { title: "终端", key: "terminal", sortable: false },
  { title: "写入", key: "write", sortable: false },
  { title: "日志", key: "log", sortable: false },
];
const desserts = ref([]);
const search = ref("");
const oprList = ref([]);
const switchLabels = {
  owned: "拥有",
  start: "启动",
  stop: "停止",
  terminal: "终端",
  write: "写入",
  log: "日志",
};

const rules = {
  required: (v) => !!v || "必填项",
  min: (v) => v.length >= 4 || "至少4个字符",
};

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(desserts.value.length / pageSize.value);
});

// 计算当前页数据
const paginatedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return desserts.value.slice(start, end);
});

// 处理页码变化
const handlePageChange = (page) => {
  currentPage.value = page;
};

onMounted(() => {
  initialize();
});
const addUserForm = ref({ account: "", password: "", role: null, remark: "" });
const userForm = ref({ account: "", password: "" });
const items = [
  { label: "管理员", value: 1 },
  { label: "普通用户", value: 2 },
];

const timeHanlder = (t) => new Date(t).toLocaleString();

const roleMap = {
  0: "root",
  1: "admin",
  2: "user",
};

const getRole = (r) => roleMap[r];

const close = () => (dialog.value = false);
const save = () => {
  dialog.value = false;
  editUser(userForm.value).then((resp) => {
    if (resp.code == 0) {
      initialize();
      snackbarStore.showSuccessMessage("操作成功");
    }
  });
};
const add = () => {
  addDialog.value = false;
  createUser(addUserForm.value).then((resp) => {
    if (resp.code == 0) {
      initialize();
      snackbarStore.showSuccessMessage("操作成功");
    }
  });
};
const deleteItem = (item) => {
  dialogDelete.value = true;
  account.value = item.account;
};
const deleteItemConfirm = () => {
  dialogDelete.value = false;
  deleteUser(account.value).then((resp) => {
    if (resp.code === 0) {
      initialize();
      snackbarStore.showSuccessMessage("操作成功");
    }
  });
};
const closeDelete = () => (dialogDelete.value = false);
const initialize = () => {
  loading.value = true;
  getUserList().then((resp) => {
    desserts.value = resp.data;
    loading.value = false;
  });
};

const refreshUsers = () => {
  initialize();
};

const editItem = (item) => {
  userForm.value = {
    account: item.account,
  };
  userForm.value.password = "";
  userForm.value.role = item.role;
  userForm.value.remark = item.remark;
  dialog.value = true;
};

const oprEdit = (e) => {
  account.value = e.account;
  uuid.value = e.uuid;
  oprEditdialog.value = true;
  getPermission(account.value, e.uuid).then((resp) => {
    if (resp.code != 0) {
      return;
    }
    oprList.value = resp.data;
  });
};

const oprEditSub = (e) => {
  permissionEditForm.value = { ...e };
  oprEditdiaFormDialog.value = true;
};

const submit = () => {
  oprEditdiaFormDialog.value = false;
  permissionEditForm.value.account = account.value;
  editPermission(permissionEditForm.value).then((resp) => {
    if (resp.code == 0) {
      snackbarStore.showSuccessMessage("操作成功");
      oprEdit({
        account: account.value,
        uuid: uuid.value,
      });
    }
  });
};

const updatePermission = (item, field, newValue) => {
  const updatedPermission = {
    ...item,
    [field]: newValue,
    account: account.value,
  };
  editPermission(updatedPermission).then((resp) => {
    if (resp.code == 0) {
      snackbarStore.showSuccessMessage("权限修改成功");
      // 更新本地数据
      item[field] = newValue;
    } else {
      // 如果失败，刷新数据恢复原状
      oprEdit({
        account: account.value,
        uuid: uuid.value,
      });
    }
  });
};
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

.v-card-title {
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}

.v-card-actions {
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}
</style>
