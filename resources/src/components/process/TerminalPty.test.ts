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

test("terminal routes WebSocket traffic through the ZMODEM sentry", () => {
  assert.match(terminalPty, /import \* as Zmodem from "zmodem\.js-ex\/src\/zmodem_browser"/);
  assert.match(terminalPty, /new Zmodem\.Sentry/);
  assert.match(terminalPty, /zmodemSentry\.consume\(event\.data\)/);
  assert.match(terminalPty, /socket\.send\(new Uint8Array\(octets\)\)/);
  assert.match(terminalPty, /try \{\s*zmodemSentry\.consume\(event\.data\)/);
});

test("terminal opens a file chooser when lrzsz rz requests an upload", () => {
  assert.match(terminalPty, /v-model="zmodemUploadDialog"/);
  assert.match(terminalPty, /Zmodem\.Browser\.send_files/);
});

test("terminal saves lrzsz sz payloads returned by the accepted offer", () => {
  assert.match(terminalPty, /offer\.accept\(\)\.then\(\(payloads\) =>/);
  assert.match(terminalPty, /Zmodem\.Browser\.save_to_disk\(payloads, fileName\)/);
});
