<template>
  <div class="space-y-6">
    <div v-if="props.searchable && hasAny" class="relative">
      <Search
        class="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
        aria-hidden="true"
      />
      <Input
        v-model="query"
        type="search"
        aria-label="Search services"
        placeholder="Search services and containers"
        class="pl-9"
      />
    </div>

    <section v-if="props.pm2Available && filteredServices.length">
      <header class="flex items-center justify-between mb-3">
        <h3 class="text-base font-semibold">PM2 Services</h3>
        <Badge tone="neutral">{{ filteredServices.length }}</Badge>
      </header>
      <div class="grid gap-3">
        <Pm2ServiceItem
          v-for="s in filteredServices"
          :key="s.pm_id"
          :service="s"
          :server-id="props.serverId"
        />
      </div>
    </section>

    <section v-if="props.dockerAvailable && filteredContainers.length">
      <header class="flex items-center justify-between mb-3">
        <h3 class="text-base font-semibold">Docker Containers</h3>
        <Badge tone="neutral">{{ filteredContainers.length }}</Badge>
      </header>
      <div class="grid gap-3">
        <DockerContainerItem
          v-for="c in filteredContainers"
          :key="c.id"
          :container="c"
          :server-id="props.serverId"
        />
      </div>
    </section>

    <EmptyState
      v-if="!props.loading && hasAny && !filteredServices.length && !filteredContainers.length"
      title="No matching services"
      description="Try adjusting your search term."
    />

    <EmptyState
      v-if="!props.loading && !hasAny"
      title="Services not available"
      description="No processes or containers are being managed right now. If a provider (PM2, Docker) is not installed or its socket is unreachable, its section is hidden."
    />

    <div v-if="props.loading" class="grid gap-3">
      <Skeleton v-for="i in 3" :key="i" height="5.5rem" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { Search } from "lucide-vue-next";
import { PM2Service, DockerContainer } from "@/types";
import Pm2ServiceItem from "@/components/Pm2ServiceItem.vue";
import DockerContainerItem from "@/components/DockerContainerItem.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import Badge from "@/components/ui/Badge.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import Input from "@/components/ui/Input.vue";

const props = defineProps<{
  services: PM2Service[];
  containers: DockerContainer[];
  pm2Available?: boolean;
  dockerAvailable?: boolean;
  loading?: boolean;
  /** Shows a search box that filters services and containers by name. */
  searchable?: boolean;
  /** Scopes History links and remote control actions. Omit for the local server. */
  serverId?: string;
}>();

const query = ref("");

const hasAny = computed(() => props.services.length > 0 || props.containers.length > 0);

const term = computed(() => query.value.trim().toLowerCase());

const filteredServices = computed(() => {
  if (!term.value) return props.services;
  return props.services.filter((s) => s.name.toLowerCase().includes(term.value));
});

const filteredContainers = computed(() => {
  if (!term.value) return props.containers;
  return props.containers.filter(
    (c) =>
      c.name.toLowerCase().includes(term.value) ||
      c.image.toLowerCase().includes(term.value),
  );
});
</script>
