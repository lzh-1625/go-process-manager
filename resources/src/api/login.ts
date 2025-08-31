import api from "./api";

interface LoginRes {
  code: number;
  token: string;
  username: string;
  role: number;
}

export const login = (query: any) => {
  return api.post<LoginRes>("/user/login", query).then((res) => res);
};
