import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";

const apiMock = vi.hoisted(() => ({
  fetchSettings: vi.fn(),
  updateSettings: vi.fn(),
}));

vi.mock("@/api/settings", () => apiMock);

import { useSettingsStore } from "@/stores/settings";

const settings = {
  appearance: "system" as const,
  historyPollIntervalSeconds: 30,
  historyRetentionSeconds: 604800,
  updatedAt: "2026-01-01T00:00:00Z",
};

describe("settings store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    apiMock.fetchSettings.mockReset();
    apiMock.updateSettings.mockReset();
  });

  it("load stores the settings", async () => {
    apiMock.fetchSettings.mockResolvedValue(settings);
    const store = useSettingsStore();
    await store.load();
    expect(store.settings).toEqual(settings);
    expect(store.error).toBeNull();
  });

  it("load records an error message on failure", async () => {
    apiMock.fetchSettings.mockRejectedValue(new Error("offline"));
    const store = useSettingsStore();
    await store.load();
    expect(store.settings).toBeNull();
    expect(store.error).toBe("offline");
  });

  it("save stores the updated settings and rethrows on failure", async () => {
    apiMock.updateSettings.mockResolvedValue({ ...settings, appearance: "dark" });
    const store = useSettingsStore();
    await store.save({ appearance: "dark" });
    expect(store.settings?.appearance).toBe("dark");

    apiMock.updateSettings.mockRejectedValue(new Error("nope"));
    await expect(store.save({ appearance: "light" })).rejects.toThrow("nope");
    expect(store.error).toBe("nope");
  });
});