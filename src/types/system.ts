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

interface PM2Env {
  status: string;
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
  pm2_env: PM2Env;
  monit: PM2Monit;
}

export interface GetPm2ServicesList {
  processes: PM2Service[];
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
}

export const enum ServiceAction {
  START = "start",
  STOP = "stop",
  RESTART = "restart",
  LOGS = "logs",
}
