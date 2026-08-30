import assert from "node:assert/strict";
import test from "node:test";
import * as loginAppearance from "./loginAppearance.ts";

test("login fields use dark-surface colors in dark mode", () => {
  assert.deepEqual(loginAppearance.getLoginAppearance?.(true), {
    cardColor: "#2B323B",
    fieldColor: "#2F3349",
  });
});

test("login fields retain light-surface colors in light mode", () => {
  assert.deepEqual(loginAppearance.getLoginAppearance?.(false), {
    cardColor: "#FFFFFF",
    fieldColor: "#F5F5F5",
  });
});
