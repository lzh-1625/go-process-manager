import api from "./api";

export function getConfig() {
  return api.get("/config", undefined).then((res) => res);
}

export function setConfig(data) {
  return api.put("/config", data).then((res) => res);
}

export function configReload() {
  return api.put("/config/reload").then((res) => res);
}
