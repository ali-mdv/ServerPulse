import { describe, it, expect, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useNotificationsPanel } from "@/composables/useNotificationsPanel";

describe("useNotificationsPanel", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("starts closed", () => {
    const { isOpen } = useNotificationsPanel();
    expect(isOpen.value).toBe(false);
  });

  it("opens and closes", () => {
    const { isOpen, open, close } = useNotificationsPanel();
    open();
    expect(isOpen.value).toBe(true);
    close();
    expect(isOpen.value).toBe(false);
  });

  it("toggle flips the flag", () => {
    const { isOpen, toggle } = useNotificationsPanel();
    toggle();
    expect(isOpen.value).toBe(true);
    toggle();
    expect(isOpen.value).toBe(false);
  });

  it("shares state across callers", () => {
    const a = useNotificationsPanel();
    const b = useNotificationsPanel();
    a.open();
    expect(b.isOpen.value).toBe(true);
  });
});