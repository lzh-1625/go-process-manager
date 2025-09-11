export interface TaskItem {
  id: number;
  processId: number;
  condition: number;
  nextId: null;
  operation: number;
  triggerEvent: null;
  triggerTarget: null;
  operationTarget: number;
  cron: string;
  enable: boolean;
  apiEnable: boolean;
  key: string;
  processName: string;
  targetName: string;
  triggerName: string;
  startTime: Date;
  running: boolean;
}
