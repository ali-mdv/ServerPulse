<template>
  <div class="space-y-6">
    <section v-if="props.services.length">
      <header class="flex items-center justify-between mb-3">
        <h3 class="text-base font-semibold">PM2 Services</h3>
        <Badge tone="neutral">{{ props.services.length }}</Badge>
      </header>
      <div class="grid gap-3">
        <Pm2ServiceItem
          v-for="s in props.services"
          :key="s.pid"
          :service="s"
        />
      </div>
    </section>

    <section v-if="props.containers.length">
      <header class="flex items-center justify-between mb-3">
        <h3 class="text-base font-semibold">Docker Containers</h3>
        <Badge tone="neutral">{{ props.containers.length }}</Badge>
      </header>
      <div class="grid gap-3">
        <DockerContainerItem
          v-for="c in props.containers"
          :key="c.id"
          :container="c"
        />
      </div>
    </section>

    <EmptyState
      v-if="!props.loading && !props.services.length && !props.containers.length"
      title="No services running"
      description="Once this server starts managing processes or containers they'll appear here."
    />

    <div v-if="props.loading" class="grid gap-3">
      <Skeleton v-for="i in 3" :key="i" height="5.5rem" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import Pm2ServiceItem from "@/components/Pm2ServiceItem.vue";
import DockerContainerItem from "@/components/DockerContainerItem.vue";
import { PM2Service, DockerContainer } from "@/types";
import EmptyState from "@/components/ui/EmptyState.vue";
import Badge from "@/components/ui/Badge.vue";
import Skeleton from "@/components/ui/Skeleton.vue";

const props = defineProps<{
  services: PM2Service[];
  containers: DockerContainer[];
  loading?: boolean;
}>();
</script>
