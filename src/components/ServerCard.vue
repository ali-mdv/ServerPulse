<template>
  <div class="bg-card p-4 rounded border flex items-center justify-between">
    <div>
      <div class="font-bold text-lg">{{ server.name }} <span class="text-sm text-muted">• {{ server.status }}</span></div>
      <div class="text-sm text-muted">CPU: {{ server.cpu }}% • Mem: {{ server.mem }}%</div>
    </div>
    <div class="flex items-center gap-3">
      <router-link :to="`/server/${server.id}`" class="px-3 py-1 bg-primary text-primary-foreground rounded">View</router-link>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
import type { Server } from '@/api/servers';
const props = defineProps<{ server: Server }>();
const { server } = props;
</script>
