import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import axios from "axios";

const apiMock = vi.hoisted(() => ({
  post: vi.fn(),
  get: vi.fn(),
}));

vi.mock("@/plugins/axios", () => ({
  useApi: () => apiMock,
}));

import { useAuthStore } from "@/stores/auth";
import { axiosError } from "@/../tests/setup/axios";
import { localStorageMock } from "@/../tests/setup";

void axios; // silence unused-import if the file later drops direct usage

function makeLocalStorage() {
  return localStorageMock();
}

function makeJwt(expSecondsFromNow: number) {
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const payload = btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + expSecondsFromNow }));
  const sig = "fakesig";
  return `${header}.${payload}.${sig}`;
}

describe("auth store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    makeLocalStorage();
    apiMock.post.mockReset();
    apiMock.get.mockReset();
  });

  it("starts unauthenticated", () => {
    const auth = useAuthStore();
    expect(auth.isAuthenticated).toBe(false);
    expect(auth.user).toBeNull();
  });

  it("getToken returns empty string when no user", () => {
    const auth = useAuthStore();
    expect(auth.getToken()).toBe("");
  });

  it("getToken returns Bearer-prefixed token after login", async () => {
    const auth = useAuthStore();
    apiMock.post.mockResolvedValue({ data: { token: "abc" } });

    await auth.login("a@b.c", "secret");
    expect(auth.getToken()).toBe("Bearer abc");
    expect(auth.isAuthenticated).toBe(true);
  });

  it("login rejects empty credentials", async () => {
    const auth = useAuthStore();
    expect(await auth.login("", "x")).toBe(false);
    expect(await auth.login("a", "")).toBe(false);
  });

  it("login throws on 401 with friendly message", async () => {
    const auth = useAuthStore();
    apiMock.post.mockRejectedValue(axiosError(401));
    await expect(auth.login("a", "b")).rejects.toThrow("Invalid username or password");
  });

  it("login throws generic on other errors", async () => {
    const auth = useAuthStore();
    apiMock.post.mockRejectedValue(axiosError(500));
    await expect(auth.login("a", "b")).rejects.toThrow("Authentication Failed");
  });

  it("logout clears state and localStorage", async () => {
    const auth = useAuthStore();
    apiMock.post.mockResolvedValue({ data: { token: "abc" } });

    await auth.login("a@b.c", "secret");
    expect(auth.isAuthenticated).toBe(true);

    auth.logout();
    expect(auth.isAuthenticated).toBe(false);
    expect(auth.user).toBeNull();
    expect(localStorage.getItem("auth")).toBeNull();
  });

  it("loadFromStorage restores user", () => {
    localStorage.setItem("auth", JSON.stringify({ email: "x@y.z", token: "abc" }));
    const auth = useAuthStore();
    auth.loadFromStorage();
    expect(auth.user?.email).toBe("x@y.z");
    expect(auth.getToken()).toBe("Bearer abc");
  });

  it("checkAuthentication logs out expired token", () => {
    localStorage.setItem("auth", JSON.stringify({ email: "x@y.z", token: makeJwt(-60) }));
    const auth = useAuthStore();
    auth.checkAuthentication();
    expect(auth.user).toBeNull();
    expect(localStorage.getItem("auth")).toBeNull();
  });

  it("checkAuthentication keeps valid token", () => {
    localStorage.setItem("auth", JSON.stringify({ email: "x@y.z", token: makeJwt(3600) }));
    const auth = useAuthStore();
    auth.checkAuthentication();
    expect(auth.user).not.toBeNull();
  });

  it("checkAuthentication with malformed storage logs out", () => {
    localStorage.setItem("auth", "not-json");
    const auth = useAuthStore();
    auth.checkAuthentication();
    expect(auth.user).toBeNull();
  });

  it("checkAuthentication with no stored auth is a no-op logout", () => {
    const auth = useAuthStore();
    auth.checkAuthentication();
    expect(auth.user).toBeNull();
  });
});