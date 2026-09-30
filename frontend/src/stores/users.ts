import axios from "axios";
import { defineStore } from "pinia";
import { useApi } from "@/plugins/axios";
import { GetUsersApi, CreateUserApi, ProfileApi, User } from "@/types";
import { useAuthStore } from "./auth";

export const useUsersStore = defineStore("users", () => {
  const api = useApi();
  const authStore = useAuthStore();

  async function getProfile(): Promise<User> {
    try {
      const response = await api.get<ProfileApi>("users/profile", {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
      return response.data.user;
    } catch (err) {
      authStore.logout();
      throw new Error("Access denied");
    }
  }

  async function updateProfile(email: string): Promise<User> {
    try {
      const response = await api.post<ProfileApi>(
        "users/profile",
        { email },
        {
          headers: {
            Authorization: authStore.getToken(),
          },
        },
      );
      return response.data.user;
    } catch (err) {
      authStore.logout();
      throw new Error("Access denied");
    }
  }

  async function getUsers(): Promise<User[]> {
    try {
      const response = await api.get<GetUsersApi>("users", {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
      return response.data.users;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response.status === 401) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to get users list");
    }
  }

  async function addUser(email: string, password: string): Promise<User> {
    try {
      const response = await api.post<CreateUserApi>(
        "users",
        { email, password },
        {
          headers: {
            Authorization: authStore.getToken(),
          },
        },
      );
      return response.data.user;
    } catch (err) {
      if (axios.isAxiosError(err)) {
        console.log(err.response.data);
        if (err.response.status === 401) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to create new user");
    }
  }

  async function deleteUser(id: string): Promise<void> {
    try {
      await api.delete(`users/${id}`, {
        headers: {
          Authorization: authStore.getToken(),
        },
      });
    } catch (err) {
      if (axios.isAxiosError(err)) {
        if (err.response?.status === 401) {
          throw new Error("Access denied");
        }
      }
      throw new Error("Failed to delete user");
    }
  }

  return {
    getProfile,
    updateProfile,
    getUsers,
    addUser,
    deleteUser,
  };
});
