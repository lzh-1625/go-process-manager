<template>
  <v-card
    variant="flat"
    height="100vh"
    class="d-flex justify-center align-center text-center"
  >
    <div>
      <img src="@/assets/images/500.svg" max-width="500" alt="500" />
      <h1 class="text-h1 pt-3">{{ t("errorPage.oops") }}</h1>
      <h4 class="text-h4 my-8">
        {{ t("errorPage.notFound") }}
      </h4>
      <v-btn flat color="#705CF6" class="mb-4 text-white" to="/"
        >{{ t("common.backToHome") }}</v-btn
      >
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
const { t } = useI18n();
</script>
