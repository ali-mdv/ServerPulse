import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import EmptyState from "@/components/ui/EmptyState.vue";

describe("EmptyState", () => {
  it("renders title and description", () => {
    const w = mount(EmptyState, {
      props: { title: "Nothing here", description: "Try later" },
    });
    expect(w.text()).toContain("Nothing here");
    expect(w.text()).toContain("Try later");
  });

  it("hides title/description when not provided", () => {
    const w = mount(EmptyState);
    expect(w.find("h3").exists()).toBe(false);
    expect(w.find("p").exists()).toBe(false);
  });

  it("renders the icon slot when present", () => {
    const w = mount(EmptyState, {
      props: { title: "x" },
      slots: { icon: "<svg data-testid='icon' />" },
    });
    expect(w.find("[data-testid='icon']").exists()).toBe(true);
  });

  it("renders the action slot when present", () => {
    const w = mount(EmptyState, {
      props: { title: "x" },
      slots: { action: "<button data-testid='action'>Go</button>" },
    });
    expect(w.find("[data-testid='action']").exists()).toBe(true);
  });
});