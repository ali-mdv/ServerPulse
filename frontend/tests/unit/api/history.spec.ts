import { describe, it, expect } from "vitest";
import { rangeSeconds, rangeBucketSeconds, type HistoryRange } from "@/api/history";

describe("api/history range tables", () => {
  const ranges: HistoryRange[] = ["1h", "6h", "24h", "7d"];

  it.each(ranges)("rangeSeconds(%s) is positive", (r) => {
    expect(rangeSeconds(r)).toBeGreaterThan(0);
  });

  it.each(ranges)("rangeBucketSeconds(%s) is positive", (r) => {
    expect(rangeBucketSeconds(r)).toBeGreaterThan(0);
  });

  it("rangeSeconds(7d) is 7 * 24 * 3600", () => {
    expect(rangeSeconds("7d")).toBe(7 * 24 * 60 * 60);
  });

  it("rangeSeconds(24h) is 86400", () => {
    expect(rangeSeconds("24h")).toBe(86400);
  });

  it("rangeSeconds(6h) is 21600", () => {
    expect(rangeSeconds("6h")).toBe(6 * 60 * 60);
  });

  it("rangeSeconds(1h) is 3600", () => {
    expect(rangeSeconds("1h")).toBe(60 * 60);
  });

  it("rangeBucketSeconds scales with the range", () => {
    expect(rangeBucketSeconds("1h")).toBeLessThan(rangeBucketSeconds("7d"));
  });
});