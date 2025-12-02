import api from "./api";
import type { GetLogReq, LogResp } from "../types/log/log";

export function getLog(data: GetLogReq) {
  return api.post<LogResp>("/log", data).then((res) => res);
}

export function getRunningLog() {
  return api.get<number>("/log/running", undefined).then((res) => res);
}
