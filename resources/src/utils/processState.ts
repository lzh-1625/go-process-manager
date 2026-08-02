import { ProcessState } from "../types/process/process.ts";

export const isProcessStartDisabled = (state: ProcessState) =>
  state === ProcessState.Stopping;
