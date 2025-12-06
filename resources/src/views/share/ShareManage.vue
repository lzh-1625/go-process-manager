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
          <v-icon color="primary" class="mr-2">mdi-share-variant</v-icon>
          <span class="flex-fill">分享链接管理</span>
          <v-btn icon variant="text" size="small" @click="refreshList">
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
        </h6>

        <!-- 分享链接列表 -->
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
            <tr v-for="item in paginatedShares" :key="item.ID">
              <td class="font-weight-bold">{{ item.id }}</td>
              <td>
                <v-chip
                  color="primary"
                  size="small"
                  class="font-weight-bold"
                >
                  {{ getProcessName(item.pid) }}
                </v-chip>
              </td>
              <td>{{ timeHandler(item.createdAt) }}</td>
              <td>{{ timeHandler(item.lastLink) }}</td>
              <td>
                <v-chip
                  :color="isExpired(item.expireTime) ? 'error' : 'success'"
                  size="small"
                  class="font-weight-bold"
                >
                  {{ isExpired(item.expireTime) ? '已过期' : timeHandler(item.expireTime) }}
                </v-chip>
              </td>
              <td>{{ item.createBy }}</td>
              <td>
                <v-chip
                  :color="item.write ? 'warning' : 'grey'"
                  size="small"
                  class="font-weight-bold"
                >
                  {{ item.write ? '可写' : '只读' }}
                </v-chip>
              </td>
              <td>
                <v-tooltip text="复制分享链接">
                  <template v-slot:activator="{ props }">
                    <v-btn
                      v-bind="props"
                      icon
                      variant="text"
                      size="small"
                      @click="copyShareLink(item.token)"
                    >
                      <v-icon color="primary">mdi-content-copy</v-icon>
                    </v-btn>
                  </template>
                </v-tooltip>
                <v-tooltip text="删除分享">
                  <template v-slot:activator="{ props }">
                    <v-btn
                      v-bind="props"
                      icon
                      variant="text"
                      size="small"
                      @click="deleteItem(item)"
                    >
                      <v-icon color="error">mdi-delete</v-icon>
                    </v-btn>
                  </template>
                </v-tooltip>
              </td>
            </tr>
            <tr v-if="shareList.length === 0">
              <td colspan="7" class="text-center text-secondary pa-8">
                暂无分享链接
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
            共 {{ shareList.length }} 个分享链接
          </div>
        </div>
      </div>
    </v-card>
  </v-container>

  <!-- 删除确认对话框 -->
  <v-dialog v-model="dialogDelete" max-width="480">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">确认删除</v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        确认删除该分享链接吗？此操作无法撤销。
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="closeDelete">取消</v-btn>
        <v-btn color="error" @click="deleteItemConfirm">确认删除</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { getShareList, deleteShare } from "~/src/api/share";
import { getProcessList } from "~/src/api/process";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const snackbarStore = useSnackbarStore();

const loading = ref(false);
const dialogDelete = ref(false);
const shareList = ref([]);
const processList = ref([]);
const deleteId = ref(0);

// 分页
const currentPage = ref(1);
const pageSize = ref(10);

const headers = [
  { title: "ID", key: "id" },
  { title: "进程", key: "pid" },
  { title: "创建时间", key: "createdAt" },
  { title: "最后使用", key: "lastLink" },
  { title: "过期时间", key: "expireTime" },
  { title: "创建者", key: "createBy" },
  { title: "权限", key: "write" },
  { title: "操作", key: "actions", sortable: false },
];

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(shareList.value.length / pageSize.value);
});

// 计算当前页数据
const paginatedShares = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return shareList.value.slice(start, end);
});

// 处理页码变化
const handlePageChange = (page) => {
  currentPage.value = page;
};

// 时间格式化
const timeHandler = (t) => {
  if (!t) return "-";
  return new Date(t).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
};

// 判断是否过期
const isExpired = (expireTime) => {
  return new Date(expireTime) < new Date();
};

// 根据pid获取进程名称
const getProcessName = (pid) => {
  const process = processList.value.find((p) => p.uuid === pid);
  return process ? process.name : `PID: ${pid}`;
};

// 复制分享链接
const copyShareLink = (token) => {
  const shareUrl = `${window.location.origin}/share?token=${token}`;
  navigator.clipboard.writeText(shareUrl).then(
    () => {
      snackbarStore.showSuccessMessage("分享链接已复制到剪贴板");
    },
    (err) => {
      snackbarStore.showErrorMessage("复制失败: " + err);
    }
  );
};

// 删除分享
const deleteItem = (item) => {
  dialogDelete.value = true;
  deleteId.value = item.ID;
};

const deleteItemConfirm = () => {
  dialogDelete.value = false;
  deleteShare(deleteId.value).then((resp) => {
    if (resp.code === 0) {
      initialize();
      snackbarStore.showSuccessMessage("删除成功");
    }
  });
};

const closeDelete = () => (dialogDelete.value = false);

const initialize = () => {
  loading.value = true;

  // 并行加载进程列表和分享列表
  Promise.all([getProcessList(), getShareList()])
    .then(([processResp, shareResp]) => {
      processList.value = processResp.data || [];
      shareList.value = shareResp.data || [];
      loading.value = false;
    })
    .catch((err) => {
      loading.value = false;
      snackbarStore.showErrorMessage("加载数据失败");
    });
};

const refreshList = () => {
  initialize();
};

onMounted(() => {
  initialize();
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

.v-card-title {
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}

.v-card-actions {
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}
</style>

