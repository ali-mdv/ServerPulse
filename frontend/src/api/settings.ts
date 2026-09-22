import axios from "axios";
import { useApi } from "@/plugins/axios";
import { useAuthStore } from "@/stores/auth";
import type {
  AppSettings,
  SettingsResponse,
  UpdateSettingsPayload,
} from "@/types";

function toError(err: unknown, fallback: string): Error {
  if (axios.isAxiosError(err) && err.response) {
    const msg = (err.response.data as { error?: string })?.error;
    return new Error(msg ?? fallback);
  }
  return new Error(fallback);
}

export async function fetchSettings(): Promise<AppSettings> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.get<SettingsResponse>("settings", {
      headers: { Authorization: auth.getToken() },
    });
    return data.settings;
  } catch (err) {
    throw toError(err, "Failed to fetch settings");
  }
}

export async function updateSettings(
  payload: UpdateSettingsPayload,
): Promise<AppSettings> {
  const api = useApi();
  const auth = useAuthStore();
  try {
    const { data } = await api.put<SettingsResponse>("settings", payload, {
      headers: { Authorization: auth.getToken() },
    });
    return data.settings;
  } catch (err) {
    throw toError(err, "Failed to save settings");
  }
}