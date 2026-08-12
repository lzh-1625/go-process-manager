import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const terminalPty = readFileSync(
  new URL("./TerminalPty.vue", import.meta.url),
  "utf8"
);

test("terminal toolbar identifies the active controller", () => {
  assert.match(terminalPty, /v-if="props\.data\.controller"/);
  assert.match(terminalPty, /processCardPage\.terminalControlledBy/);
});

test("process card keeps terminal access enabled when another user controls it", () => {
  const processCard = readFileSync(
    new URL("./ProcessCard.vue", import.meta.url),
    "utf8"
  );

  assert.match(processCard, /:disabled="terminalUnavailable"/);
});
