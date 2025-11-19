<template>
  <div class="w-full">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-semibold">Users</h2>
      <router-link
        to="/users/add"
        class="px-3 py-2 bg-primary text-white rounded"
        >Add user</router-link
      >
    </div>

    <div class="bg-card p-4 rounded shadow">
      <div v-if="users.length === 0" class="text-muted">No users yet.</div>
      <ul class="space-y-2">
        <li
          v-for="(u, i) in users"
          :key="i + 1"
          class="flex items-center justify-between p-2 border rounded"
        >
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 bg-primary rounded-full flex items-center justify-center text-white font-medium"
            >
              {{ i + 1 }}
            </div>
            <div>
              <div class="text-sm text-muted">{{ u.email }}</div>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { useToast } from "primevue/usetoast";
import { useUsersStore } from "@/stores/users";
import { User } from "@/types";

const userStore = useUsersStore();

const toast = useToast();

const users = ref<Array<User>>([]);

onMounted(async () => {
  try {
    users.value = await userStore.getUsers();
  } catch (err) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  }
});
</script>
