import axios from "axios";
import { defineStore } from "pinia";
import { ref } from "vue";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "./auth";
import {
  SystemUsage,
  GetSystemUsageApi,
  PM2Service,
  GetPm2ServicesList,
  DockerContainer,
  GetDockerContainersList,
  GetLogsApi,
  HistoryProvider,
  ServiceSnapshot,
  ProviderState,
} from "@/types/system";
import {
  fetchServiceHistory as apiFetchServiceHistory,
  HistoryRange,
} from "@/api/history";
import {
  fetchServicesState as apiFetchServicesState,
  fetchSystemState as apiFetchSystemState,
} from "@/api/state";

const SETTINGS_KEY = "settings";

/** Extra wait for remote agent round-trips (sync Socket.IO ack). */
const REMOTE_ACTION_TIMEOUT_MS = 15000;
const REMOTE_LOGS_TIMEOUT_MS = 25000;

function readStoredInterval(): number {
  if (typeof localStorage === "undefined") return 10000;
  try {
    const raw = localStorage.getItem(SETTINGS_KEY);
    if (!raw) return 10000;
    const parsed = JSON.parse(raw);
    const v = Number(parsed?.refreshInterval);
    if (!Number.isFinite(v) || v <= 0) return 10000;
    return v * 1000;
  } catch {
    return 10000;
  }
}

