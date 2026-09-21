import { describe, it, expect, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useRootStore } from "@/stores/root";

describe("root store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("getApiUrl joins baseUrl + path", () => {
    const store = useRootStore();
    store.serverAddress = "https://api.example.com";
    store.apiBaseUrl = "/api";
    expect(store.getApiUrl("users")).toBe("https://api.example.com/api/users");
  });

  it("strips trailing slash from baseUrl", () => {
    const store = useRootStore();
    store.serverAddress = "https://api.example.com";
    store.apiBaseUrl = "/api/";
    expect(store.getApiUrl("users")).toBe("https://api.example.com/api/users");
  });

  it("prepends a leading slash to a bare endpoint", () => {
    const store = useRootStore();
    store.serverAddress = "https://api.example.com";
    store.apiBaseUrl = "/api";
    expect(store.getApiUrl("/users")).toBe("https://api.example.com/api/users");
  });
});