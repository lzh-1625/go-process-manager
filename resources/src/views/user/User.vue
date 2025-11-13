<template>
  <!-- 用户表格 -->
  <v-card class="pa-4">
    <v-card-title class="d-flex justify-space-between align-center">
      <span>任务</span>
      <v-btn color="primary" variant="tonal" @click="addDialog = true">
        <v-icon left>mdi-plus</v-icon> 添加用户
      </v-btn>
    </v-card-title>

    <v-data-table
      :loading="loading"
      :headers="headers"
      :items="desserts"
      class="text-body-2"
      hover
      density="comfortable"
    >
      <template #item.createTime="{ item }">
        {{ timeHanlder(item.createTime) }}
      </template>

      <template #item.role="{ item }">
        <v-chip
          :color="
            item.role === 1 ? 'primary' : item.role === 2 ? 'success' : 'grey'
          "
          variant="flat"
          class="rounded-pill"
          small
        >
          {{ getRole(item.role) }}
        </v-chip>
      </template>

      <template #item.actions="{ item }">
        <v-btn
          icon
          variant="text"
          size="small"
          :disabled="item.role != 2"
          @click="oprEdit(item)"
        >
          <v-icon color="primary">mdi-application-edit</v-icon>
        </v-btn>
        <v-btn icon variant="text" size="small" @click="editItem(item)">
          <v-icon color="warning">mdi-pencil</v-icon>
        </v-btn>
        <v-btn icon variant="text" size="small" @click="deleteItem(item)">
          <v-icon color="error">mdi-delete</v-icon>
        </v-btn>
      </template>
    </v-data-table>
  </v-card>

  <!-- 修改密码对话框 -->
  <v-dialog v-model="dialog" max-width="480">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">修改</v-card-title>
      <v-card-text>
        <v-select
          v-model="userForm.role"
          :items="items"
          item-title="label"
          label="选择用户角色"
        ></v-select>
        <v-text-field
          v-model="userForm.password"
          :append-inner-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
          :type="show1 ? 'text' : 'password'"
          label="新密码"
          hint="长度不能小于4"
          @click:append-inner="show1 = !show1"
        ></v-text-field>
        <v-text-field v-model="userForm.remark" label="备注"></v-text-field>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn variant="text" @click="close">取消</v-btn>
        <v-btn color="primary" variant="flat" @click="save">确认</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- 删除确认对话框 -->
  <v-dialog v-model="dialogDelete" max-width="420">
    <v-card>
      <v-card-title>确认删除该用户？</v-card-title>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn variant="text" @click="closeDelete">取消</v-btn>
        <v-btn color="error" variant="flat" @click="deleteItemConfirm"
          >确认</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- 添加用户对话框 -->
  <v-dialog v-model="addDialog" max-width="520">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">添加新用户</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="addUserForm.account"
          label="用户名"
        ></v-text-field>
        <v-text-field
          v-model="addUserForm.password"
          :append-inner-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
          :type="show1 ? 'text' : 'password'"
          label="密码"
          @click:append-inner="show1 = !show1"
        ></v-text-field>
        <v-select
          v-model="addUserForm.role"
          :items="items"
          item-title="label"
          label="选择用户角色"
        ></v-select>
        <v-text-field v-model="addUserForm.remark" label="备注"></v-text-field>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn variant="text" @click="addDialog = false">取消</v-btn>
        <v-btn color="primary" variant="flat" @click="add">确认</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ✅ 操作权限对话框 -->
  <v-dialog v-model="oprEditdialog" max-width="900">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">
        <v-icon color="primary" class="mr-2">mdi-shield-edit-outline</v-icon>
        <span class="text-h6 font-weight-medium">修改权限</span></v-card-title
      >
      <v-card-text class="pa-4">
        <v-data-table
          :headers="permissionHeaders"
          :items="oprList"
          class="elevation-1 rounded-lg"
          density="compact"
          hover
        >
          <template #item.actions="{ item }">
            <v-btn icon variant="text" size="small" @click="oprEditSub(item)">
              <v-icon color="primary">mdi-application-edit</v-icon>
            </v-btn>
          </template>

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
            <v-icon small :color="item[field] ? 'success' : 'grey'">{{
              item[field] ? "mdi-check" : "mdi-close"
            }}</v-icon>
          </template>
        </v-data-table>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn variant="text" @click="oprEditdialog = false">取消</v-btn>
        <v-btn color="primary" variant="flat" @click="oprEditdialog = false"
          >确认</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ✅ 修改权限对话框 -->
  <v-dialog
    v-model="oprEditdiaFormDialog"
    max-width="420"
    transition="dialog-fade-transition"
  >
    <v-card class="prime-card rounded-xl elevation-2">
      <!-- Header -->
      <div class="prime-header d-flex align-center justify-space-between">
        <v-card-title class="text-h6 font-weight-medium">
          <v-icon color="primary" class="mr-2">mdi-shield-edit-outline</v-icon>
          <span class="text-h6 font-weight-medium">修改权限</span></v-card-title
        >
        <v-btn
          icon="mdi-close"
          variant="text"
          @click="oprEditdiaFormDialog = false"
        ></v-btn>
      </div>

      <v-divider></v-divider>

      <!-- Content -->
      <v-card-text class="px-6 py-5">
        <div v-for="(label, key) in switchLabels" :key="key" class="prime-item">
          <div class="d-flex align-center justify-space-between">
            <span class="font-weight-medium text-body-1">{{ label }}</span>
            <v-switch
              color="primary"
              inset
              density="compact"
              hide-details
              v-model="permissionEditForm[key]"
            ></v-switch>
          </div>
        </div>
      </v-card-text>

      <v-divider></v-divider>

      <!-- Footer -->
      <v-card-actions class="pa-4 d-flex justify-end">
        <v-btn
          variant="text"
          color="grey-darken-1"
          @click="oprEditdiaFormDialog = false"
          >取消</v-btn
        >
        <v-btn color="primary" variant="flat" class="prime-btn" @click="submit"
          >确认</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref } from "vue";
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
  { title: "修改", key: "actions", sortable: false },
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

const editItem = (item) => {
  userForm.value = {
    account: item.account,
  };
  userForm.value.password = "";
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
</script>

<style scoped>
.v-data-table {
  border-radius: 16px;
}
.v-card-title {
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}
.v-card-actions {
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}
</style>
