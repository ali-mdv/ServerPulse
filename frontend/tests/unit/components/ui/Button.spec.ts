import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Button from "@/components/ui/Button.vue";

describe("Button", () => {
  it("renders the default variant class", () => {
    const w = mount(Button, { slots: { default: "Click" } });
    expect(w.classes()).toContain("btn-primary");
    expect(w.text()).toBe("Click");
  });

  it("applies the selected variant class", () => {
    const w = mount(Button, { props: { variant: "danger" }, slots: { default: "x" } });
    expect(w.classes()).toContain("btn-danger");
  });

  it("applies the size class", () => {
    const w = mount(Button, { props: { size: "sm" }, slots: { default: "x" } });
    expect(w.classes()).toContain("btn-sm");
  });

  it("disables when disabled", () => {
    const w = mount(Button, { props: { disabled: true }, slots: { default: "x" } });
    expect(w.attributes("disabled")).toBeDefined();
  });

  it("disables when loading and shows spinner", () => {
    const w = mount(Button, { props: { loading: true }, slots: { default: "x" } });
    expect(w.attributes("disabled")).toBeDefined();
    expect(w.attributes("aria-busy")).toBe("true");
    expect(w.find("span").exists()).toBe(true);
  });

  it("renders icon-left / icon-right slots", () => {
    const w = mount(Button, {
      slots: {
        default: "Save",
        "icon-left": "<span data-testid='icon-left'>+</span>",
        "icon-right": "<span data-testid='icon-right'>*</span>",
      },
    });
    expect(w.find("[data-testid='icon-left']").exists()).toBe(true);
    expect(w.find("[data-testid='icon-right']").exists()).toBe(true);
  });

  it("respects the type prop", () => {
    const w = mount(Button, { props: { type: "submit" }, slots: { default: "x" } });
    expect(w.attributes("type")).toBe("submit");
  });

  it("merges an extra class via $attrs", () => {
    const w = mount(Button, { attrs: { class: "extra-class" }, slots: { default: "x" } });
    expect(w.classes()).toContain("extra-class");
    expect(w.classes()).toContain("btn-primary");
  });
});