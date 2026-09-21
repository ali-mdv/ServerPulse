import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Badge from "@/components/ui/Badge.vue";

describe("Badge", () => {
  it.each([
    ["neutral", "badge-neutral"],
    ["critical", "badge-critical"],
    ["warning", "badge-warning"],
    ["info", "badge-info"],
    ["success", "badge-success"],
  ])("maps tone=%s to %s", (tone, cls) => {
    const w = mount(Badge, { props: { tone }, slots: { default: "x" } });
    expect(w.classes()).toContain(cls);
  });

  it("renders slot content", () => {
    const w = mount(Badge, { slots: { default: "online" } });
    expect(w.text()).toBe("online");
  });

  it("merges an extra class via $attrs", () => {
    const w = mount(Badge, { attrs: { class: "extra" } });
    expect(w.classes()).toContain("extra");
  });
});