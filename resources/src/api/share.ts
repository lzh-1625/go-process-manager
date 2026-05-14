import http from "./api";

export interface ShareLink {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  pid: number;
  write: boolean;
  expireTime: string;
  createBy: string;
  token: string;
}

export interface ShareListResponse {
  code: number;
  data: ShareLink[];
  msg: string;
}

export const getShareList = () => {
  return http.get<ShareLink[]>("/ws/token/list");
};

export const deleteShare = (id: number) => {
  return http.delete(`/ws/token`, { id });
};

