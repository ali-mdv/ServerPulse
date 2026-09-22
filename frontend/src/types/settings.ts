import type { Theme } from "@/composables/useTheme";

export type Appearance = Theme;

export interface AppSettings {
  appearance: Appearance;
  historyPollIntervalSeconds: number;
  historyRetentionSeconds: number;
  updatedAt: string;
}

export interface SettingsResponse {
  settings: AppSettings;
}

export interface UpdateSettingsPayload {
  appearance?: Appearance;
  historyPollIntervalSeconds?: number;
  historyRetentionSeconds?: number;
}