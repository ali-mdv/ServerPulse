import axios from "axios";
import { defineStore } from "pinia";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "./auth";
import { SystemUsage, GetSystemUsageApi } from "@/types/system";

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

  return {
    fetchSystemUsage,
  };
});
