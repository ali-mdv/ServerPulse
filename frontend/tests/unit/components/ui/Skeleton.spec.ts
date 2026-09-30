import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Skeleton from "@/components/ui/Skeleton.vue";

describe("Skeleton", () => {
  it("renders a div", () => {
    const w = mount(Skeleton);
    expect(w.element.tagName).toBe("DIV");
  });

  it("applies the width and height as inline styles", () => {
    const w = mount(Skeleton, { props: { width: "200px", height: "12px" } });
    const style = (w.element as HTMLElement).style;
    expect(style.width).toBe("200px");
    expect(style.height).toBe("12px");
  });

  it("has default size 100% / 1rem", () => {
    const w = mount(Skeleton);
    const style = (w.element as HTMLElement).style;
    expect(style.width).toBe("100%");
    expect(style.height).toBe("1rem");
  });

  it("is aria-hidden", () => {
    const w = mount(Skeleton);
    expect(w.attributes("aria-hidden")).toBe("true");
  });
});