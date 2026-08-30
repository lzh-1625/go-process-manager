import assert from "node:assert/strict";
import test from "node:test";
import * as chartAppearance from "./chartAppearance.ts";

test("chart tooltip uses readable colors in dark mode", () => {
  assert.deepEqual(chartAppearance.getChartTooltipStyle(true), {
    backgroundColor: "#2F3349",
    borderColor: "#5E6692",
    textStyle: { color: "#E7E9F6" },
  });
});

test("chart tooltip keeps the existing light color contrast", () => {
  assert.deepEqual(chartAppearance.getChartTooltipStyle(false), {
    backgroundColor: "#FFFFFF",
    borderColor: "#E0E0E0",
    textStyle: { color: "#2F2B3D" },
  });
});

test("log metric processing badge is compact enough for the card title", () => {
  assert.deepEqual(chartAppearance.logMetricProcessingBadgeStyle, {
    display: "inline-flex",
    alignItems: "center",
    gap: "6px",
    minHeight: "28px",
    padding: "0 10px",
    borderRadius: "999px",
    background: "rgba(255, 152, 0, 0.12)",
    border: "1px solid rgba(255, 152, 0, 0.45)",
  });
});

test("chart tooltips escape clipped cards and remain scrollable", () => {
  assert.deepEqual(chartAppearance.getChartTooltipOverflowBehavior?.(), {
    appendTo: "body",
    enterable: true,
    extraCssText: "max-height: calc(100vh - 24px); overflow-y: auto;",
  });
});
