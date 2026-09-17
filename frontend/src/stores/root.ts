import { defineStore } from "pinia";
import { ref, computed } from "vue";

export const useRootStore = defineStore("root", () => {
  const serverAddress = ref(import.meta.env.VITE_SERVER_ADDRESS);
  const apiBaseUrl = ref(import.meta.env.VITE_API_BASE_URL || "/api");
  const apiTimeout = ref(Number(import.meta.env.VITE_API_TIME_OUT) || 3000);
  const environment = ref<"development" | "staging" | "production">(
    (import.meta.env.VITE_ENVIRONMENT as
      | "development"
      | "staging"
      | "production") || "development",
  );
  const appName = ref(import.meta.env.VITE_APP_NAME || "ServerMon");
  const appVersion = ref(import.meta.env.VITE_APP_VERSION || "1.0.0");
  const isDevelopment = computed(() => environment.value === "development");
  const isProduction = computed(() => environment.value === "production");
  const isStaging = computed(() => environment.value === "staging");

  // Get full API URL for a specific endpoint
  function getApiUrl(endpoint: string): string {
    const base = apiBaseUrl.value.endsWith("/")
      ? apiBaseUrl.value.slice(0, -1)
      : apiBaseUrl.value;
    const path = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
    return `${serverAddress.value}${base}${path}`;
  }

  return {
    // State
    serverAddress,
    apiBaseUrl,
    apiTimeout,
    environment,
    appName,
    appVersion,
    // Computed
    isDevelopment,
    isProduction,
    isStaging,
    // Methods
    getApiUrl,
  };
});
