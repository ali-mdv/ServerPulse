<template>
  <div class="space-y-6">
    <PageHeader
      title="Users"
      :subtitle="`${users.length} ${users.length === 1 ? 'user' : 'users'}`"
    >
      <template #default>
        <router-link to="/users/add">
          <Button variant="primary">
            <UserPlus class="w-4 h-4" aria-hidden="true" />
            Add user
          </Button>
        </router-link>
      </template>
    </PageHeader>

    <Card>
      <div
        v-if="loading"
        class="grid gap-3"
        role="status"
        aria-live="polite"
        aria-label="Loading users"
      >
        <Skeleton v-for="i in 3" :key="i" height="3rem" />
      </div>
      <EmptyState
        v-else-if="users.length === 0"
        title="No users yet"
        description="Add your first user to get started."
      />
      <ul v-else class="divide-y divide-border -my-2">
        <li
          v-for="u in users"
          :key="String(u.id)"
          class="flex items-center justify-between gap-3 py-3"
        >
          <div class="flex items-center gap-3 min-w-0">
            <span
              class="w-9 h-9 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-semibold shrink-0"
              aria-hidden="true"
            >
              {{ initials(u.email) }}
            </span>
            <div class="min-w-0">
              <div class="text-sm font-medium truncate">{{ u.email }}</div>
              <div class="text-xs text-muted-foreground">
                Joined {{ u.createdAt }}
              </div>
            </div>
          </div>
        </li>
      </ul>
    </Card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { useToast } from "primevue/usetoast";
import { UserPlus } from "lucide-vue-next";
import { useUsersStore } from "@/stores/users";
import { User } from "@/types";
import Card from "@/components/Card.vue";
import Button from "@/components/ui/Button.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Skeleton from "@/components/ui/Skeleton.vue";

const userStore = useUsersStore();

const toast = useToast();

const users = ref<Array<User>>([]);
const loading = ref(true);

function initials(email?: string) {
  if (!email) return "?";
  const name = email.split("@")[0] ?? "";
  return name.charAt(0).toUpperCase();
}

onMounted(async () => {
  try {
    users.value = await userStore.getUsers();
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  } finally {
    loading.value = false;
  }
});
</script>
