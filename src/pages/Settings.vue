<template>
  <div class="space-y-6 w-full max-w-full">
    <h2 class="text-2xl font-semibold">Settings</h2>

    <div class="bg-card p-4 rounded border">
      <h3 class="font-semibold mb-2">Appearance</h3>
      <div class="flex items-center gap-4">
        <label class="flex items-center gap-2">
          <input type="radio" value="system" v-model="theme" /> System
        </label>
        <label class="flex items-center gap-2">
          <input type="radio" value="light" v-model="theme" /> Light
        </label>
        <label class="flex items-center gap-2">
          <input type="radio" value="dark" v-model="theme" /> Dark
        </label>
      </div>
    </div>

    <div class="bg-card p-4 rounded border">
      <h3 class="font-semibold mb-2">Preferences</h3>
      <div class="flex items-center justify-between">
        <div>
          <div class="font-medium">Metrics refresh interval</div>
          <div class="text-sm text-muted">How frequently metrics auto-refresh</div>
        </div>
        <select v-model.number="refreshInterval" class="border p-2 rounded">
          <option :value="5">5s</option>
          <option :value="10">10s</option>
          <option :value="30">30s</option>
        </select>
      </div>
    </div>

    <div class="bg-card p-4 rounded border">
      <h3 class="font-semibold mb-2">Connections</h3>
      <div class="space-y-2">
        <div>
          <label class="block text-sm font-medium mb-1">WebSocket endpoint</label>
          <input v-model="wsEndpoint" class="w-full border p-2 rounded" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">REST API base</label>
          <input v-model="apiBase" class="w-full border p-2 rounded" />
        </div>
      </div>
    </div>

    <div class="flex gap-3">
      <button @click="save" class="px-4 py-2 bg-primary text-primary-foreground rounded">Save</button>
      <button @click="reset" class="px-4 py-2 border rounded">Reset</button>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted } from 'vue';
import { useToast } from 'primevue/usetoast';
const toast = useToast();

const theme = ref<'system'|'light'|'dark'>('system');
const refreshInterval = ref<number>(10);
const wsEndpoint = ref('wss://example.com/ws');
const apiBase = ref('https://api.example.com');

onMounted(() => {
  const saved = localStorage.getItem('settings');
  if (saved) {
    const s = JSON.parse(saved);
    theme.value = s.theme || 'system';
    refreshInterval.value = s.refreshInterval || 10;
    wsEndpoint.value = s.wsEndpoint || wsEndpoint.value;
    apiBase.value = s.apiBase || apiBase.value;
    applyTheme();
  }
});

watch(theme, applyTheme);

function applyTheme() {
  if (theme.value === 'dark') document.documentElement.classList.add('dark');
  else if (theme.value === 'light') document.documentElement.classList.remove('dark');
  else {
    // system
    const prefers = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (prefers) document.documentElement.classList.add('dark');
    else document.documentElement.classList.remove('dark');
  }
}

function save() {
  localStorage.setItem('settings', JSON.stringify({ theme: theme.value, refreshInterval: refreshInterval.value, wsEndpoint: wsEndpoint.value, apiBase: apiBase.value }));
  toast.add({ severity: 'success', summary: 'Saved', detail: 'Settings saved', life: 3000 });
}

function reset() {
  theme.value = 'system';
  refreshInterval.value = 10;
  wsEndpoint.value = 'wss://example.com/ws';
  apiBase.value = 'https://api.example.com';
  applyTheme();
}
</script>
