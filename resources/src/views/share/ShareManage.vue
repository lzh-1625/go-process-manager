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
          <v-icon color="primary" class="mr-2">mdi-share-variant</v-icon>
          <span class="flex-fill">{{ $t('sharePage.title') }}</span>
          <v-btn icon variant="text" size="small" @click="refreshList">
            <v-icon>mdi-refresh</v-icon>
          </v-btn>
        </h6>

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
            <tr v-for="item in paginatedShares" :key="item.id">
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
                  {{ isExpired(item.expireTime) ? $t('sharePage.expired') : timeHandler(item.expireTime) }}
                </v-chip>
              </td>
              <td>{{ item.createBy }}</td>
              <td>
                <v-chip
                  :color="item.write ? 'warning' : 'grey'"
                  size="small"
                  class="font-weight-bold"
                >
                  {{ item.write ? $t('sharePage.writable') : $t('sharePage.readonly') }}
                </v-chip>
              </td>
              <td>
                <v-tooltip :text="$t('sharePage.copyLink')">
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
                <v-tooltip :text="$t('sharePage.deleteShare')">
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
                {{ $t('sharePage.noShares') }}
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
            {{ $t('sharePage.totalShares', { n: shareList.length }) }}
          </div>
        </div>
      </div>
    </v-card>
  </v-container>

  <v-dialog v-model="dialogDelete" max-width="480">
    <v-card class="rounded-xl">
      <v-card-title class="text-h6 font-weight-medium">{{ $t('sharePage.confirmDelete') }}</v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        {{ $t('sharePage.confirmDeleteMsg') }}
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions class="justify-end pa-4">
        <v-btn text @click="closeDelete">{{ $t('common.cancel') }}</v-btn>
        <v-btn color="error" @click="deleteItemConfirm">{{ $t('sharePage.confirmDeleteBtn') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { getShareList, deleteShare } from "~/src/api/share";
import { getProcessList } from "~/src/api/process";
import { useSnackbarStore } from "~/src/stores/snackbarStore";

const { t } = useI18n();
const snackbarStore = useSnackbarStore();

const loading = ref(false);
const dialogDelete = ref(false);
const shareList = ref([]);
const processList = ref([]);
const deleteId = ref(0);

const currentPage = ref(1);
const pageSize = ref(10);

const headers = computed(() => [
  { title: "ID", key: "id" },
  { title: t("sharePage.process"), key: "pid" },
  { title: t("sharePage.createTime"), key: "createdAt" },
  { title: t("sharePage.lastUsed"), key: "lastLink" },
  { title: t("sharePage.expireTime"), key: "expireTime" },
  { title: t("sharePage.creator"), key: "createBy" },
  { title: t("sharePage.permission"), key: "write" },
  { title: t("common.operation"), key: "actions", sortable: false },
]);

const totalPages = computed(() => {
  return Math.ceil(shareList.value.length / pageSize.value);
});

const paginatedShares = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return shareList.value.slice(start, end);
});

const handlePageChange = (page) => {
  currentPage.value = page;
};

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

const isExpired = (expireTime) => {
  return new Date(expireTime) < new Date();
};

const getProcessName = (pid) => {
  const process = processList.value.find((p) => p.uuid === pid);
  return process ? process.name : `PID: ${pid}`;
};

const copyShareLink = (token) => {
  const shareUrl = `${window.location.origin}/share?token=${token}`;
  navigator.clipboard.writeText(shareUrl).then(
    () => {
      snackbarStore.showSuccessMessage(t("sharePage.copySuccess"));
    },
    (err) => {
      snackbarStore.showErrorMessage(t("sharePage.copyFailed"));
    }
  );
};

const deleteItem = (item) => {
  dialogDelete.value = true;
  deleteId.value = item.id;
};

const deleteItemConfirm = () => {
  dialogDelete.value = false;
  deleteShare(deleteId.value).then((resp) => {
    if (resp.code === 0) {
      initialize();
      snackbarStore.showSuccessMessage(t("common.deleteSuccess"));
    }
  });
};

const closeDelete = () => (dialogDelete.value = false);

const initialize = () => {
  loading.value = true;

  Promise.all([getProcessList(), getShareList()])
    .then(([processResp, shareResp]) => {
      processList.value = processResp.data || [];
      shareList.value = shareResp.data || [];
      loading.value = false;
    })
    .catch((err) => {
      loading.value = false;
      snackbarStore.showErrorMessage(t("sharePage.loadFailed"));
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
