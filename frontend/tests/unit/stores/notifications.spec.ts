import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";

const mocks = vi.hoisted(() => {
  const handlers = new Map<string, (...args: unknown[]) => void>();
  const socket = {
    on: vi.fn((event: string, cb: (...args: unknown[]) => void) => {
      handlers.set(event, cb);
      return socket;
    }),
    disconnect: vi.fn(),
  };
  const api = {
    fetchNotifications: vi.fn(),
    fetchUnreadCount: vi.fn(),
    markNotificationRead: vi.fn(),
    markAllNotificationsRead: vi.fn(),
  };
  return { handlers, socket, api };
});

vi.mock("@/plugins/socket", () => ({
  createNotificationSocket: () => mocks.socket,
}));

vi.mock("@/api/notifications", () => mocks.api);

import { useNotificationsStore } from "@/stores/notifications";
import type { AppNotification } from "@/types";

function notification(overrides: Partial<AppNotification> = {}): AppNotification {
  return {
    id: "n1",
    serverId: "s1",
    type: "service_down",
    severity: "critical",
    title: "api is down",
    read: false,
    createdAt: new Date().toISOString(),
    ...overrides,
  };
}

describe("notifications store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    mocks.handlers.clear();
    mocks.socket.on.mockClear();
    mocks.socket.disconnect.mockClear();
    Object.values(mocks.api).forEach((fn) => fn.mockReset());
  });

  it("registers socket listeners and loads on connect", async () => {
    mocks.api.fetchNotifications.mockResolvedValue([notification()]);
    mocks.api.fetchUnreadCount.mockResolvedValue(1);

    const store = useNotificationsStore();
    store.connect();

    expect(mocks.socket.on).toHaveBeenCalledWith("notification", expect.any(Function));

    mocks.handlers.get("connect")?.();
    await vi.waitFor(() => expect(store.items).toHaveLength(1));

    expect(store.connected).toBe(true);
    expect(store.unreadCount).toBe(1);
  });

  it("prepends a notification received over the socket", () => {
    const store = useNotificationsStore();
    store.connect();

    mocks.handlers.get("notification")?.({
      type: "notification",
      notification: notification({ id: "n2", title: "new alert" }),
    });

    expect(store.items[0].id).toBe("n2");
    expect(store.unreadCount).toBe(1);
  });

  it("markRead acknowledges and decrements the unread count", async () => {
    mocks.api.fetchNotifications.mockResolvedValue([notification()]);
    mocks.api.fetchUnreadCount.mockResolvedValue(1);
    mocks.api.markNotificationRead.mockResolvedValue(undefined);

    const store = useNotificationsStore();
    await store.load();
    await store.markRead("n1");

    expect(mocks.api.markNotificationRead).toHaveBeenCalledWith("n1");
    expect(store.items[0].read).toBe(true);
    expect(store.unreadCount).toBe(0);
  });

  it("markAllRead clears the unread count", async () => {
    mocks.api.markAllNotificationsRead.mockResolvedValue(2);

    const store = useNotificationsStore();
    store.items = [notification(), notification({ id: "n2" })];
    store.unreadCount = 2;

    await store.markAllRead();

    expect(store.unreadCount).toBe(0);
    expect(store.items.every((n) => n.read)).toBe(true);
  });
});
