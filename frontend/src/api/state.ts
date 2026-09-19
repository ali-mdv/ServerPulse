import axios from "axios";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "@/stores/auth";
import {
  GetServicesStateApi,
  GetSystemStateApi,
  ProviderState,
} from "@/types/system";

export async function fetchServicesState(): Promise<{
  pm2: ProviderState;
  docker: ProviderState;
}> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.get<GetServicesStateApi>("state/services", {
      headers: { Authorization: auth.getToken() },
    });
    return {
      pm2: data.pm2,
      docker: data.docker,
    };
  } catch (err) {
    if (axios.isAxiosError(err) && err.response) {
      const msg = (err.response.data as { error?: string })?.error;
      throw new Error(msg ?? "Failed to fetch services state");
    }
    throw new Error("Failed to fetch services state");
  }
}

export async function fetchSystemState(): Promise<GetSystemStateApi> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.get<GetSystemStateApi>("state/system", {
      headers: { Authorization: auth.getToken() },
    });
    return data;
  } catch (err) {
    if (axios.isAxiosError(err) && err.response) {
      const msg = (err.response.data as { error?: string })?.error;
      throw new Error(msg ?? "Failed to fetch system state");
    }
    throw new Error("Failed to fetch system state");
  }
}
