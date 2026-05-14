export interface ProcessLog {
  id?: number;
  log: string;
  time: number;
  name: string;
  using: string;
}

export interface GetLogReq {
  filterName?: string[];
  match?: {
    log?: string;
    name?: string;
    using?: string;
  };
  time?: {
    startTime?: number;
    endTime?: number;
  };
  page?: {
    from?: number;
    size?: number;
  };
  sort?: string;
}

export interface LogResp {
  total: number;
  data: ProcessLog[];
}
