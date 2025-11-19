<template>
  <div class="space-y-4">
    <div>
      <h3 class="text-lg font-semibold mb-2">PM2 Services</h3>
      <div class="grid gap-3">
        <ServiceItem v-for="s in pm2" :key="s.id" :service="s" />
      </div>
    </div>

    <div>
      <h3 class="text-lg font-semibold mb-2">Docker Containers</h3>
      <div class="grid gap-3">
        <ServiceItem v-for="s in docker" :key="s.id" :service="s" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import ServiceItem from '@/components/ServiceItem.vue';
const props = defineProps<{ pm2?: any[]; docker?: any[] }>();
const pm2 = props.pm2 ?? [];
const docker = props.docker ?? [];
</script>
