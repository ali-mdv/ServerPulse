import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import { useMobile } from "@/hooks/use-mobile";

function withWidth(width: number) {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: width });
}

function makeHarness() {
  let captured: ReturnType<typeof useMobile> | null = null;
  const Comp = defineComponent({
    setup() {
      captured = useMobile();
      return () => h("div");
    },
  });
  const wrapper = mount(Comp);
  return { wrapper, get captured() { return captured!; } };
}

describe("useMobile", () => {
  beforeEach(() => {
    withWidth(1200);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("flags isMobile=false on desktop", () => {
    withWidth(1200);
    const { captured } = makeHarness();
    expect(captured.isMobile.value).toBe(false);
  });

  it("flags isMobile=true below 768", () => {
    withWidth(500);
    const { captured } = makeHarness();
    expect(captured.isMobile.value).toBe(true);
  });

  it("updates on resize", async () => {
    withWidth(1200);
    const { captured, wrapper } = makeHarness();
    expect(captured.isMobile.value).toBe(false);

    withWidth(500);
    window.dispatchEvent(new Event("resize"));
    await wrapper.vm.$nextTick();
    expect(captured.isMobile.value).toBe(true);

    withWidth(1200);
    window.dispatchEvent(new Event("resize"));
    await wrapper.vm.$nextTick();
    expect(captured.isMobile.value).toBe(false);

    wrapper.unmount();
  });
});