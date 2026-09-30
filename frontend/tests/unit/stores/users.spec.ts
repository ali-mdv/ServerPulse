import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import axios from "axios";

const apiMock = vi.hoisted(() => ({
  post: vi.fn(),
  get: vi.fn(),
  delete: vi.fn(),
}));

vi.mock("@/plugins/axios", () => ({
  useApi: () => apiMock,
}));

import { useUsersStore } from "@/stores/users";
import { useAuthStore } from "@/stores/auth";
import { axiosError } from "@/../tests/setup/axios";
import { localStorageMock } from "@/../tests/setup";

void axios;

describe("users store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorageMock();
    apiMock.post.mockReset();
    apiMock.get.mockReset();
    apiMock.delete.mockReset();
  });

  describe("getProfile", () => {
    it("returns user on success", async () => {
      apiMock.get.mockResolvedValue({
        data: { user: { id: "u1", email: "me@x.io" } },
      });

      const users = useUsersStore();
      const profile = await users.getProfile();
      expect(profile).toEqual({ id: "u1", email: "me@x.io" });
      expect(apiMock.get).toHaveBeenCalledWith(
        "users/profile",
        expect.objectContaining({ headers: expect.objectContaining({ Authorization: expect.any(String) }) }),
      );
    });

    it("logs out and throws Access denied on error", async () => {
      apiMock.get.mockRejectedValue(new Error("network"));
      const auth = useAuthStore();
      // Pre-seed an authenticated user so logout is meaningful.
      auth.user = { email: "me@x.io", token: "abc" } as any;

      const users = useUsersStore();
      await expect(users.getProfile()).rejects.toThrow("Access denied");
      expect(auth.user).toBeNull();
    });
  });

  describe("updateProfile", () => {
    it("returns updated user", async () => {
      apiMock.post.mockResolvedValue({
        data: { user: { id: "u1", email: "new@x.io" } },
      });
      const users = useUsersStore();
      const u = await users.updateProfile("new@x.io");
      expect(u.email).toBe("new@x.io");
    });

    it("logs out and throws on error", async () => {
      apiMock.post.mockRejectedValue(new Error("boom"));
      const auth = useAuthStore();
      auth.user = { email: "me@x.io", token: "abc" } as any;
      const users = useUsersStore();
      await expect(users.updateProfile("x@y.z")).rejects.toThrow("Access denied");
      expect(auth.user).toBeNull();
    });
  });

  describe("getUsers", () => {
    it("returns the list", async () => {
      apiMock.get.mockResolvedValue({
        data: { users: [{ id: "1" }, { id: "2" }] },
      });
      const users = useUsersStore();
      const list = await users.getUsers();
      expect(list).toHaveLength(2);
    });

    it("throws Access denied on 401", async () => {
      apiMock.get.mockRejectedValue(axiosError(401));
      const users = useUsersStore();
      await expect(users.getUsers()).rejects.toThrow("Access denied");
    });

    it("throws generic message on other errors", async () => {
      apiMock.get.mockRejectedValue(axiosError(500));
      const users = useUsersStore();
      await expect(users.getUsers()).rejects.toThrow("Failed to get users list");
    });
  });

  describe("addUser", () => {
    it("creates a user", async () => {
      apiMock.post.mockResolvedValue({
        data: { user: { id: "9", email: "new@x.io" } },
      });
      const users = useUsersStore();
      const u = await users.addUser("new@x.io", "secret");
      expect(u.email).toBe("new@x.io");
    });

    it("throws Access denied on 401", async () => {
      apiMock.post.mockRejectedValue(axiosError(401));
      const users = useUsersStore();
      await expect(users.addUser("a@b.c", "secret")).rejects.toThrow("Access denied");
    });

    it("throws generic message on other errors", async () => {
      apiMock.post.mockRejectedValue(axiosError(500));
      const users = useUsersStore();
      await expect(users.addUser("a@b.c", "secret")).rejects.toThrow("Failed to create new user");
    });
  });

  describe("deleteUser", () => {
    it("deletes the user", async () => {
      apiMock.delete.mockResolvedValue({ data: { message: "user deleted" } });
      const users = useUsersStore();
      await users.deleteUser("1");
      expect(apiMock.delete).toHaveBeenCalledWith(
        "users/1",
        expect.objectContaining({ headers: expect.objectContaining({ Authorization: expect.any(String) }) }),
      );
    });

    it("throws Access denied on 401", async () => {
      apiMock.delete.mockRejectedValue(axiosError(401));
      const users = useUsersStore();
      await expect(users.deleteUser("1")).rejects.toThrow("Access denied");
    });

    it("throws generic message on other errors", async () => {
      apiMock.delete.mockRejectedValue(axiosError(500));
      const users = useUsersStore();
      await expect(users.deleteUser("1")).rejects.toThrow("Failed to delete user");
    });
  });
});