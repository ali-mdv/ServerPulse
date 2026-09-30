import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Spinner from "@/components/ui/Spinner.vue";

describe("Spinner", () => {
  it.each([
    ["xs", "w-3"],
    ["sm", "w-3.5"],
    ["md", "w-4"],
    ["lg", "w-6"],
  ])("size=%s maps to width class", (size, cls) => {
    const w = mount(Spinner, { props: { size } });
    expect(w.classes().join(" ")).toContain(cls);
  });

  it("defaults to md", () => {
    const w = mount(Spinner);
    expect(w.classes()).toContain("w-4");
  });

  it("is role=status", () => {
    const w = mount(Spinner);
    expect(w.attributes("role")).toBe("status");
  });
});