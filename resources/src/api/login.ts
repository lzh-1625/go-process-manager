import api from "./api";

export const login = (query: any) => {
  return api.post("/user/login", query).then((res) => res);
};
