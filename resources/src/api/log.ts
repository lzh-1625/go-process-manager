import api from "./api";

export function getLog(data: any) {
  return api.post("/log", data).then((res) => res);
}
