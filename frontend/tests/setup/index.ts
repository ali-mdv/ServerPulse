/**
 * Shared test utilities for the ServerPulse frontend.
 *
 * Each helper is intentionally small and self-contained — the goal is
 * to reduce duplication without coupling unrelated test files. Importers
 * should still prefer setting up exactly what they need; the helpers
 * here exist to remove the boilerplate around the common cases.
 */

export { axiosMock, authMock, axiosError } from "./axios";

import { vi } from "vitest";

/**
 * localStorageMock returns a fresh Map-backed localStorage that
 * survives until clearStorage is called. Each call returns a new
 * instance so tests don't leak state across runs.
 */
export function localStorageMock(): {
  storage: Map<string, string>;
  restore(): void;
} {
  const store = new Map<string, string>();
  const original = (globalThis as any).localStorage;
  (globalThis as any).localStorage = {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    removeItem: (k: string) => void store.delete(k),
    clear: () => store.clear(),
    key: (i: number) => Array.from(store.keys())[i] ?? null,
    get length() { return store.size; },
  } as Storage;
  return {
    storage: store,
    restore() {
      (globalThis as any).localStorage = original;
    },
  };
}

/**
 * vi.resetModules is sometimes needed because some composables cache
 * module-level state. Re-export here so tests can find it next to
 * localStorageMock without reaching into vitest directly.
 */
export { vi };
export const resetModules = (): Promise<void> => vi.resetModules();