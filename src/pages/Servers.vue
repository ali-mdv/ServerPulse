<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-semibold">Servers</h2>
      <div class="flex items-center gap-4">
        <input v-model="q" placeholder="Search by name" class="border rounded px-3 py-1 bg-card" />
        <select v-model="filter" class="border rounded px-2 py-1 bg-card">
          <option value="all">All</option>
          <option value="online">Online</option>
          <option value="down">Down</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="p-6 bg-card rounded text-center">Loading servers...</div>
    <div v-else-if="servers.length === 0" class="p-6 bg-card rounded text-center">No servers available.</div>

    <div class="grid gap-4">
      <ServerCard v-for="s in filtered" :key="s.id" :server="s" />
    </div>

    <div v-if="backendMissing" class="text-sm text-muted mt-4">Note: No backend detected, showing sample data. Connect your API at /api/servers to show real servers.</div>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { fetchServers, type Server } from '@/api/servers';
import ServerCard from '@/components/ServerCard.vue';

const servers = ref<Server[]>([]);
const loading = ref(true);
const backendMissing = ref(false);
const q = ref('');
const filter = ref<'all'|'online'|'down'>('all');

onMounted(async () => {
  loading.value = true;
  const data = await fetchServers();
  servers.value = data;
  loading.value = false;
  // If all servers are sample (id starts with srv1..), mark backend missing when fetched result equals sample length
  backendMissing.value = servers.value.length > 0 && servers.value.every(s => s.id && s.id.startsWith('srv'));
});

const filtered = computed(() => {
  return servers.value.filter(s => {
    if (filter.value !== 'all' && s.status !== filter.value) return false;
    if (q.value && !s.name.toLowerCase().includes(q.value.toLowerCase())) return false;
    return true;
  });
});
</script>
