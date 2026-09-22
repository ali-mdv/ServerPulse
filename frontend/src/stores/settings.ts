import { ref } from "vue";
import { defineStore } from "pinia";
import { fetchSettings, updateSettings } from "@/api/settings";
import type { AppSettings, UpdateSettingsPayload } from "@/types";

export const useSettingsStore = defineStore("settings", () => {
  const settings = ref<AppSettings | null>(null);
  const loading = ref(false);
  const saving = ref(false);
  const error = ref<string | null>(null);

  async function load() {
    loading.value = true;
    try {
      settings.value = await fetchSettings();
      error.value = null;
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "Failed to load settings";
    } finally {
      loading.value = false;
    }
  }

  async function save(payload: UpdateSettingsPayload) {
    saving.value = true;
    try {
      settings.value = await updateSettings(payload);
      error.value = null;
      return settings.value;
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "Failed to save settings";
      throw err;
    } finally {
      saving.value = false;
    }
  }

  return { settings, loading, saving, error, load, save };
});