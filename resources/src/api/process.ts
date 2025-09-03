import { ProcessItem } from "../types/process/process";
import api from "./api";

export function getProcessList() {
  return api.get<ProcessItem[]>("/process", undefined).then((res) => res);
}

export function getProcessListWait() {
  return api.get<ProcessItem[]>("/process/wait", undefined).then((res) => res);
}

export function killProcessAll(uuid) {
  return api.delete("/process/all", { uuid }).then((res) => res);
}

export function startProcessAll(uuid) {
  return api.put("/process/all", { uuid }).then((res) => res);
}

export function killProcess(uuid) {
  return api.delete("/process", { uuid }).then((res) => res);
}

export function startProcess(uuid) {
  return api.put("/process", { uuid }).then((res) => res);
}

export function getContorl(uuid) {
  return api.get("/process/control", { uuid }).then((res) => res);
}

export function getProcessConfig(uuid) {
  return api.get("/process/config", { uuid }).then((res) => res);
}

export function deleteProcessConfig(uuid) {
  return api.delete("/process/config", { uuid }).then((res) => res);
}

export function putProcessConfig(data) {
  return api.put("/process/config", data).then((res) => res);
}

export function postProcessConfig(data) {
  return api.post("/process/config", data).then((res) => res);
}

export function createProcessShare(data) {
  return api.post("/process/share", data).then((res) => res);
}
