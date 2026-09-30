import { vi } from "vitest";
import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";

/**
 * axiosMock is a shared mock for the @/plugins/axios `useApi` factory.
 * Tests that exercise a Pinia store should `vi.mock("@/plugins/axios",
 * () => ({ useApi: () => axiosMock }))` and then set
 * `axiosMock.get.mockResolvedValue(...)` to program the response.
 */
export const axiosMock = {
  post: vi.fn(),
  get: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
};

/** authMock mirrors useAuthStore's getToken. */
export const authMock = {
  getToken: vi.fn(() => "Bearer stub"),
};

/**
 * axiosError constructs an AxiosError with the given status. Using a
 * real AxiosError (instead of a `{ isAxiosError: true, response: ... }`
 * stub) is required because the real `axios.isAxiosError` does an
 * instanceof check, not just a property lookup.
 */
export function axiosError(status: number, data?: unknown): AxiosError {
  const response = {
    status,
    data: data ?? {},
    statusText: "",
    headers: {},
    config: {} as InternalAxiosRequestConfig,
  };
  return new AxiosError("boom", String(status), {} as InternalAxiosRequestConfig, undefined, response);
}