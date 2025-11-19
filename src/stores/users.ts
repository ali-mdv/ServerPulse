import axios from "axios";
import { defineStore } from "pinia";
import { useApi } from "@/plugins/axios";
import { GetUsersApi, CreateUserApi, User } from "@/types";
import { useAuthStore } from "./auth";

export const useUsersStore = defineStore("users", () => {
  const api = useApi();
  const authStore = useAuthStore();

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

  return {
    getUsers,
    addUser,
  };
});
