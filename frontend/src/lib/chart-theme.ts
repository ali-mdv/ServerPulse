export function readCssVar(name: string, fallback = "#000"): string {
  if (typeof window === "undefined") return fallback;
  const raw = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  if (!raw) return fallback;
  // Values like "0 0% 100%" need to be wrapped in hsl()
  if (/^[\d.]+\s+[\d.]+%\s+[\d.]+%$/.test(raw)) {
    return `hsl(${raw})`;
  }
  if (/^hsl\(/.test(raw) || /^#/.test(raw) || /^rgb/.test(raw)) {
    return raw;
  }
  return `hsl(${raw})`;
}

export interface ChartTheme {
  text: string;
  textMuted: string;
  border: string;
  background: string;
  popover: string;
}

export function readChartTheme(): ChartTheme {
  return {
    text: readCssVar("--foreground"),
    textMuted: readCssVar("--muted-foreground"),
    border: readCssVar("--border"),
    background: readCssVar("--background"),
    popover: readCssVar("--popover"),
  };
}
