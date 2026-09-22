export type NotificationSeverity = "info" | "warning" | "critical";

export type NotificationType =
  | "server_down"
  | "server_up"
  | "service_down"
  | "service_up"
  | "threshold_exceeded"
  | "provider_unavailable"
  | "provider_available";

export interface AppNotification {
  id: string;
  serverId: string;
  type: NotificationType;
  severity: NotificationSeverity;
  title: string;
  message?: string;
  meta?: Record<string, unknown>;
  read: boolean;
  createdAt: string;
}

export interface NotificationEnvelope {
  type: "notification";
  notification: AppNotification;
}

export interface NotificationsResponse {
  notifications: AppNotification[];
}

export interface UnreadCountResponse {
  unread: number;
}

export interface MarkAllReadResponse {
  updated: number;
}
