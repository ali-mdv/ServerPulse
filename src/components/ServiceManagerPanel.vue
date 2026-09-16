<template>
  <div class="space-y-4">
    <div v-if="props.services.length">
      <h3 class="text-lg font-semibold mb-2">PM2 Services</h3>
      <div class="grid gap-3">
        <Pm2ServiceItem v-for="s in props.services" :key="s.pid" :service="s" />
      </div>
    </div>

    <div v-if="props.containers.length">
      <h3 class="text-lg font-semibold mb-2">Docker Containers</h3>
      <div class="grid gap-3">
        <DockerContainerItem
          v-for="c in props.containers"
          :key="c.id"
          :container="c"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import Pm2ServiceItem from "@/components/Pm2ServiceItem.vue";
import DockerContainerItem from "@/components/DockerContainerItem.vue";
import { PM2Service, DockerContainer } from "@/types";

const props = defineProps<{
  services: PM2Service[];
  containers: DockerContainer[];
}>();
</script>
