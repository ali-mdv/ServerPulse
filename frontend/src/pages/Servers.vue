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
        </Select>
      </template>
    </PageHeader>

    <div
      v-if="backendMissing"
      class="card card-body border-warning/40 bg-warning/10 text-sm"
      role="status"
    >
      <strong class="font-semibold">No backend detected.</strong>
      Showing sample data. Connect your API at <code>/api/servers</code> to see
      real servers.
    </div>

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
      description="Once servers are registered they'll appear here."
    />

    <EmptyState
      v-else-if="filtered.length === 0"
      title="No matches"
      description="Try adjusting your search or filter."
    />

    <div v-else class="grid gap-3">
      <ServerCard v-for="s in filtered" :key="s.id" :server="s" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from "vue";
import { Search } from "lucide-vue-next";
import { fetchServers, type Server } from "@/api/servers";
import ServerCard from "@/components/ServerCard.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Skeleton from "@/components/ui/Skeleton.vue";

const servers = ref<Server[]>([]);
const loading = ref(true);
const backendMissing = ref(false);
const q = ref("");
const filter = ref<"all" | "online" | "down">("all");

onMounted(async () => {
  loading.value = true;
  const data = await fetchServers();
  servers.value = data;
  loading.value = false;
  backendMissing.value =
    servers.value.length > 0 &&
    servers.value.every((s) => s.id.startsWith("srv"));
});

const filtered = computed(() => {
  return servers.value.filter((s) => {
    if (filter.value !== "all" && s.status !== filter.value) return false;
    if (q.value && !s.name.toLowerCase().includes(q.value.toLowerCase()))
      return false;
    return true;
  });
});
</script>
