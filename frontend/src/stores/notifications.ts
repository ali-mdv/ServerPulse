import { ref } from "vue";
import { defineStore } from "pinia";
import type { Socket } from "socket.io-client";
import { createNotificationSocket } from "@/plugins/socket";
import {
  fetchNotifications,
  fetchUnreadCount,
  markAllNotificationsRead,
  markNotificationRead,
  type NotificationQuery,
} from "@/api/notifications";
import type { AppNotification, NotificationEnvelope } from "@/types";

const MAX_BUFFERED = 200;

export const useNotificationsStore = defineStore("notifications", () => {
  const items = ref<AppNotification[]>([]);
  const unreadCount = ref(0);
  const connected = ref(false);
  const loading = ref(false);
  const error = ref<string | null>(null);

  let socket: Socket | null = null;

  function upsert(notification: AppNotification) {
    const index = items.value.findIndex((n) => n.id === notification.id);
    if (index >= 0) {
      items.value[index] = notification;
    } else {
      items.value.unshift(notification);
      if (items.value.length > MAX_BUFFERED) {
        items.value = items.value.slice(0, MAX_BUFFERED);
      }
    }
  }

  /** Opens the Socket.IO connection. Safe to call repeatedly. */
  function connect() {
    if (socket) return;

    socket = createNotificationSocket();

    socket.on("connect", () => {
      connected.value = true;
      error.value = null;
      void load();
    });

    socket.on("disconnect", () => {
      connected.value = false;
    });

    socket.on("connect_error", (err: Error) => {
      connected.value = false;
      error.value = err.message;
    });

    socket.on("notification", (envelope: NotificationEnvelope) => {
      const notification = envelope?.notification;
      if (!notification) return;
      upsert(notification);
      if (!notification.read) unreadCount.value += 1;
    });
  }

  function disconnect() {
    socket?.disconnect();
    socket = null;
    connected.value = false;
  }

  async function load(query?: NotificationQuery) {
    loading.value = true;
    try {
      const [notifications, unread] = await Promise.all([
        fetchNotifications(query),
        fetchUnreadCount(query?.serverId),
      ]);
      items.value = notifications;
      unreadCount.value = unread;
      error.value = null;
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "Failed to load notifications";
    } finally {
      loading.value = false;
    }
  }

  async function markRead(id: string) {
    await markNotificationRead(id);
    const notification = items.value.find((n) => n.id === id);
    if (notification && !notification.read) {
      notification.read = true;
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    }
  }

  async function markAllRead(serverId?: string) {
    await markAllNotificationsRead(serverId);
    items.value = items.value.map((n) => ({ ...n, read: true }));
    unreadCount.value = 0;
  }

  return {
    items,
    unreadCount,
    connected,
    loading,
    error,
    connect,
    disconnect,
    load,
    markRead,
    markAllRead,
  };
});
