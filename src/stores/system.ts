import axios from "axios";
import { defineStore } from "pinia";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "./auth";
import {
  SystemUsage,
  GetSystemUsageApi,
  PM2Service,
  GetPm2ServicesList,
  DockerContainer,
  GetDockerContainersList,
  GetContainerLogs,
} from "@/types/system";

export const useSystemStore = defineStore("system", () => {
  const api = useApi();
  const authStore = useAuthStore();

  // const cpuUsageList: Number[] = [];

  async function fetchSystemUsage(): Promise<SystemUsage> {
    try {
      const response = await api.get<GetSystemUsageApi>("system/monit", {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
      const cpuUsage = response.data.systemUsage.cpuUsage;
      // cpuUsageList.push(cpuUsage);
      return response.data.systemUsage;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === 401) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch system usage info");
    }
  }

  async function fetchPM2Services(): Promise<PM2Service[]> {
    try {
      const response = await api.get<GetPm2ServicesList>("pm2/services", {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
      return response.data.processes;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === 401) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch pm2 processes list");
    }
  }

  async function fetchDockerContainers(): Promise<DockerContainer[]> {
    try {
      const response = await api.get<GetDockerContainersList>(
        "docker/containers",
        {
          headers: {
            Authorization: authStore.getToken(),
          },
        },
      );
      return response.data.containers;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === 401) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to fetch docker containers list");
    }
  }

  async function startDockerContainer(container: DockerContainer) {
    try {
      await api.get(`docker/containers/${container.id}/start`, {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.Ok) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(`Unable to start container '${container.name}'.`);
    }
  }

  async function stopDockerContainer(container: DockerContainer) {
    try {
      await api.get(`docker/containers/${container.id}/stop`, {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.Ok) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(`Unable to stop container '${container.name}'.`);
    }
  }

  async function restartDockerContainer(container: DockerContainer) {
    try {
      await api.get(`docker/containers/${container.id}/restart`, {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.Ok) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(`Unable to restart container '${container.name}'.`);
    }
  }

  async function getContainerLogs(container: DockerContainer): Promise<string> {
    try {
      const response = await api.get<GetContainerLogs>(
        `docker/containers/${container.id}/logs`,
        {
          headers: {
            Authorization: authStore.getToken(),
          },
        },
      );
      return response.data.logs;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === axios.HttpStatusCode.Unauthorized) {
          throw new Error("Access denied");
        } else if (err.response.status === axios.HttpStatusCode.Ok) {
          throw new Error(`Container '${container.name}' was not found.`);
        }
      }
      throw new Error(
        `Unable to retrieve logs for container '${container.name}'.`,
      );
    }
  }

  return {
    fetchSystemUsage,
    fetchPM2Services,
    fetchDockerContainers,
    startDockerContainer,
    stopDockerContainer,
    restartDockerContainer,
    getContainerLogs,
  };
});
