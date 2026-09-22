import axios from "axios";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "@/stores/auth";
import type {
  AppNotification,
  MarkAllReadResponse,
  NotificationsResponse,
  UnreadCountResponse,
} from "@/types";

export interface NotificationQuery {
  serverId?: string;
  severity?: string;
  unread?: boolean;
  limit?: number;
}

function paramsFor(query?: NotificationQuery) {
  if (!query) return undefined;
  const params: Record<string, string | number | boolean> = {};
  if (query.serverId) params.serverId = query.serverId;
  if (query.severity) params.severity = query.severity;
  if (query.unread) params.unread = true;
  if (query.limit) params.limit = query.limit;
  return params;
}

function toError(err: unknown, fallback: string): Error {
  if (axios.isAxiosError(err) && err.response) {
    const msg = (err.response.data as { error?: string })?.error;
    return new Error(msg ?? fallback);
  }
  return new Error(fallback);
}

export async function fetchNotifications(
  query?: NotificationQuery,
): Promise<AppNotification[]> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.get<NotificationsResponse>("notifications", {
      headers: { Authorization: auth.getToken() },
      params: paramsFor(query),
    });
    return data.notifications ?? [];
  } catch (err) {
    throw toError(err, "Failed to fetch notifications");
  }
}

export async function fetchUnreadCount(
  serverId?: string,
): Promise<number> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.get<UnreadCountResponse>(
      "notifications/unread-count",
      {
        headers: { Authorization: auth.getToken() },
        params: serverId ? { serverId } : undefined,
      },
    );
    return data.unread ?? 0;
  } catch (err) {
    throw toError(err, "Failed to fetch unread count");
  }
}

export async function markNotificationRead(id: string): Promise<void> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    await api.post(
      `notifications/${id}/read`,
      {},
      { headers: { Authorization: auth.getToken() } },
    );
  } catch (err) {
    throw toError(err, "Failed to mark notification read");
  }
}

export async function markAllNotificationsRead(
  serverId?: string,
): Promise<number> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.post<MarkAllReadResponse>(
      "notifications/read-all",
      {},
      {
        headers: { Authorization: auth.getToken() },
        params: serverId ? { serverId } : undefined,
      },
    );
    return data.updated ?? 0;
  } catch (err) {
    throw toError(err, "Failed to mark notifications read");
  }
}
