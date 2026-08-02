import {
  ProcessState,
  type ProcessItem,
} from "../types/process/process.ts";

type ProcessChangeBase = Pick<ProcessItem, "uuid" | "name">;

export type ProcessChange =
  | (ProcessChangeBase & { type: "created" })
  | (ProcessChangeBase & { type: "deleted" })
  | (ProcessChangeBase & { type: "started" })
  | (ProcessChangeBase & { type: "stopping" })
  | (ProcessChangeBase & { type: "stopped" })
  | (ProcessChangeBase & { type: "warning"; info: string });

const processChange = (
  type: Exclude<ProcessChange["type"], "warning">,
  process: ProcessItem,
): ProcessChange => ({ type, uuid: process.uuid, name: process.name });

export const detectProcessChanges = (
  previous: ProcessItem[],
  current: ProcessItem[],
): ProcessChange[] => {
  const changes: ProcessChange[] = [];
  const previousById = new Map(previous.map((process) => [process.uuid, process]));
  const currentById = new Map(current.map((process) => [process.uuid, process]));

  for (const process of current) {
    if (!previousById.has(process.uuid)) {
      changes.push(processChange("created", process));
      if (process.state.state === ProcessState.Warning) {
        changes.push({
          type: "warning",
          uuid: process.uuid,
          name: process.name,
          info: process.state.info,
        });
      }
    }
  }

  for (const process of previous) {
    if (!currentById.has(process.uuid)) {
      changes.push(processChange("deleted", process));
    }
  }

  for (const process of current) {
    const oldProcess = previousById.get(process.uuid);
    if (!oldProcess || oldProcess.state.state === process.state.state) {
      if (
        oldProcess?.state.state === ProcessState.Warning &&
        process.state.state === ProcessState.Warning &&
        oldProcess.state.info !== process.state.info
      ) {
        changes.push({
          type: "warning",
          uuid: process.uuid,
          name: process.name,
          info: process.state.info,
        });
      }
      continue;
    }

    if (process.state.state === ProcessState.Warning) {
      changes.push({
        type: "warning",
        uuid: process.uuid,
        name: process.name,
        info: process.state.info,
      });
    } else if (process.state.state === ProcessState.Running) {
      changes.push(processChange("started", process));
    } else if (process.state.state === ProcessState.Stopping) {
      changes.push(processChange("stopping", process));
    } else if (process.state.state === ProcessState.Stopped) {
      changes.push(processChange("stopped", process));
    }
  }

  return changes;
};
