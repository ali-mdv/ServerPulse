import { ref, computed } from "vue";
import { defineStore } from "pinia";
import { jwtDecode } from "jwt-decode";
import { useApi } from "@/plugins/axios";
import axios from "axios";

let timer: ReturnType<typeof setInterval> | null = null;

export const useAuthStore = defineStore("auth", () => {
  const api = useApi();
  const isAuthenticated = computed(() => !!user.value);
  const user = ref<{ email: string; token: string } | null>(null);

  async function login(email: string, password: string, remember = false) {
    if (!email || !password) return false;
    try {
      const response = await api.post("auth/login", { email, password });
      user.value = { email, token: response.data.token };
      if (remember)
        localStorage.setItem(
          "auth",
          JSON.stringify({ email, token: response.data.token }),
        );
      return;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === 401) {
          throw new Error("Invalid username or password");
        }
      }
      throw new Error("Authentication Failed");
    }
  }

  function logout() {
    user.value = null;
    localStorage.removeItem("auth");

    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  function loadFromStorage() {
    const stored = localStorage.getItem("auth");
    if (stored) user.value = JSON.parse(stored);
  }

  function checkAuthentication() {
    try {
      loadFromStorage();
      if (!user.value || !user.value.token) {
        logout();
        return;
      }

      const decoded: any = jwtDecode(user.value.token);

      // exp is in seconds → convert to milliseconds
      const msUntilExpire = decoded.exp * 1000 - Date.now();
      if (msUntilExpire <= 0) {
        logout();
        return;
      }

      // Set auto logout timer only once
      if (!timer) {
        timer = setTimeout(async () => {
          logout();
        }, msUntilExpire);
      }
    } catch {
      logout();
    }
  }

  function getToken() {
    if (user.value.token) {
      return `Bearer ${user.value.token}`;
    }
    return "";
  }

  return {
    isAuthenticated,
    user,
    login,
    logout,
    loadFromStorage,
    checkAuthentication,
    getToken,
  };
});
