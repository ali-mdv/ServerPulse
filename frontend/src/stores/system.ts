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

  async function fetchSystemState(): Promise<SystemUsage> {
    try {
      const response = await apiFetchSystemState();
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

  async function fetchServicesState(): Promise<{
    pm2: ProviderState;
    docker: ProviderState;
  }> {
    try {
      return await apiFetchServicesState();
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

  async function startDockerContainer(container: DockerContainer) {
    try {
      await api.get(`docker/containers/${container.id}/start`, {
        headers: { Authorization: authStore.getToken() },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(`Unable to start container '${container.name}'.`);
    }
  }

  async function stopDockerContainer(container: DockerContainer) {
    try {
      await api.get(`docker/containers/${container.id}/stop`, {
        headers: { Authorization: authStore.getToken() },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(`Unable to stop container '${container.name}'.`);
    }
  }

  async function restartDockerContainer(container: DockerContainer) {
    try {
      await api.get(`docker/containers/${container.id}/restart`, {
        headers: { Authorization: authStore.getToken() },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(`Unable to restart container '${container.name}'.`);
    }
  }

  async function getContainerLogs(container: DockerContainer): Promise<string> {
    try {
      const response = await api.get<GetLogsApi>(
        `docker/containers/${container.id}/logs`,
        {
          headers: { Authorization: authStore.getToken() },
        },
      );
      return response.data.logs;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(
        `Unable to retrieve logs for container '${container.name}'.`,
      );
    }
  }

  async function startPM2Service(service: PM2Service) {
    try {
      await api.get(`pm2/services/${service.pm_id}/start`, {
        headers: { Authorization: authStore.getToken() },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Service '${service.name}' was not found.`);
        }
      }
      throw new Error(`Unable to start service '${service.name}'.`);
    }
  }

  async function stopPM2Service(service: PM2Service) {
    try {
      await api.get(`pm2/services/${service.pm_id}/stop`, {
        headers: { Authorization: authStore.getToken() },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Service '${service.name}' was not found.`);
        }
      }
      throw new Error(`Unable to stop service '${service.name}'.`);
    }
  }

  async function restartPM2Service(service: PM2Service) {
    try {
      await api.get(`pm2/services/${service.pm_id}/restart`, {
        headers: { Authorization: authStore.getToken() },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Service '${service.name}' was not found.`);
        }
      }
      throw new Error(`Unable to restart service '${service.name}'.`);
    }
  }

  async function getPM2ServiceLogs(service: PM2Service): Promise<string> {
    try {
      const response = await api.get<GetLogsApi>(
        `pm2/services/${service.pm_id}/logs`,
        {
          headers: { Authorization: authStore.getToken() },
        },
      );
      return response.data.logs;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.NotFound) {
          throw new Error(`Service '${service.name}' was not found.`);
        }
      }
      throw new Error(`Unable to retrieve logs for service '${service.name}'.`);
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
    ): Promise<ServiceSnapshot[]> =>
      apiFetchServiceHistory(provider, serviceId, range),
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
