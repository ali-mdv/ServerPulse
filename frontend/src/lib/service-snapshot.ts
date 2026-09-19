import {
  ServiceSnapshot,
  PM2Service,
  PM2ServiceState,
  DockerContainer,
  DockerContainerState,
} from "@/types/system";

export function snapshotToPM2Service(s: ServiceSnapshot): PM2Service {
  const status = (s.status ?? "") as PM2ServiceState;
  return {
    name: s.meta.name,
    pid: 0,
    pm_id: s.meta.serviceId,
    pm2_env: {
      status:
        status === PM2ServiceState.ONLINE || status === PM2ServiceState.STOPPED
          ? status
          : PM2ServiceState.STOPPED,
      pm_uptime: 0,
      restart_time: 0,
      unstable_restarts: 0,
      pm_exec_path: "",
      pm_cwd: "",
      pmx_monitoring: false,
    },
    monit: {
      cpu: s.cpu ?? 0,
      memory: s.memory ?? 0,
    },
  };
}

export function snapshotsToPM2Services(snapshots: ServiceSnapshot[]): PM2Service[] {
  return snapshots.map(snapshotToPM2Service);
}

const validDockerStates: string[] = [
  "created",
  "running",
  "paused",
  "restarting",
  "removing",
  "exited",
  "dead",
];

export function snapshotToDockerContainer(s: ServiceSnapshot): DockerContainer {
  const state = s.status ?? "";
  const safeState = validDockerStates.includes(state)
    ? (state as DockerContainerState)
    : DockerContainerState.EXITED;

  return {
    id: s.meta.serviceId,
    name: s.meta.name,
    image: "",
    port: "",
    state: safeState,
    upTime: "",
    usage: {
      cpuPercent: s.cpu ?? 0,
      memPercent: s.memory ?? 0,
      memUsage: s.memUsage ?? 0,
    },
    createdAt: new Date(s.ts),
  };
}

export function snapshotsToDockerContainers(
  snapshots: ServiceSnapshot[],
): DockerContainer[] {
  return snapshots.map(snapshotToDockerContainer);
}
