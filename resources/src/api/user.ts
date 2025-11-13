import api from "./api";

// 登录方法
export function login(data) {
  return api.post("/user/login", {
    account: data.account,
    password: data.password,
  }).then((res) => res);
}

export function createUser(data) {
  return api.post("/user", data).then((res) => res);
}

export function deleteUser(account) {
  return api.delete("/user", { account }).then((res) => res);
}

export function editUser(data) {
  return api.put("/user", data).then((res) => res);
}

export function registerAdmin(password) {
  return api.get("/user/register/admin", { password }).then((res) => res);
}

export function getUserList() {
  return api.get("/user", undefined).then((res) => res);
}

export function getPermission(account, pid) {
  return api.get("/permission/list", { account, pid }).then((res) => res);
}



export function editPermission(data) {
  return api.put("/permission", data).then((res) => res);
}