export const useSystemStore = defineStore("system", () => {
  const api = useApi();
  const authStore = useAuthStore();

  const intervalMS = ref<number>(readStoredInterval());

  /**
   * Builds the URL for a docker action. When serverId is present the
   * request is scoped to that server (local short-circuit or remote agent
   * over Socket.IO on the backend); otherwise it hits the legacy local-only
   * routes used by the Dashboard.
   */
  function dockerActionPath(
    container: DockerContainer,
    action: string,
    serverId?: string,
  ): { path: string; timeout?: number } {
    if (serverId) {
      return {
        path: `servers/${serverId}/docker/containers/${container.id}/${action}`,
        timeout:
          action === "logs" ? REMOTE_LOGS_TIMEOUT_MS : REMOTE_ACTION_TIMEOUT_MS,
      };
    }
    return { path: `docker/containers/${container.id}/${action}` };
  }

  function pm2ActionPath(
    service: PM2Service,
    action: string,
    serverId?: string,
  ): { path: string; timeout?: number } {
    if (serverId) {
      return {
        path: `servers/${serverId}/pm2/services/${service.pm_id}/${action}`,
        timeout:
          action === "logs" ? REMOTE_LOGS_TIMEOUT_MS : REMOTE_ACTION_TIMEOUT_MS,
      };
    }
    return { path: `pm2/services/${service.pm_id}/${action}` };
  }

  function mapActionError(
    err: unknown,
    kind: "container" | "service",
    name: string,
    action: string,
  ): Error {
    if (axios.isAxiosError(err)) {
      const status = err.response?.status;
      if (status === axios.HttpStatusCode.Unauthorized) {
        return new Error("Access denied");
      }
      if (status === axios.HttpStatusCode.NotFound) {
        return kind === "container"
          ? new Error(`Container '${name}' was not found.`)
          : new Error(`Service '${name}' was not found.`);
      }
      if (status === axios.HttpStatusCode.ServiceUnavailable) {
        return new Error(
          err.response?.data?.error ?? "Service unavailable on this server.",
        );
      }
      if (status === axios.HttpStatusCode.GatewayTimeout) {
        return new Error("Agent did not respond in time.");
      }
      const serverMsg = err.response?.data?.error;
      if (typeof serverMsg === "string" && serverMsg) {
        return new Error(serverMsg);
      }
    } else if (err instanceof Error && err.message) {
      return err;
    }
    return kind === "container"
      ? new Error(`Unable to ${action} container '${name}'.`)
      : new Error(`Unable to ${action} service '${name}'.`);
  }

  if (typeof window !== "undefined") {
    window.addEventListener("settings-updated", (e: Event) => {
      const detail = (e as CustomEvent<{ refreshInterval?: number }>).detail;
      if (detail?.refreshInterval && detail.refreshInterval > 0) {
        intervalMS.value = detail.refreshInterval * 1000;
      }
    });
    window.addEventListener("storage", (e) => {
      if (e.key === SETTINGS_KEY) {
        intervalMS.value = readStoredInterval();
      }
    });
  }

  async function fetchSystemUsage(): Promise<SystemUsage> {
    try {
      const response = await api.get<GetSystemUsageApi>("system/monit", {
        headers: { Authorization: authStore.getToken() },
      });
      return response.data.systemUsage;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch system usage info");
    }
  }

  async function fetchSystemState(serverId?: string): Promise<SystemUsage> {
    try {
      const response = await apiFetchSystemState(serverId);
      if (!response.systemUsage) {
        throw new Error("No system state recorded yet");
      }
      return response.systemUsage;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch system state");
    }
  }

  async function fetchServicesState(serverId?: string): Promise<{
    pm2: ProviderState;
    docker: ProviderState;
  }> {
    try {
      return await apiFetchServicesState(serverId);
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch services state");
    }
  }

  async function fetchPM2Services():
    Promise<{ services: PM2Service[]; available: boolean }> {
    try {
      const response = await api.get<GetPm2ServicesList>("pm2/services", {
        headers: { Authorization: authStore.getToken() },
      });
      return {
        services: response.data.processes,
        available: response.data.available,
      };
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch pm2 processes list");
    }
  }

  async function fetchDockerContainers():
    Promise<{ containers: DockerContainer[]; available: boolean }> {
    try {
      const response = await api.get<GetDockerContainersList>(
        "docker/containers",
        {
          headers: { Authorization: authStore.getToken() },
        },
      );
      return {
        containers: response.data.containers,
        available: response.data.available,
      };
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch docker containers list");
    }
  }

  async function startDockerContainer(
    container: DockerContainer,
    serverId?: string,
  ) {
    try {
      const { path, timeout } = dockerActionPath(container, "start", serverId);
      await api.get(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
    } catch (err) {
      throw mapActionError(err, "container", container.name, "start");
    }
  }

  async function stopDockerContainer(
    container: DockerContainer,
    serverId?: string,
  ) {
    try {
      const { path, timeout } = dockerActionPath(container, "stop", serverId);
      await api.get(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
    } catch (err) {
      throw mapActionError(err, "container", container.name, "stop");
    }
  }

  async function restartDockerContainer(
    container: DockerContainer,
    serverId?: string,
  ) {
    try {
      const { path, timeout } = dockerActionPath(container, "restart", serverId);
      await api.get(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
    } catch (err) {
      throw mapActionError(err, "container", container.name, "restart");
    }
  }

  async function getContainerLogs(
    container: DockerContainer,
    serverId?: string,
  ): Promise<string> {
    try {
      const { path, timeout } = dockerActionPath(container, "logs", serverId);
      const response = await api.get<GetLogsApi>(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
      return response.data.logs;
    } catch (err) {
      throw mapActionError(err, "container", container.name, "retrieve logs for");
    }
  }

  async function startPM2Service(service: PM2Service, serverId?: string) {
    try {
      const { path, timeout } = pm2ActionPath(service, "start", serverId);
      await api.get(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
    } catch (err) {
      throw mapActionError(err, "service", service.name, "start");
    }
  }

  async function stopPM2Service(service: PM2Service, serverId?: string) {
    try {
      const { path, timeout } = pm2ActionPath(service, "stop", serverId);
      await api.get(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
    } catch (err) {
      throw mapActionError(err, "service", service.name, "stop");
    }
  }

  async function restartPM2Service(service: PM2Service, serverId?: string) {
    try {
      const { path, timeout } = pm2ActionPath(service, "restart", serverId);
      await api.get(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
    } catch (err) {
      throw mapActionError(err, "service", service.name, "restart");
    }
  }

  async function getPM2ServiceLogs(
    service: PM2Service,
    serverId?: string,
  ): Promise<string> {
    try {
      const { path, timeout } = pm2ActionPath(service, "logs", serverId);
      const response = await api.get<GetLogsApi>(path, {
        headers: { Authorization: authStore.getToken() },
        timeout,
      });
      return response.data.logs;
    } catch (err) {
      throw mapActionError(err, "service", service.name, "retrieve logs for");
    }
  }

  return {
    intervalMS,
    fetchSystemUsage,
    fetchSystemState,
    fetchPM2Services,
    fetchDockerContainers,
    fetchServicesState,
    fetchServiceHistory: (
      provider: HistoryProvider,
      serviceId: string,
      range: HistoryRange,
      serverId?: string,
    ): Promise<ServiceSnapshot[]> =>
      apiFetchServiceHistory(provider, serviceId, range, serverId),
    startDockerContainer,
    stopDockerContainer,
    restartDockerContainer,
    getContainerLogs,
    startPM2Service,
    stopPM2Service,
    restartPM2Service,
    getPM2ServiceLogs,
  };
});
