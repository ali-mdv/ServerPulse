export function readCssVar(name: string, fallback = "#000"): string {
  if (typeof window === "undefined") return fallback;
  const raw = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  if (!raw) return fallback;
  if (/^[\d.]+\s+[\d.]+%\s+[\d.]+%$/.test(raw)) {
    return `hsl(${raw})`;
  }
  if (/^hsl\(/.test(raw) || /^#/.test(raw) || /^rgb/.test(raw)) {
    return raw;
  }
  return `hsl(${raw})`;
}

/**
 * Take a value like "hsl(142 71% 45%)" and produce an alpha version
 * using the comma-separated hsla() syntax: "hsla(142, 71%, 45%, 0.33)".
 * This format is reliably parsed by ECharts / zrender.
 * If the input is already a CSS color, returns it unchanged.
 */
export function withAlpha(color: string, alpha: number): string {
  const m = color.match(/^hsl\(\s*([\d.]+)\s+([\d.]+)%\s+([\d.]+)%\s*\)$/);
  if (m) {
    return `hsla(${m[1]}, ${m[2]}%, ${m[3]}%, ${alpha})`;
  }
  const m2 = color.match(/^rgb\(\s*([\d.,\s]+)\s*\)$/);
  if (m2) {
    return `rgba(${m2[1]}, ${alpha})`;
  }
  const m3 = color.match(/^#([0-9a-fA-F]{6})$/);
  if (m3) {
    const a = Math.round(alpha * 255)
      .toString(16)
      .padStart(2, "0");
    return `#${m3[1]}${a}`;
  }
  return color;
}

export interface ChartTheme {
  text: string;
  textMuted: string;
  border: string;
  background: string;
  popover: string;
  info: string;
  success: string;
  warning: string;
  critical: string;
  muted: string;
  free: string;
}

export function readChartTheme(): ChartTheme {
  const isDark =
    typeof document !== "undefined" &&
    document.documentElement.classList.contains("dark");
  return {
    text: readCssVar("--foreground"),
    textMuted: readCssVar("--muted-foreground"),
    border: readCssVar("--border"),
    background: readCssVar("--background"),
    popover: readCssVar("--popover"),
    info: readCssVar("--info"),
    success: readCssVar("--success"),
    warning: readCssVar("--warning"),
    critical: readCssVar("--critical"),
    muted: readCssVar("--muted"),
    // "Free" / secondary slice: a neutral that's visible in both modes
    free: isDark ? "hsl(215 16% 38%)" : "hsl(215 16% 78%)",
  };
}
