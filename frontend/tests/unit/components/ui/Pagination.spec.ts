import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Pagination from "@/components/ui/Pagination.vue";

describe("Pagination", () => {
  it("renders nothing when there is a single page", () => {
    const w = mount(Pagination, { props: { page: 1, total: 5, pageSize: 10 } });
    expect(w.find("nav").exists()).toBe(false);
  });

  it("renders the page count when there are multiple pages", () => {
    const w = mount(Pagination, { props: { page: 2, total: 25, pageSize: 10 } });
    expect(w.find("nav").exists()).toBe(true);
    expect(w.text()).toContain("Page 2 of 3");
  });

  it("applies the aria-label", () => {
    const w = mount(Pagination, {
      props: { page: 1, total: 25, pageSize: 10, ariaLabel: "Alerts pagination" },
    });
    expect(w.attributes("aria-label")).toBe("Alerts pagination");
  });

  it("disables Prev on the first page", () => {
    const w = mount(Pagination, { props: { page: 1, total: 25, pageSize: 10 } });
    const [prev] = w.findAll("button");
    expect(prev.attributes("disabled")).toBeDefined();
  });

  it("disables Next on the last page", () => {
    const w = mount(Pagination, { props: { page: 3, total: 25, pageSize: 10 } });
    const buttons = w.findAll("button");
    expect(buttons[buttons.length - 1].attributes("disabled")).toBeDefined();
  });

  it("emits update:page when Next is clicked", async () => {
    const w = mount(Pagination, { props: { page: 1, total: 25, pageSize: 10 } });
    const buttons = w.findAll("button");
    await buttons[buttons.length - 1].trigger("click");
    expect(w.emitted("update:page")?.[0]).toEqual([2]);
  });

  it("emits update:page when Prev is clicked", async () => {
    const w = mount(Pagination, { props: { page: 3, total: 25, pageSize: 10 } });
    const [prev] = w.findAll("button");
    await prev.trigger("click");
    expect(w.emitted("update:page")?.[0]).toEqual([2]);
  });
});
