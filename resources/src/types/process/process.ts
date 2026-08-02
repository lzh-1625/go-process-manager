export interface ProcessItem {
  name: string;
  uuid: number;
  startTime: Date;
  user: string;
  controller: string;
  controlExpiredTime: string;
  usage: Usage;
  state: State;
  termType: TermType;
  cgroupEnable: boolean;
  memoryLimit: number | null;
  cpuLimit: number | null;
}

export interface State {
  state: ProcessState;
  info: string;
}

export enum ProcessState {
  Stopped = 0,
  Starting = 1,
  Warning = 2,
  Running = 3,
  Stopping = 4,
}

export enum TermType {
  Pty = "pty",
}

export interface Usage {
  cpuCapacity: number;
  memCapacity: number;
  cpu: number[] | null;
  mem: number[] | null;
  time: string[] | null;
}

export interface ProcessConfig {
  uuid: number;
  name: string;
  cmd: string;
  cwd: string;
  autoRestart: boolean;
  compulsoryRestart: boolean;
  pushIds: number[] | string;
  logReport: boolean;
  termType: string;
  cgroupEnable: boolean;
  memoryLimit: null;
  cpuLimit: null;
  env: string;
}
