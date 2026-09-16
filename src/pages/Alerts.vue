<template>
  <div class="space-y-4">
    <h2 class="text-2xl font-semibold">Alerts & Events</h2>
    <div class="flex gap-3 items-center">
      <select v-model="severity" class="border p-2 rounded">
        <option value="all">All</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </select>
      <input type="date" v-model="from" class="border p-2 rounded" />
      <input type="date" v-model="to" class="border p-2 rounded" />
    </div>

    <div class="grid gap-3">
      <div v-for="alert in filtered" :key="alert.id" class="bg-card p-4 rounded border flex justify-between">
        <div>
          <div class="font-bold">{{ alert.title }}</div>
          <div class="text-sm text-muted">{{ alert.server }} • {{ alert.time }}</div>
        </div>
        <div class="flex items-center gap-2">
          <div :class="['px-2 py-1 rounded text-sm', alert.severity === 'critical' ? 'bg-red-100 text-red-700' : alert.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-800']">{{ alert.severity }}</div>
          <button v-if="!alert.resolved" @click="resolve(alert.id)" class="px-3 py-1 bg-primary text-primary-foreground rounded text-sm">Resolve</button>
          <div v-else class="text-sm text-muted">Resolved</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';

const severity = ref('all');
const from = ref('');
const to = ref('');

const alerts = ref([
  { id: 'a1', title: 'High CPU on web-01', server: 'web-01', time: '2025-10-24 10:12', severity: 'critical', resolved: false },
  { id: 'a2', title: 'Disk near full on db-01', server: 'db-01', time: '2025-10-24 09:50', severity: 'warning', resolved: false },
  { id: 'a3', title: 'Container restart on cache-01', server: 'cache-01', time: '2025-10-23 22:05', severity: 'info', resolved: true },
]);

const filtered = computed(() => {
  return alerts.value.filter(a => (severity.value === 'all' || a.severity === severity.value));
});

function resolve(id: string) {
  const found = alerts.value.find(a => a.id === id);
  if (found) found.resolved = true;
}
</script>
