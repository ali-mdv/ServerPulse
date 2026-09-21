import { describe, it, expect } from "vitest";
import {
  snapshotToPM2Service,
  snapshotsToPM2Services,
  snapshotToDockerContainer,
  snapshotsToDockerContainers,
} from "@/lib/service-snapshot";
import { PM2ServiceState, DockerContainerState } from "@/types/system";
import type { ServiceSnapshot } from "@/types/system";

describe("snapshotToPM2Service", () => {
  it("maps an online snapshot to PM2ServiceState.ONLINE", () => {
    const s: ServiceSnapshot = {
      ts: "2024-01-01T00:00:00Z",
      meta: { provider: "pm2", serviceId: "0", name: "api", serverId: "local" },
      status: "online",
      available: true,
      cpu: 5.0,
      memory: 1024,
    };
    const got = snapshotToPM2Service(s);
    expect(got.name).toBe("api");
    expect(got.pm_id).toBe("0");
    expect(got.pm2_env.status).toBe(PM2ServiceState.ONLINE);
    expect(got.monit.cpu).toBe(5.0);
    expect(got.monit.memory).toBe(1024);
  });

  it("maps a stopped snapshot to PM2ServiceState.STOPPED", () => {
    const s: ServiceSnapshot = {
      ts: "2024-01-01T00:00:00Z",
      meta: { provider: "pm2", serviceId: "0", name: "api", serverId: "local" },
      status: "stopped",
      available: false,
    };
    expect(snapshotToPM2Service(s).pm2_env.status).toBe(PM2ServiceState.STOPPED);
  });

  it("falls back to STOPPED for unknown status strings", () => {
    const s: ServiceSnapshot = {
      ts: "2024-01-01T00:00:00Z",
      meta: { provider: "pm2", serviceId: "0", name: "x", serverId: "local" },
      status: "launching",
      available: false,
    };
    expect(snapshotToPM2Service(s).pm2_env.status).toBe(PM2ServiceState.STOPPED);
  });

  it("treats a missing status as stopped", () => {
    const s: ServiceSnapshot = {
      ts: "2024-01-01T00:00:00Z",
      meta: { provider: "pm2", serviceId: "0", name: "x", serverId: "local" },
      available: false,
    };
    expect(snapshotToPM2Service(s).pm2_env.status).toBe(PM2ServiceState.STOPPED);
  });
});

describe("snapshotsToPM2Services", () => {
  it("maps a list", () => {
    const snaps: ServiceSnapshot[] = [
      {
        ts: "2024-01-01T00:00:00Z",
        meta: { provider: "pm2", serviceId: "0", name: "a", serverId: "local" },
        status: "online",
        available: true,
      },
      {
        ts: "2024-01-01T00:00:00Z",
        meta: { provider: "pm2", serviceId: "1", name: "b", serverId: "local" },
        status: "stopped",
        available: false,
      },
    ];
    const got = snapshotsToPM2Services(snaps);
    expect(got).toHaveLength(2);
    expect(got[0].name).toBe("a");
    expect(got[1].name).toBe("b");
  });

  it("returns [] for an empty input", () => {
    expect(snapshotsToPM2Services([])).toEqual([]);
  });
});

describe("snapshotToDockerContainer", () => {
  const baseSnapshot: ServiceSnapshot = {
    ts: "2024-01-01T00:00:00Z",
    meta: { provider: "docker", serviceId: "abc", name: "web", serverId: "local" },
    available: false,
  };

  it("passes through valid docker states", () => {
    for (const state of [
      DockerContainerState.RUNNING,
      DockerContainerState.EXITED,
      DockerContainerState.PAUSED,
      DockerContainerState.RESTARTING,
    ]) {
      const got = snapshotToDockerContainer({ ...baseSnapshot, status: state });
      expect(got.state).toBe(state);
    }
  });

  it("falls back to EXITED for unknown status", () => {
    const got = snapshotToDockerContainer({ ...baseSnapshot, status: "exploded" });
    expect(got.state).toBe(DockerContainerState.EXITED);
  });

  it("uses EXITED when status is missing", () => {
    const got = snapshotToDockerContainer(baseSnapshot);
    expect(got.state).toBe(DockerContainerState.EXITED);
  });

  it("preserves CPU/memory metrics", () => {
    const got = snapshotToDockerContainer({
      ...baseSnapshot,
      status: "running",
      cpu: 12.5,
      memory: 67.5,
      memUsage: 9999,
    });
    expect(got.usage?.cpuPercent).toBe(12.5);
    expect(got.usage?.memPercent).toBe(67.5);
    expect(got.usage?.memUsage).toBe(9999);
  });

  it("parses ts as a Date", () => {
    const got = snapshotToDockerContainer({ ...baseSnapshot });
    expect(got.createdAt).toBeInstanceOf(Date);
  });
});

describe("snapshotsToDockerContainers", () => {
  it("maps a list", () => {
    const snaps: ServiceSnapshot[] = [
      {
        ts: "2024-01-01T00:00:00Z",
        meta: { provider: "docker", serviceId: "a", name: "web", serverId: "local" },
        status: "running",
        available: true,
      },
      {
        ts: "2024-01-01T00:00:00Z",
        meta: { provider: "docker", serviceId: "b", name: "db", serverId: "local" },
        status: "exited",
        available: false,
      },
    ];
    const got = snapshotsToDockerContainers(snaps);
    expect(got).toHaveLength(2);
    expect(got[0].name).toBe("web");
    expect(got[1].state).toBe(DockerContainerState.EXITED);
  });
});