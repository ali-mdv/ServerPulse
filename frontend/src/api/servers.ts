import axios from "axios";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "@/stores/auth";

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

export interface ServerUsage {
  serverId: string;
  updatedAt: string;
  usage: SystemUsage;
}

export interface ServerProviderAvailability {
  available: boolean;
  updatedAt?: string;
}

export type ServerStatus = "online" | "down" | "unknown";

export interface Server {
  id: string;
  name: string;
  host: string;
  port: number;
  description?: string;
  status: ServerStatus;
  lastSeen?: string;
  createdAt: string;
  updatedAt: string;
  usage?: ServerUsage;
  providers?: Record<string, ServerProviderAvailability>;
}

export interface CreateServerData {
  name: string;
  host?: string;
  port?: number;
  description?: string;
}

export interface UpdateServerData {
  name?: string;
  host?: string;
  port?: number;
  description?: string;
}

function getApi() {
  return useApi();
}

function getAuthHeader() {
  const auth = useAuthStore();
  return { Authorization: auth.getToken() };
}

function extractError(err: unknown): string {
  if (axios.isAxiosError(err) && err.response) {
    const data = err.response.data as { error?: string };
    return data.error ?? `Request failed (${err.response.status})`;
  }
  return "Network error";
}

export async function fetchServers(): Promise<Server[]> {
  try {
    const { data } = await getApi().get<{ servers: Server[] }>("servers", {
      headers: getAuthHeader(),
    });
    return data.servers;
  } catch (err) {
    throw new Error(extractError(err));
  }
}

export async function fetchServer(id: string): Promise<Server | null> {
  try {
    const { data } = await getApi().get<{ server: Server }>(`servers/${id}`, {
      headers: getAuthHeader(),
    });
    return data.server;
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      return null;
    }
    throw new Error(extractError(err));
  }
}

export async function createServer(payload: CreateServerData): Promise<Server> {
  try {
    const { data } = await getApi().post<{ server: Server }>("servers", payload, {
      headers: getAuthHeader(),
    });
    return data.server;
  } catch (err) {
    throw new Error(extractError(err));
  }
}

export async function updateServer(
  id: string,
  payload: UpdateServerData,
): Promise<Server> {
  try {
    const { data } = await getApi().put<{ server: Server }>(
      `servers/${id}`,
      payload,
      {
        headers: getAuthHeader(),
      },
    );
    return data.server;
  } catch (err) {
    throw new Error(extractError(err));
  }
}

export async function deleteServer(id: string): Promise<void> {
  try {
    await getApi().delete(`servers/${id}`, {
      headers: getAuthHeader(),
    });
  } catch (err) {
    throw new Error(extractError(err));
  }
}

export async function generateServerApiKey(id: string): Promise<string> {
  try {
    const { data } = await getApi().post<{ server: Server; apiKey: string }>(
      `servers/${id}/api-key`,
      {},
      {
        headers: getAuthHeader(),
      },
    );
    return data.apiKey;
  } catch (err) {
    throw new Error(extractError(err));
  }
}
