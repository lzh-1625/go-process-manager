import assert from "node:assert/strict";
import test from "node:test";
import { CONTROL_DURATION_DEFAULT, isControlDuration } from "./controlDuration.ts";

test("control duration accepts whole seconds only within 1 through 3600", () => {
  assert.equal(CONTROL_DURATION_DEFAULT, 60);
  assert.equal(isControlDuration(1), true);
  assert.equal(isControlDuration(60), true);
  assert.equal(isControlDuration(3600), true);
  assert.equal(isControlDuration(0), false);
  assert.equal(isControlDuration(3601), false);
  assert.equal(isControlDuration(1.5), false);
  assert.equal(isControlDuration("60"), false);
});
