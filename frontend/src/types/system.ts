export interface UsageInfo {
  percent: number;
  used: number;
  total: number;
}

export interface NetIOInfo {
  sent: number;
  received: number;
}

export interface SystemUsage {
  memUsage: UsageInfo;
  cpuUsage: number;
  diskUsage: UsageInfo;
  netIO: NetIOInfo;
}

export interface GetSystemUsageApi {
  systemUsage: SystemUsage;
}

export const enum PM2ServiceState {
  ONLINE = "online",
  STOPPED = "stopped",
}

interface PM2Env {
  status: PM2ServiceState;
  pm_uptime: number;
  restart_time: number;
  unstable_restarts: number;
  pm_exec_path: string;
  pm_cwd: string;
  pmx_monitoring: boolean;
}

interface PM2Monit {
  cpu: number;
  memory: number;
}

export interface PM2Service {
  name: string;
  pid: number;
  pm_id: string;
  pm2_env: PM2Env;
  monit: PM2Monit;
}

export interface GetPm2ServicesList {
  processes: PM2Service[];
  available: boolean;
}

export const enum DockerContainerState {
  CREATED = "created",
  RUNNING = "running",
  PAUSED = "paused",
  RESTARTING = "restarting",
  REMOVING = "removing",
  EXITED = "exited",
  DEAD = "dead",
}

export interface DockerUsageContainer {
  cpuPercent: number;
  memPercent: number;
  memUsage: number;
}

export interface DockerContainer {
  id: string;
  name: string;
  image: string;
  port: string;
  state: DockerContainerState;
  upTime: string;
  usage?: DockerUsageContainer;
  createdAt: Date;
}

export interface GetDockerContainersList {
  containers: DockerContainer[];
  available: boolean;
}

export const enum ServiceAction {
  START = "start",
  STOP = "stop",
  RESTART = "restart",
  LOGS = "logs",
}

export interface GetLogsApi {
  logs: string;
}

export type HistoryProvider = "pm2" | "docker" | "system";

export type SystemMetricKey = "cpu" | "memory" | "disk" | "network";

export interface SnapshotMeta {
  serverId?: string;
  provider: HistoryProvider;
  serviceId: string;
  name: string;
}

export interface ServiceSnapshot {
  ts: string;
  meta: SnapshotMeta;
  status?: string;
  available: boolean;
  cpu?: number;
  memory?: number;
  memUsage?: number;
  disk?: number;
  netSent?: number;
  netRecv?: number;
}

export interface TrackedService {
  provider: HistoryProvider;
  serviceId: string;
  name: string;
}

export interface GetTrackedServicesApi {
  provider: HistoryProvider;
  services: TrackedService[];
}

export interface GetServiceHistoryApi {
  provider: HistoryProvider;
  serviceId: string;
  from: string;
  to: string;
  bucket: string;
  points: ServiceSnapshot[];
}

export interface ProviderState {
  provider: HistoryProvider;
  updatedAt: string;
  available: boolean;
  services: ServiceSnapshot[];
}

export interface GetServicesStateApi {
  pm2: ProviderState;
  docker: ProviderState;
}

export interface GetSystemStateApi {
  systemUsage: SystemUsage | null;
  updatedAt: string | null;
}
