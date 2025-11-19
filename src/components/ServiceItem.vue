<template>
  <div class="p-3 bg-card rounded border flex items-center justify-between">
    <div>
      <div class="font-medium">{{ service.name }} <span class="text-sm text-muted">• {{ service.status }}</span></div>
      <div class="text-sm text-muted">CPU: {{ service.cpu ?? 0 }}% • Mem: {{ service.mem ?? 0 }} MB</div>
    </div>
    <div class="text-right">
      <div class="text-sm">CPU: <span class="font-semibold">{{ service.cpu ?? 0 }}%</span></div>
      <div class="text-sm">Mem: <span class="font-semibold">{{ service.mem ?? 0 }} MB</span></div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import type { Service } from '@/api/servers';

const props = defineProps<{ service: Service }>();
const service = props.service;
</script>
