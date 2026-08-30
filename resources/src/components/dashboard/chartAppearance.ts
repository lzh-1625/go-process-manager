export function getChartTooltipStyle(dark: boolean) {
  return dark
    ? {
        backgroundColor: "#2F3349",
        borderColor: "#5E6692",
        textStyle: { color: "#E7E9F6" },
      }
    : {
        backgroundColor: "#FFFFFF",
        borderColor: "#E0E0E0",
        textStyle: { color: "#2F2B3D" },
      };
}

export function getChartTooltipOverflowBehavior() {
  return {
    appendTo: "body",
    enterable: true,
    extraCssText: "max-height: calc(100vh - 24px); overflow-y: auto;",
  };
}

export const logMetricProcessingBadgeStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "6px",
  minHeight: "28px",
  padding: "0 10px",
  borderRadius: "999px",
  background: "rgba(255, 152, 0, 0.12)",
  border: "1px solid rgba(255, 152, 0, 0.45)",
} as const;
