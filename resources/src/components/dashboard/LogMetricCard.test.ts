import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const logMetricCard = readFileSync(
  new URL("./LogMetricCard.vue", import.meta.url),
  "utf8"
);

test("log statistics show an empty-data prompt instead of an empty chart", () => {
  assert.match(
    logMetricCard,
    /v-else[^>]*>[\s\S]*?\$t\("common\.noData"\)/
  );
});

test("log query no longer uses the no-logs-retrieved copy", () => {
  const logView = readFileSync(
    new URL("../../views/log/Log.vue", import.meta.url),
    "utf8"
  );
  const englishLocale = readFileSync(
    new URL("../../locales/en.ts", import.meta.url),
    "utf8"
  );
  const chineseLocale = readFileSync(
    new URL("../../locales/zhHans.ts", import.meta.url),
    "utf8"
  );

  assert.doesNotMatch(logView, /noLogsRetrieved/);
  assert.doesNotMatch(englishLocale, /noLogsRetrieved/);
  assert.doesNotMatch(chineseLocale, /noLogsRetrieved/);
});

test("log query leaves failed API feedback to the response interceptor", () => {
  const logView = readFileSync(
    new URL("../../views/log/Log.vue", import.meta.url),
    "utf8"
  );
  const pageLevelErrors =
    logView.match(/showErrorMessage\(t\("logPage\.loadLogsFailed"\)\)/g) || [];

  assert.equal(pageLevelErrors.length, 1);
});
