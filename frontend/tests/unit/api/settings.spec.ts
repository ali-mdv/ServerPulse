import { describe, it, expect, beforeEach, vi } from "vitest";

const apiMock = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn(),
}));

const authMock = vi.hoisted(() => ({
  getToken: vi.fn(() => "Bearer stub"),
}));

vi.mock("@/plugins/axios", () => ({ useApi: () => apiMock }));
vi.mock("@/stores/auth", () => ({ useAuthStore: () => authMock }));

import { fetchSettings, updateSettings } from "@/api/settings";
import { axiosError } from "@/../tests/setup/axios";

const settings = {
  appearance: "dark",
  historyPollIntervalSeconds: 30,
  historyRetentionSeconds: 604800,
  updatedAt: "2026-01-01T00:00:00Z",
};

describe("api/settings", () => {
  beforeEach(() => {
    apiMock.get.mockReset();
    apiMock.put.mockReset();
    authMock.getToken.mockClear();
    authMock.getToken.mockReturnValue("Bearer stub");
  });

  it("fetchSettings returns the settings singleton", async () => {
    apiMock.get.mockResolvedValue({ data: { settings } });
    await expect(fetchSettings()).resolves.toEqual(settings);
    expect(apiMock.get).toHaveBeenCalledWith("settings", expect.anything());
  });

  it("updateSettings PUTs the payload and returns the result", async () => {
    apiMock.put.mockResolvedValue({
      data: { settings: { ...settings, appearance: "light" } },
    });
    const result = await updateSettings({ appearance: "light" });
    expect(result.appearance).toBe("light");
    expect(apiMock.put).toHaveBeenCalledWith(
      "settings",
      { appearance: "light" },
      expect.anything(),
    );
  });

  it("surfaces the backend error message", async () => {
    apiMock.get.mockRejectedValue(axiosError(500, { error: "boom" }));
    await expect(fetchSettings()).rejects.toThrow("boom");
  });
});