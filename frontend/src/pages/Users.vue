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
          <Button
            v-if="u.email !== authStore.user?.email"
            variant="outline"
            size="icon"
            aria-label="Delete user"
            @click="openDeleteModal(u)"
          >
            <Trash2 class="w-4 h-4" aria-hidden="true" />
          </Button>
        </li>
      </ul>
    </Card>

    <ConfirmModal
      :visible="showDeleteModal"
      title="Delete user"
      :message="deleteMessage"
      confirm-label="Delete"
      :loading="deleting"
      @update="showDeleteModal = $event"
      @confirm="executeDelete"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from "vue";
import { useToast } from "primevue/usetoast";
import { UserPlus, Trash2 } from "lucide-vue-next";
import { useUsersStore } from "@/stores/users";
import { useAuthStore } from "@/stores/auth";
import { User } from "@/types";
import Card from "@/components/Card.vue";
import Button from "@/components/ui/Button.vue";
import ConfirmModal from "@/components/ConfirmModal.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Skeleton from "@/components/ui/Skeleton.vue";

const userStore = useUsersStore();
const authStore = useAuthStore();

const toast = useToast();

const users = ref<Array<User>>([]);
const loading = ref(true);

const showDeleteModal = ref(false);
const deleting = ref(false);
const userToDelete = ref<User | null>(null);

const deleteMessage = computed(() =>
  userToDelete.value
    ? `Are you sure you want to delete "${userToDelete.value.email}"? This cannot be undone.`
    : "",
);

function initials(email?: string) {
  if (!email) return "?";
  const name = email.split("@")[0] ?? "";
  return name.charAt(0).toUpperCase();
}

function openDeleteModal(user: User) {
  if (user.email === authStore.user?.email) return;
  userToDelete.value = user;
  showDeleteModal.value = true;
}

async function executeDelete() {
  if (!userToDelete.value) return;
  if (userToDelete.value.email === authStore.user?.email) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: "You cannot delete your own account.",
      life: 3000,
    });
    showDeleteModal.value = false;
    userToDelete.value = null;
    return;
  }
  deleting.value = true;
  try {
    await userStore.deleteUser(String(userToDelete.value.id));
    toast.add({
      severity: "success",
      summary: "Deleted",
      detail: `User '${userToDelete.value.email}' removed.`,
      life: 3000,
    });
    showDeleteModal.value = false;
    users.value = await userStore.getUsers();
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  } finally {
    deleting.value = false;
    userToDelete.value = null;
  }
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
