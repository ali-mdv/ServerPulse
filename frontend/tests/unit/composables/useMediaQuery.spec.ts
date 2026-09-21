import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import { useMediaQuery } from "@/composables/useMediaQuery";

function withMatchMedia(matches: boolean) {
  const listeners: Array<(e: { matches: boolean }) => void> = [];
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: (query: string) => ({
      matches,
      media: query,
      onchange: null,
      addEventListener: (_: string, l: (e: any) => void) => listeners.push(l),
      removeEventListener: vi.fn(),
      addListener: (l: (e: any) => void) => listeners.push(l),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }),
  });
  return { listeners, setMatches(next: boolean) { matches = next; } };
}

function harness() {
  let captured: ReturnType<typeof useMediaQuery> | null = null;
  const Comp = defineComponent({
    setup() {
      captured = useMediaQuery("(min-width: 768px)");
      return () => h("div");
    },
  });
  const wrapper = mount(Comp);
  return { wrapper, get captured() { return captured!; } };
}

describe("useMediaQuery", () => {
  beforeEach(() => {
    withMatchMedia(false);
  });

  afterEach(() => {
    // restore default for next tests
    Object.defineProperty(window, "matchMedia", { configurable: true, value: undefined });
  });

  it("starts as false on mount when matchMedia returns false", () => {
    withMatchMedia(false);
    const { captured } = harness();
    expect(captured.value).toBe(false);
  });

  it("updates when matchMedia changes", async () => {
    const { listeners, setMatches } = withMatchMedia(false);
    const { captured, wrapper } = harness();
    expect(captured.value).toBe(false);

    setMatches(true);
    for (const l of listeners) l({ matches: true });
    await wrapper.vm.$nextTick();
    expect(captured.value).toBe(true);
  });

  it("returns false when window has no matchMedia", () => {
    Object.defineProperty(window, "matchMedia", { configurable: true, value: undefined });
    const { captured } = harness();
    expect(captured.value).toBe(false);
  });

  it("is readonly", () => {
    withMatchMedia(true);
    const { captured } = harness();
    // @ts-expect-error — readonly ref shouldn't accept assignment
    captured.value = false;
    expect(captured.value).toBe(true);
  });
});