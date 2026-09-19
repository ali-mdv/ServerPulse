import axios from "axios";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "@/stores/auth";
import {
  GetServiceHistoryApi,
  GetTrackedServicesApi,
  HistoryProvider,
  ServiceSnapshot,
  TrackedService,
} from "@/types/system";

export type HistoryRange = "1h" | "6h" | "24h" | "7d";

const RANGE_SECONDS: Record<HistoryRange, number> = {
  "1h": 60 * 60,
  "6h": 6 * 60 * 60,
  "24h": 24 * 60 * 60,
  "7d": 7 * 24 * 60 * 60,
};

const RANGE_BUCKET: Record<HistoryRange, number> = {
  "1h": 60,
  "6h": 5 * 60,
  "24h": 15 * 60,
  "7d": 60 * 60,
};

export function rangeBucketSeconds(range: HistoryRange): number {
  return RANGE_BUCKET[range];
}

export function rangeSeconds(range: HistoryRange): number {
  return RANGE_SECONDS[range];
}

export async function fetchTrackedServices(
  provider: HistoryProvider,
): Promise<TrackedService[]> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.get<GetTrackedServicesApi>(
      `history/${provider}/services`,
      { headers: { Authorization: auth.getToken() } },
    );
    return data.services;
  } catch (err) {
    if (axios.isAxiosError(err) && err.response) {
      const msg = (err.response.data as { error?: string })?.error;
      throw new Error(msg ?? `Failed to fetch tracked ${provider} services`);
    }
    throw new Error(`Failed to fetch tracked ${provider} services`);
  }
}

export async function fetchServiceHistory(
  provider: HistoryProvider,
  serviceId: string,
  range: HistoryRange,
): Promise<ServiceSnapshot[]> {
  const api = useApi();
  const auth = useAuthStore();
  const now = new Date();
  const from = new Date(now.getTime() - RANGE_SECONDS[range] * 1000);
  const bucket = RANGE_BUCKET[range];
  try {
    const { data } = await api.get<GetServiceHistoryApi>(
      `history/${provider}/services/${encodeURIComponent(serviceId)}`,
      {
        params: {
          from: from.toISOString(),
          to: now.toISOString(),
          bucket,
        },
        headers: { Authorization: auth.getToken() },
      },
    );
    return data.points;
  } catch (err) {
    if (axios.isAxiosError(err) && err.response) {
      const msg = (err.response.data as { error?: string })?.error;
      throw new Error(msg ?? `Failed to fetch history for ${serviceId}`);
    }
    throw new Error(`Failed to fetch history for ${serviceId}`);
  }
}