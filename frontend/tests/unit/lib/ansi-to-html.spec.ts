import { describe, it, expect } from "vitest";
import { ansiToHtml } from "@/lib/ansi-to-html";

describe("ansiToHtml", () => {
  it("returns plain text unchanged", () => {
    expect(ansiToHtml("hello world")).toBe("hello world");
  });

  it("strips null bytes and control chars but keeps the printable text", () => {
    expect(ansiToHtml("foo\x00\x01\x02bar")).toBe("foobar");
  });

  it("preserves newline and tab characters", () => {
    expect(ansiToHtml("a\nb\tc")).toBe("a\nb\tc");
  });

  it("converts ANSI color escapes to HTML spans", () => {
    // ESC[31m = red
    const out = ansiToHtml("\x1B[31mred\x1B[0m");
    expect(out).toContain("<span");
    expect(out).toContain("red");
  });

  it("removes the bell character", () => {
    expect(ansiToHtml("ring\x07")).toBe("ring");
  });
});