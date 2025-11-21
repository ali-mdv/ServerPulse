<template>
  <div class="space-y-4">
    <div v-if="props.pm2.length">
      <h3 class="text-lg font-semibold mb-2">PM2 Services</h3>
      <div class="grid gap-3">
        <Pm2ServiceItem v-for="s in props.pm2" :key="s.pid" :service="s" />
      </div>
    </div>

    <div v-if="props.docker.length">
      <h3 class="text-lg font-semibold mb-2">Docker Containers</h3>
      <div class="grid gap-3">
        <ServiceItem v-for="s in props.docker" :key="s.id" :service="s" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import Pm2ServiceItem from "@/components/Pm2ServiceItem.vue";
import ServiceItem from "@/components/ServiceItem.vue";
import { PM2Service } from "@/types";

const props = defineProps<{ pm2: PM2Service[]; docker: any[] }>();
</script>
