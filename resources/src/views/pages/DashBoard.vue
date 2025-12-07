<script setup lang="ts">
import PerformanceCard from "@/components/dashboard/PerformanceCard.vue";
import LogMetricCard from "@/components/dashboard/LogMetricCard.vue";
</script>
<template>
  <div class="pa-5">
    <!-- ---------------------------------------------- -->
    <!---First Row - 系统性能监控 -->
    <!-- ---------------------------------------------- -->
    <v-row dense>
      <v-col cols="12">
        <v-card class="card-shadow" height="420">
          <performance-card></performance-card>
        </v-card>
      </v-col>
    </v-row>
    <!-- ---------------------------------------------- -->
    <!---Second Row - 日志统计 -->
    <!-- ---------------------------------------------- -->
    <v-row dense>
      <v-col cols="12">
        <v-card class="card-shadow" height="480">
          <log-metric-card></log-metric-card>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>
<style lang="scss" scoped></style>
