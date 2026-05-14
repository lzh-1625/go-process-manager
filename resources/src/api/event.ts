import api from "./api";
import type { EventListReq, EventListResp } from "../types/event/event";

export function getEventList(params: EventListReq) {
  return api.get<EventListResp>("/event", params).then((res) => res);
}

