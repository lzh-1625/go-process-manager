import api from "./api";

export function getTaskAll() {
  return api.get("/task/all", undefined).then((res) => res);
}

export function getTaskAllWait() {
  return api.get("/task/all/wait", undefined).then((res) => res);
}

export function getTaskById(id) {
  return api.get("/task", { id }).then((res) => res);
}

export function startTaskById(id) {
  return api.get("/task/start", { id }).then((res) => res);
}

export function stopTaskById(id) {
  return api.get("/task/stop", { id }).then((res) => res);
}

export function editTaskEnable(body) {
  return api.put("/task/enable", body).then((res) => res);
}

export function editTask(body) {
  return api.put("/task", body).then((res) => res);
}

export function addTask(body) {
  return api.post("/task", body).then((res) => res);
}

export function deleteTaskById(id) {
  return api.delete("/task", { id }).then((res) => res);
}

export function changeTaskKey(id) {
  return api.post("/task/key", { id }).then((res) => res);
}
