import { io, type Socket } from "socket.io-client";
import { useAuthStore } from "@/stores/auth";
import { useRootStore } from "@/stores/root";

// NotificationSocketPath mirrors the backend constant
// (services.NotificationSocketPath) relative to the configured API base.
export const NOTIFICATION_SOCKET_SUFFIX = "/notifications/socket.io";

/**
 * Builds a Socket.IO client pointed at the backend notification stream.
 *
 * The token travels in the Socket.IO `auth` payload because browsers
 * can't set an Authorization header on the WebSocket handshake. The
 * origin/path are derived from the same API base the REST client uses,
 * so a configured VITE_SERVER_ADDRESS or a same-origin deploy both work.
 */
export function createNotificationSocket(): Socket {
  const root = useRootStore();
  const auth = useAuthStore();

  const apiUrl = new URL(root.getApiUrl(""), window.location.origin);
  const basePath = apiUrl.pathname.replace(/\/$/, "");

  return io(apiUrl.origin, {
    path: `${basePath}${NOTIFICATION_SOCKET_SUFFIX}`,
    transports: ["websocket", "polling"],
    auth: { token: auth.user?.token ?? "" },
    withCredentials: false,
    reconnection: true,
  });
}
