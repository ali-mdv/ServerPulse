import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import axios from "axios";

const apiMock = vi.hoisted(() => ({
  get: vi.fn(),
}));

vi.mock("@/plugins/axios", () => ({
  useApi: () => apiMock,
}));

vi.mock("@/api/state", () => ({
  fetchServicesState: vi.fn(),
  fetchSystemState: vi.fn(),
}));

vi.mock("@/api/history", () => ({
  fetchServiceHistory: vi.fn(),
}));

import { useSystemStore } from "@/stores/system";
import { fetchServicesState, fetchSystemState } from "@/api/state";
import { axiosError } from "@/../tests/setup/axios";
import { localStorageMock } from "@/../tests/setup";

void axios;

describe("system store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorageMock();
    apiMock.get.mockReset();
    vi.mocked(fetchServicesState).mockReset();
    vi.mocked(fetchSystemState).mockReset();
  });

  it("defaults intervalMS to 10s when no settings stored", () => {
    const store = useSystemStore();
    expect(store.intervalMS).toBe(10_000);
  });

  it("reads refreshInterval from settings JSON", () => {
    localStorage.setItem("settings", JSON.stringify({ refreshInterval: 30 }));
    setActivePinia(createPinia());
    const store = useSystemStore();
    expect(store.intervalMS).toBe(30_000);
  });

  it("ignores non-numeric / non-positive refreshInterval", () => {
    localStorage.setItem("settings", JSON.stringify({ refreshInterval: -1 }));
    setActivePinia(createPinia());
    const store = useSystemStore();
    expect(store.intervalMS).toBe(10_000);
  });

  it("ignores malformed settings JSON", () => {
    localStorage.setItem("settings", "not-json");
    setActivePinia(createPinia());
    const store = useSystemStore();
    expect(store.intervalMS).toBe(10_000);
  });

  it("fetchSystemUsage maps 401 to Access denied", async () => {
    apiMock.get.mockRejectedValue(axiosError(401));
    const store = useSystemStore();
    await expect(store.fetchSystemUsage()).rejects.toThrow("Access denied");
  });

  it("fetchSystemUsage maps other errors to generic", async () => {
    apiMock.get.mockRejectedValue(axiosError(500));
    const store = useSystemStore();
    await expect(store.fetchSystemUsage()).rejects.toThrow("Failed to fetch system usage info");
  });

  it("fetchPM2Services returns processes + available", async () => {
    apiMock.get.mockResolvedValue({ data: { processes: [{ name: "a" }], available: true } });
    const store = useSystemStore();
    const out = await store.fetchPM2Services();
    expect(out.services).toHaveLength(1);
    expect(out.available).toBe(true);
  });

  it("fetchDockerContainers maps 401 to Access denied", async () => {
    apiMock.get.mockRejectedValue(axiosError(401));
    const store = useSystemStore();
    await expect(store.fetchDockerContainers()).rejects.toThrow("Access denied");
  });

  it("startDockerContainer maps 404 to friendly message", async () => {
    apiMock.get.mockRejectedValue(axiosError(404));
    const store = useSystemStore();
    await expect(
      store.startDockerContainer({ id: "abc", name: "web" } as any),
    ).rejects.toThrow("Container 'web' was not found.");
  });

  it("startDockerContainer maps 401 to Access denied", async () => {
    apiMock.get.mockRejectedValue(axiosError(401));
    const store = useSystemStore();
    await expect(
      store.startDockerContainer({ id: "abc", name: "web" } as any),
    ).rejects.toThrow("Access denied");
  });

  it("startDockerContainer maps other errors to friendly", async () => {
    apiMock.get.mockRejectedValue(axiosError(500));
    const store = useSystemStore();
    await expect(
      store.startDockerContainer({ id: "abc", name: "web" } as any),
    ).rejects.toThrow("Unable to start container 'web'.");
  });

  it("startPM2Service maps 404 to friendly message", async () => {
    apiMock.get.mockRejectedValue(axiosError(404));
    const store = useSystemStore();
    await expect(
      store.startPM2Service({ pm_id: "0", name: "api" } as any),
    ).rejects.toThrow("Service 'api' was not found.");
  });

  it("getPM2ServiceLogs returns the logs", async () => {
    apiMock.get.mockResolvedValue({ data: { logs: "hello" } });
    const store = useSystemStore();
    const logs = await store.getPM2ServiceLogs({ pm_id: "0", name: "api" } as any);
    expect(logs).toBe("hello");
  });

  it("fetchServicesState delegates to api helper", async () => {
    vi.mocked(fetchServicesState).mockResolvedValue({
      pm2: { provider: "pm2", available: true, services: [], updatedAt: "" },
      docker: { provider: "docker", available: true, services: [], updatedAt: "" },
    });
    const store = useSystemStore();
    const out = await store.fetchServicesState();
    expect(out.pm2.provider).toBe("pm2");
  });
});