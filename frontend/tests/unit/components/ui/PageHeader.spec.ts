import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import PageHeader from "@/components/ui/PageHeader.vue";

describe("PageHeader", () => {
  it("renders the title", () => {
    const w = mount(PageHeader, { props: { title: "Settings" } });
    expect(w.find("h1").text()).toBe("Settings");
  });

  it("renders subtitle when provided", () => {
    const w = mount(PageHeader, { props: { title: "T", subtitle: "S" } });
    expect(w.find("p").text()).toBe("S");
  });

  it("omits subtitle when missing", () => {
    const w = mount(PageHeader, { props: { title: "T" } });
    expect(w.find("p").exists()).toBe(false);
  });

  it("renders default slot", () => {
    const w = mount(PageHeader, {
      props: { title: "T" },
      slots: { default: "<button data-testid='action'>Add</button>" },
    });
    expect(w.find("[data-testid='action']").exists()).toBe(true);
  });
});