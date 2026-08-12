export const CONTROL_DURATION_DEFAULT = 60;
export const CONTROL_DURATION_MIN = 1;
export const CONTROL_DURATION_MAX = 3600;

export function isControlDuration(value: unknown): value is number {
  return (
    typeof value === "number" &&
    Number.isInteger(value) &&
    value >= CONTROL_DURATION_MIN &&
    value <= CONTROL_DURATION_MAX
  );
}
