import { useRootStore } from "@/stores/root";
import axios from "axios";

export function useApi() {
  const rootStore = useRootStore();

  return axios.create({
    baseURL: rootStore.getApiUrl(""),
    timeout: rootStore.apiTimeout,
  });
}
