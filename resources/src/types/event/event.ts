// 事件类型枚举
export type EventType =
  | "ProcessStart"
  | "ProcessStop"
  | "ProcessWarning"
  | "ProcessConnect"
  | "TaskStart"
  | "TaskStop"
  | "ApiRequest"
  | "PushRequest";

// 事件模型
export interface Event {
  id: number;
  name: string;
  user: string;
  type: EventType;
  additional: string;
  createdTime: string;
}

// 事件列表请求参数
export interface EventListReq {
  page?: number;
  size?: number;
  startTime?: number;
  endTime?: number;
  type?: EventType;
  name?: string;
  user?: string;
  keyValue?: string;
}

// 事件列表响应
export interface EventListResp {
  total: number;
  data: Event[];
}
