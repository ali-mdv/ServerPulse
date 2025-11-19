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
