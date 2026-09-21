import { describe, it, expect, beforeEach, vi } from "vitest";

const apiMock = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}));

const authMock = vi.hoisted(() => ({
  getToken: vi.fn(() => "Bearer stub"),
}));

vi.mock("@/plugins/axios", () => ({ useApi: () => apiMock }));
vi.mock("@/stores/auth", () => ({ useAuthStore: () => authMock }));

import {
  fetchServers,
  fetchServer,
  createServer,
  updateServer,
  deleteServer,
  generateServerApiKey,
} from "@/api/servers";
import { axiosError } from "@/../tests/setup/axios";

describe("api/servers", () => {
  beforeEach(() => {
    apiMock.get.mockReset();
    apiMock.post.mockReset();
    apiMock.put.mockReset();
    apiMock.delete.mockReset();
    authMock.getToken.mockClear();
    authMock.getToken.mockReturnValue("Bearer stub");
  });

  describe("fetchServers", () => {
    it("returns the servers list", async () => {
      apiMock.get.mockResolvedValue({ data: { servers: [{ id: "1" }, { id: "2" }] } });
      const list = await fetchServers();
      expect(list).toHaveLength(2);
    });

    it("throws friendly message on axios error", async () => {
      apiMock.get.mockRejectedValue(axiosError(500, { error: "boom" }));
      await expect(fetchServers()).rejects.toThrow("boom");
    });

    it("throws Network error on non-axios error", async () => {
      apiMock.get.mockRejectedValue(new Error("offline"));
      await expect(fetchServers()).rejects.toThrow("Network error");
    });
  });

  describe("fetchServer", () => {
    it("returns server", async () => {
      apiMock.get.mockResolvedValue({ data: { server: { id: "1" } } });
      const s = await fetchServer("1");
      expect(s?.id).toBe("1");
    });

    it("returns null on 404", async () => {
      apiMock.get.mockRejectedValue(axiosError(404));
      const s = await fetchServer("missing");
      expect(s).toBeNull();
    });

    it("throws on other errors", async () => {
      apiMock.get.mockRejectedValue(axiosError(500, { error: "down" }));
      await expect(fetchServer("1")).rejects.toThrow("down");
    });
  });

  describe("createServer", () => {
    it("posts and returns server", async () => {
      apiMock.post.mockResolvedValue({ data: { server: { id: "1" } } });
      const s = await createServer({ name: "n" });
      expect(s.id).toBe("1");
      expect(apiMock.post).toHaveBeenCalledWith(
        "servers",
        { name: "n" },
        expect.objectContaining({ headers: expect.objectContaining({ Authorization: "Bearer stub" }) }),
      );
    });

    it("throws friendly on error", async () => {
      apiMock.post.mockRejectedValue(axiosError(422, { error: "bad" }));
      await expect(createServer({ name: "n" })).rejects.toThrow("bad");
    });
  });

  describe("updateServer", () => {
    it("puts and returns server", async () => {
      apiMock.put.mockResolvedValue({ data: { server: { id: "1", name: "renamed" } } });
      const s = await updateServer("1", { name: "renamed" });
      expect(s.name).toBe("renamed");
    });
  });

  describe("deleteServer", () => {
    it("deletes", async () => {
      apiMock.delete.mockResolvedValue({});
      await expect(deleteServer("1")).resolves.toBeUndefined();
    });

    it("throws on error", async () => {
      apiMock.delete.mockRejectedValue(axiosError(500, { error: "x" }));
      await expect(deleteServer("1")).rejects.toThrow("x");
    });
  });

  describe("generateServerApiKey", () => {
    it("returns apiKey", async () => {
      apiMock.post.mockResolvedValue({ data: { server: { id: "1" }, apiKey: "k" } });
      const k = await generateServerApiKey("1");
      expect(k).toBe("k");
    });
  });
});