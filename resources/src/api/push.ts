import api from "./api";

export function createPush(data) {
  return api.post("/push", data).then((res) => res);
}

export function getPushList() {
  return api.get("/push/list", undefined).then((res) => res);
}

export function deletePush(id) {
  return api.delete("/push", { id }).then((res) => res);
}

export function getPush(id) {
  return api.get("/push", { id }).then((res) => res);
}

export function editPush(data) {
  return api.put("/push", data).then((res) => res);
}
