import api from "./api";

export interface PerformceUsageItem {
  name: string;
  cpu: number;
  mem: number;
}

export interface PerformceUsage {
  items: PerformceUsageItem[];
  cpuFree: number;
  memFree: number;
}

export interface LogStatsticMetricItem {
  count: number;
  date: string;
}

export interface LogStatsticMetric {
  executing: number;
  items: LogStatsticMetricItem[];
}

export function getPerformceUsage() {
  return api.get<PerformceUsage>("/metric/performce").then((res) => res);
}

export function getLogMetric(dateType: number) {
  return api.get<LogStatsticMetric>("/metric/log", { dateType }).then((res) => res);
}

