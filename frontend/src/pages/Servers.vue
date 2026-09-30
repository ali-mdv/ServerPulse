<template>
  <div class="space-y-6">
    <PageHeader
      title="Servers"
      :subtitle="`${filtered.length} of ${servers.length} shown`"
    >
      <template #default>
        <div class="relative">
          <Search
            class="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
            aria-hidden="true"
          />
          <Input
            v-model="q"
            type="search"
            aria-label="Search servers"
            placeholder="Search by name"
            class="pl-9 w-56"
          />
        </div>
        <Select
          v-model="filter"
          aria-label="Filter by status"
          class="w-auto"
        >
          <option value="all">All</option>
          <option value="online">Online</option>
          <option value="down">Down</option>
          <option value="unknown">Unknown</option>
        </Select>
        <router-link to="/servers/new">
          <Button variant="primary" size="sm">
            <Plus class="w-4 h-4" aria-hidden="true" />
            Add server
          </Button>
        </router-link>
      </template>
    </PageHeader>

    <div
      v-if="loading"
      class="grid gap-3"
      role="status"
      aria-live="polite"
      aria-label="Loading servers"
    >
      <Skeleton v-for="i in 4" :key="i" height="5rem" />
    </div>

    <EmptyState
      v-else-if="servers.length === 0"
      title="No servers yet"
      description="Register a server to start monitoring it."
    >
      <router-link to="/servers/new">
        <Button variant="primary" size="sm">Add your first server</Button>
      </router-link>
    </EmptyState>

    <EmptyState
      v-else-if="filtered.length === 0"
      title="No matches"
      description="Try adjusting your search or filter."
    />

    <div v-else class="grid gap-3">
      <ServerCard
        v-for="s in filtered"
        :key="s.id"
        :server="s"
        @delete="openDeleteModal"
      />
    </div>

    <ConfirmModal
      :visible="showDeleteModal"
      title="Delete server"
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
import { Search, Plus } from "lucide-vue-next";
import { fetchServers, deleteServer, type Server } from "@/api/servers";
import ServerCard from "@/components/ServerCard.vue";
import ConfirmModal from "@/components/ConfirmModal.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import Button from "@/components/ui/Button.vue";

const toast = useToast();

const servers = ref<Server[]>([]);
const loading = ref(true);
const q = ref("");
const filter = ref<"all" | "online" | "down" | "unknown">("all");

const showDeleteModal = ref(false);
const deleting = ref(false);
const serverToDelete = ref<Server | null>(null);

const deleteMessage = computed(() =>
  serverToDelete.value
    ? `Are you sure you want to delete "${serverToDelete.value.name}"? This cannot be undone.`
    : "",
);

async function load() {
  loading.value = true;
  try {
    servers.value = await fetchServers();
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
}

onMounted(load);

const filtered = computed(() => {
  return servers.value.filter((s) => {
    if (filter.value !== "all" && s.status !== filter.value) return false;
    if (q.value && !s.name.toLowerCase().includes(q.value.toLowerCase()))
      return false;
    return true;
  });
});

function openDeleteModal(server: Server) {
  serverToDelete.value = server;
  showDeleteModal.value = true;
}

async function executeDelete() {
  if (!serverToDelete.value) return;
  deleting.value = true;
  try {
    await deleteServer(serverToDelete.value.id);
    toast.add({
      severity: "success",
      summary: "Deleted",
      detail: `Server '${serverToDelete.value.name}' removed.`,
      life: 3000,
    });
    showDeleteModal.value = false;
    await load();
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  } finally {
    deleting.value = false;
    serverToDelete.value = null;
  }
}
</script>
