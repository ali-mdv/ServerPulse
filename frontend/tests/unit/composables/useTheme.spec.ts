import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h, nextTick } from "vue";
import type { Theme } from "@/composables/useTheme";

function localStorageMock() {
  const store = new Map<string, string>();
  globalThis.localStorage = {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    removeItem: (k: string) => void store.delete(k),
    clear: () => store.clear(),
    key: (i: number) => Array.from(store.keys())[i] ?? null,
    get length() { return store.size; },
  } as Storage;
}

function withMatchMedia(dark: boolean) {
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: (query: string) => ({
      matches: dark && query.includes("dark"),
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }),
  });
}

function harness() {
  type UseTheme = () => {
    theme: { value: Theme };
    setTheme: (v: Theme) => void;
    toggle: () => void;
    isDark: { value: boolean };
  };
  let captured: ReturnType<UseTheme> | null = null;
  const Comp = defineComponent({
    setup() {
      const ut = (globalThis as any).__useTheme as UseTheme;
      captured = ut();
      return () => h("div");
    },
  });
  const wrapper = mount(Comp);
  return { wrapper, get captured() { return captured!; } };
}

describe("useTheme", () => {
  beforeEach(async () => {
    // The composable caches "initialised" state at module scope. We
    // reset the module each time so every harness sees a fresh theme.
    vi.resetModules();
    localStorageMock();
    withMatchMedia(false);
    document.documentElement.classList.remove("dark");
  });

  afterEach(() => {
    document.documentElement.classList.remove("dark");
    localStorage.clear();
  });

  it("starts as system with no preference stored", async () => {
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured } = makeHarness();
    expect(captured.theme.value).toBe("system");
  });

  it("setTheme switches class and stores choice", async () => {
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured, wrapper } = makeHarness();
    captured.setTheme("dark");
    await nextTick();
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    expect(localStorage.getItem("theme")).toBe("dark");
    wrapper.unmount();
  });

  it("setTheme('light') removes dark class", async () => {
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    document.documentElement.classList.add("dark");
    const { captured, wrapper } = makeHarness();
    captured.setTheme("light");
    await nextTick();
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    wrapper.unmount();
  });

  it("reads the stored theme on first invocation", async () => {
    localStorage.setItem("theme", "dark");
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured, wrapper } = makeHarness();
    await nextTick();
    expect(captured.theme.value).toBe("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    wrapper.unmount();
  });

  it("reads from the settings JSON if present", async () => {
    localStorage.setItem("settings", JSON.stringify({ theme: "dark" }));
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured, wrapper } = makeHarness();
    await nextTick();
    expect(captured.theme.value).toBe("dark");
    wrapper.unmount();
  });

  it("toggle from system → dark when system pref is light", async () => {
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured, wrapper } = makeHarness();
    captured.toggle();
    await nextTick();
    expect(captured.theme.value).toBe("dark");
    wrapper.unmount();
  });

  it("toggle from dark → light", async () => {
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured, wrapper } = makeHarness();
    captured.setTheme("dark");
    await nextTick();
    captured.toggle();
    await nextTick();
    expect(captured.theme.value).toBe("light");
    wrapper.unmount();
  });

  it("isDark reflects theme + system pref", async () => {
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured, wrapper } = makeHarness();
    captured.setTheme("dark");
    await nextTick();
    expect(captured.isDark.value).toBe(true);
    captured.setTheme("light");
    await nextTick();
    expect(captured.isDark.value).toBe(false);
    wrapper.unmount();
  });

  it("rejects malformed settings JSON", async () => {
    localStorage.setItem("settings", "not json");
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    expect(() => makeHarness()).not.toThrow();
  });

  it("ignores unknown stored theme values", async () => {
    localStorage.setItem("theme", "fuchsia");
    const { useTheme } = await import("@/composables/useTheme");
    (globalThis as any).__useTheme = useTheme;
    const { captured } = makeHarness();
    expect(captured.theme.value).toBe<Theme>("system");
  });
});

// makeHarness mounts a tiny component so Vue's onMounted + watch fire
// exactly as they would in a real app. The composable is read off
// globalThis because each test re-imports it after vi.resetModules().
function makeHarness() {
  type UseTheme = () => {
    theme: { value: Theme };
    setTheme: (v: Theme) => void;
    toggle: () => void;
    isDark: { value: boolean };
  };
  let captured: ReturnType<UseTheme> | null = null;
  const Comp = defineComponent({
    setup() {
      const ut = (globalThis as any).__useTheme as UseTheme;
      captured = ut();
      return () => h("div");
    },
  });
  const wrapper = mount(Comp);
  return { wrapper, get captured() { return captured!; } };
}