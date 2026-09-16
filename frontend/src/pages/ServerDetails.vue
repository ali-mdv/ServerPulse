<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-semibold">{{ server?.name || 'Server details' }}</h2>
        <div class="text-sm text-muted">Status: {{ server?.status || 'unknown' }}</div>
      </div>
      <div>
        <router-link to="/servers" class="px-3 py-1 border rounded">Back</router-link>
      </div>
    </div>

    <div v-if="loading" class="p-6 bg-card rounded">Loading...</div>
    <div v-else-if="!server" class="p-6 bg-card rounded">Server not found.</div>

    <div v-else class="space-y-4">
      <!-- Global charts -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="col-span-1 bg-card p-4 rounded border h-44">
          <h3 class="font-semibold mb-2">CPU Usage</h3>
          <v-chart :option="cpuOption" style="height:120px; width:100%" />
        </div>
        <div class="col-span-1 bg-card p-4 rounded border h-44">
          <h3 class="font-semibold mb-2">Memory Usage</h3>
          <v-chart :option="memOption" style="height:120px; width:100%" />
        </div>
        <div class="col-span-1 bg-card p-4 rounded border h-44">
          <h3 class="font-semibold mb-2">Network Traffic (KB/s)</h3>
          <v-chart :option="netOption" style="height:120px; width:100%" />
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="col-span-2 space-y-4">
          <div class="bg-card p-4 rounded border">
            <h3 class="font-semibold mb-2">Overview</h3>
            <div class="flex gap-4">
              <div>CPU: <span class="font-bold">{{ server.cpu }}%</span></div>
              <div>Memory: <span class="font-bold">{{ server.mem }}%</span></div>
              <div>Network: <span class="font-bold">{{ latestNet }} KB/s</span></div>
            </div>
          </div>

          <ServiceManagerPanel :pm2="server.pm2" :docker="server.docker" />
        </div>

        <div class="space-y-4">
          <div class="bg-card p-4 rounded border">
            <h3 class="font-semibold mb-2">Alerts</h3>
            <div v-for="a in alerts" :key="a.id" class="p-2 border-b last:border-b-0">
              <div class="font-medium">{{ a.title }}</div>
              <div class="text-sm text-muted">{{ a.time }}</div>
            </div>
          </div>

          <div class="bg-card p-4 rounded border">
            <h3 class="font-semibold mb-2">Service Summary</h3>
            <div>Total PM2: {{ server.pm2?.length || 0 }}</div>
            <div>Total Docker: {{ server.docker?.length || 0 }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue';
import { useRoute } from 'vue-router';
import { fetchServer } from '@/api/servers';
import ServiceManagerPanel from '@/components/ServiceManagerPanel.vue';

const route = useRoute();
const id = route.params.id as string;
const server = ref<any | null>(null);
const loading = ref(true);

const alerts = ref([
  { id: 'a1', title: 'High CPU detected', time: '10:12' },
  { id: 'a2', title: 'Container restarted', time: '09:50' }
]);

// histories
const cpuHistory = ref<number[]>([]);
const memHistory = ref<number[]>([]);
const netHistory = ref<number[]>([]);
let timer: any = null;

const smallOptions = { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false } }, scales: { y: { display: false }, x: { display: false } } };

const cpuChart = computed(() => ({ labels: cpuHistory.value.map((_,i)=>''), datasets: [{ label: 'CPU', data: cpuHistory.value.slice(), borderColor: '#ef4444', backgroundColor: 'rgba(239,68,68,0.08)', tension: 0.3 }] }));
const memChart = computed(() => ({ labels: memHistory.value.map((_,i)=>''), datasets: [{ label: 'Mem', data: memHistory.value.slice(), borderColor: '#2563EB', backgroundColor: 'rgba(37,99,235,0.08)', tension: 0.3 }] }));
const netChart = computed(() => ({ labels: netHistory.value.map((_,i)=>''), datasets: [{ label: 'Net', data: netHistory.value.slice(), borderColor: '#059669', backgroundColor: 'rgba(5,150,105,0.08)', tension: 0.3 }] }));

const cpuOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: cpuChart.value.labels },
  yAxis: { type: 'value', show: false },
  grid: { left: 0, right: 0, top: 10, bottom: 10 },
  series: [{ name: cpuChart.value.datasets[0].label, type: 'line', data: cpuChart.value.datasets[0].data, smooth: true, lineStyle: { color: cpuChart.value.datasets[0].borderColor }, areaStyle: { color: cpuChart.value.datasets[0].backgroundColor } }]
}));

const memOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: memChart.value.labels },
  yAxis: { type: 'value', show: false },
  grid: { left: 0, right: 0, top: 10, bottom: 10 },
  series: [{ name: memChart.value.datasets[0].label, type: 'line', data: memChart.value.datasets[0].data, smooth: true, lineStyle: { color: memChart.value.datasets[0].borderColor }, areaStyle: { color: memChart.value.datasets[0].backgroundColor } }]
}));

const netOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: netChart.value.labels },
  yAxis: { type: 'value', show: false },
  grid: { left: 0, right: 0, top: 10, bottom: 10 },
  series: [{ name: netChart.value.datasets[0].label, type: 'line', data: netChart.value.datasets[0].data, smooth: true, lineStyle: { color: netChart.value.datasets[0].borderColor }, areaStyle: { color: netChart.value.datasets[0].backgroundColor } }]
}));

const latestNet = computed(() => netHistory.value.length ? netHistory.value[netHistory.value.length-1] : 0);

onMounted(async () => {
  loading.value = true;
  const data = await fetchServer(id);
  server.value = data;

  // init histories with values
  const initialCpu = server.value?.cpu ?? 0;
  const initialMem = server.value?.mem ?? 0;
  const initialNet = 0;
  cpuHistory.value = Array.from({length:20}, ()=>initialCpu);
  memHistory.value = Array.from({length:20}, ()=>initialMem);
  netHistory.value = Array.from({length:20}, ()=>initialNet);

  // simulate updates every 2s
  timer = setInterval(()=>{
    if (!server.value) return;
    // update server metrics slightly
    const cpu = Math.max(0, Math.min(100, (server.value.cpu ?? initialCpu) + Math.round((Math.random()-0.5)*8)));
    const mem = Math.max(0, Math.min(100, (server.value.mem ?? initialMem) + Math.round((Math.random()-0.5)*6)));
    const net = Math.max(0, Math.round(Math.random()*200));
    server.value.cpu = cpu;
    server.value.mem = mem;

    cpuHistory.value.push(cpu);
    memHistory.value.push(mem);
    netHistory.value.push(net);
    if (cpuHistory.value.length > 20) cpuHistory.value.shift();
    if (memHistory.value.length > 20) memHistory.value.shift();
    if (netHistory.value.length > 20) netHistory.value.shift();
  }, 2000);

  loading.value = false;
});

onBeforeUnmount(()=>{
  if (timer) clearInterval(timer);
});
</script>
