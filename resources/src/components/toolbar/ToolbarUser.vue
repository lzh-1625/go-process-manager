<!--
* @Component: ToolbarNotifications
* @Maintainer: J.K. Yang
* @Description:
-->
<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
const router = useRouter();

const userName = computed(() => localStorage.getItem("name") || "User");

const handleLogout = () => {
  localStorage.removeItem("token");
  router.push("/login");
};

const openGithub = () => {
  window.open("https://github.com/lzh-1625/go_process_manager");
};

</script>

<template>
  <v-menu
    :close-on-content-click="false"
    location="bottom right"
    transition="slide-y-transition"
  >
    <!-- ---------------------------------------------- -->
    <!-- Activator Btn -->
    <!-- ---------------------------------------------- -->
    <template v-slot:activator="{ props }">
      <v-btn class="mx-2" icon v-bind="props">
          <v-avatar size="40">
            <v-icon>mdi-account-circle</v-icon>
          </v-avatar>
      </v-btn>
    </template>
    <v-card max-width="300">
      <v-list lines="three" density="compact">
        <!-- ---------------------------------------------- -->
        <!-- Profile Area -->
        <!-- ---------------------------------------------- -->
        <v-list-item>

          <v-list-item-title class="font-weight-bold text-primary">
          {{ userName }}
          </v-list-item-title>
        </v-list-item>
      </v-list>
      <v-divider />
      <!-- ---------------------------------------------- -->
      <!-- Logout Area -->
      <!-- ---------------------------------------------- -->
      <v-list variant="flat" elevation="0" :lines="false" density="compact">
        <v-list-item v-permission="0" color="primary" @click="openGithub" link density="compact" >
          <template v-slot:prepend>
            <v-avatar size="30">
              <v-icon>mdi-github</v-icon>
            </v-avatar>
          </template>

          <div>
            <v-list-item-subtitle class="text-body-2">
              Github
            </v-list-item-subtitle>
          </div>
        </v-list-item>
        <v-list-item
          color="primary"
          link
          @click="handleLogout"
          density="compact"
        >
          <template v-slot:prepend>
            <v-avatar size="30">
              <v-icon>mdi-logout</v-icon>
            </v-avatar>
          </template>

          <div>
            <v-list-item-subtitle class="text-body-2">
              Logout
            </v-list-item-subtitle>
          </div>
        </v-list-item>
      </v-list>
    </v-card>
  </v-menu>
</template>

<style scoped lang="scss">
// ::v-deep .v-list-item__append,
// ::v-deep .v-list-item__prepend {
//   height: 100%;
// }
</style>
