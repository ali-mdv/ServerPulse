import { describe, it, expect } from "vitest";
import { withAlpha, readChartTheme, readCssVar } from "@/lib/chart-theme";

describe("withAlpha", () => {
  it("converts space-separated hsl() to comma-separated hsla()", () => {
    expect(withAlpha("hsl(142 71% 45%)", 0.33)).toBe("hsla(142, 71%, 45%, 0.33)");
  });

  it("preserves extra precision", () => {
    expect(withAlpha("hsl(200.5 12.3% 99%)", 1)).toBe(
      "hsla(200.5, 12.3%, 99%, 1)",
    );
  });

  it("passes through rgb() with appended alpha", () => {
    expect(withAlpha("rgb(10, 20, 30)", 0.5)).toBe("rgba(10, 20, 30, 0.5)");
  });

  it("appends alpha as hex byte for #rrggbb", () => {
    expect(withAlpha("#1a2b3c", 0.5)).toBe("#1a2b3c80");
  });

  it("rounds the alpha byte to two hex digits", () => {
    // alpha 1.0 → 255 → ff
    expect(withAlpha("#000000", 1)).toBe("#000000ff");
    // alpha 0 → 0 → 00
    expect(withAlpha("#ffffff", 0)).toBe("#ffffff00");
  });

  it("returns unknown formats unchanged", () => {
    expect(withAlpha("red", 0.5)).toBe("red");
    expect(withAlpha("hsl(0 0% 0% / 0.5)", 0.5)).toBe("hsl(0 0% 0% / 0.5)");
  });
});

describe("readCssVar", () => {
  it("returns fallback when window is undefined", () => {
    const originalWindow = (globalThis as any).window;
    delete (globalThis as any).window;
    try {
      expect(readCssVar("--anything")).toBe("#000");
    } finally {
      (globalThis as any).window = originalWindow;
    }
  });

  it("returns fallback when the var is unset", () => {
    // happy-dom returns "" for unknown css vars on documentElement.
    document.documentElement.style.removeProperty("--unset-var");
    expect(readCssVar("--unset-var", "fallback")).toBe("fallback");
  });

  it("prefixes hsl() space-separated values", () => {
    document.documentElement.style.setProperty("--hsl-space", "200 50% 50%");
    expect(readCssVar("--hsl-space")).toBe("hsl(200 50% 50%)");
  });
});

describe("readChartTheme", () => {
  it("uses light free color when document is not in dark mode", () => {
    document.documentElement.classList.remove("dark");
    const theme = readChartTheme();
    expect(theme.free).toContain("215 16%");
  });

  it("uses dark free color when document is in dark mode", () => {
    document.documentElement.classList.add("dark");
    const theme = readChartTheme();
    expect(theme.free).toContain("215 16%");
    document.documentElement.classList.remove("dark");
  });
});