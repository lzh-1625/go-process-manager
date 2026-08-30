export function getLoginAppearance(dark: boolean) {
  return dark
    ? {
        cardColor: "#2B323B",
        fieldColor: "#2F3349",
      }
    : {
        cardColor: "#FFFFFF",
        fieldColor: "#F5F5F5",
      };
}
