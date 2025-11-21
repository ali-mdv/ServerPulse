<template>
  <div class="space-y-6">
    <!-- Window Size Debug -->
    <div class="text-sm text-muted">
      Window: {{ Math.round(width) }} × {{ Math.round(height) }}
    </div>

    <!-- Top Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card v-for="(item, i) in topStats" :key="i">
        <div>
          <div class="text-sm text-muted">{{ item.label }}</div>
          <div :class="['text-2xl font-bold', item.color]">
            {{ item.value }}
          </div>
        </div>
      </Card>
    </div>

    <!-- CPU -->
    <div class="grid grid-cols-1 gap-4">
      <Card>
        <h3 class="font-semibold mb-3">CPU Usage (Total)</h3>
        <v-chart :option="cpuOption" style="height: 260px; width: 100%" />
      </Card>
    </div>

    <!-- MEMORY & DISK -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <h3 class="font-semibold mb-3">Memory Usage</h3>
        <v-chart :option="memOption" style="height: 240px; width: 100%" />
      </Card>

      <Card>
        <h3 class="font-semibold mb-3">Disk Usage</h3>
        <v-chart :option="diskOption" style="height: 240px; width: 100%" />
      </Card>
    </div>

    <!-- Network Charts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <h3 class="font-semibold mb-2">Network Traffic Sent (KB/s)</h3>
        <v-chart :option="netSendOption" class="h-32 w-full" />
      </Card>

      <Card>
        <h3 class="font-semibold mb-2">Network Traffic Received (KB/s)</h3>
        <v-chart :option="netReceivedOption" class="h-32 w-full" />
      </Card>
    </div>

    <!-- Services List -->
    <ServiceManagerPanel :pm2="pm2Services" :docker="[]" />

    <!-- High Usage Servers -->
    <div>
      <h3 class="font-semibold mb-3">Servers with High Usage</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card v-for="s in highServers" :key="s.id">
          <div class="font-bold">{{ s.name }}</div>
          <div class="text-sm text-muted">
            CPU: {{ s.cpu }}% • Mem: {{ s.mem }}%
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, onMounted, ref } from "vue";
import { useWindowSize } from "@vueuse/core";
import { useToast } from "primevue/usetoast";
import Card from "@/components/Card.vue";
import { useSystemStore } from "@/stores/system";
import { UsageInfo, NetIOInfo, PM2Service } from "@/types";
import ServiceManagerPanel from "@/components/ServiceManagerPanel.vue";

const { width, height } = useWindowSize();
const toast = useToast();
const systemStore = useSystemStore();
const intervalMS = ref<number>(10000);

// ---------------------------
// Helpers
// ---------------------------
function parseBytes(value: number, base = 1024, decimals = 2) {
  if (value === 0) return { number: 0, unit: "Bytes" };

  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(value) / Math.log(base));
  const number = Number((value / Math.pow(base, index)).toFixed(decimals));

  return { number, unit: units[index] };
}

// ---------------------------
// Top Stats
// ---------------------------
const topStats = [
  { label: "Total Servers", value: 12 },
  { label: "Online", value: 10, color: "text-green-600" },
  { label: "Down", value: 2, color: "text-red-600" },
  { label: "High Load", value: 3, color: "text-yellow-600" },
];

// ---------------------------
// CPU Chart
// ---------------------------
const cpuData = reactive({
  labels: [],
  values: [],
});

const cpuOption = computed(() => ({
  tooltip: { trigger: "axis" },
  xAxis: { type: "category", data: cpuData.labels },
  yAxis: { type: "value" },
  series: [
    {
      name: "CPU %",
      type: "line",
      data: cpuData.values,
      smooth: true,
      lineStyle: { color: "#2563EB" },
      areaStyle: { color: "rgba(37,99,235,0.1)" },
    },
  ],
}));

function updateCpu(cpuUsage: number) {
  const time = new Date().toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });

  if (cpuData.labels.length >= 12) {
    cpuData.labels.shift();
    cpuData.values.shift();
  }

  cpuData.labels.push(time);
  cpuData.values.push(Math.round(cpuUsage));
}

// ---------------------------
// Memory Chart
// ---------------------------
const memOption = reactive({
  tooltip: { trigger: "item" },
  legend: { bottom: 0 },
  title: {
    text: "Total: 0 GB",
    left: "center",
    bottom: 15,
  },
  series: [
    {
      name: "Memory",
      type: "pie",
      radius: "75%",
      center: ["50%", "45%"],
      data: [
        { value: 0, name: "Used" },
        { value: 0, name: "Free" },
      ],
    },
  ],
});

function updateMemory(mem: UsageInfo) {
  const total = parseBytes(mem.total);
  const used = parseBytes(mem.used);
  const free = parseBytes(mem.total - mem.used);

  memOption.series[0].data[0].value = used.number;
  memOption.series[0].data[1].value = free.number;
  memOption.title.text = `Total: ${total.number} ${total.unit}`;
}

// ---------------------------
// Disk Chart
// ---------------------------
const diskOption = reactive({
  tooltip: { trigger: "item" },
  legend: { bottom: 0 },
  title: {
    text: "Total: 0 GB",
    left: "center",
    bottom: 15,
  },
  series: [
    {
      name: "Disk",
      type: "pie",
      radius: "75%",
      center: ["50%", "45%"],
      data: [
        { value: 0, name: "Used" },
        { value: 0, name: "Free" },
      ],
    },
  ],
});

function updateDiskUsageChart(diskUsage: UsageInfo) {
  const total = parseBytes(diskUsage.total, 1000);
  const used = parseBytes(diskUsage.used, 1000);
  const free = parseBytes(diskUsage.total - diskUsage.used, 1000);

  diskOption.series[0].data[0].value = used.number;
  diskOption.series[0].data[1].value = free.number;
  diskOption.title.text = `Total: ${total.number} ${total.unit}`;
}

// ---------------------------
// Network Charts (Shared Logic)
// ---------------------------
const netSent = ref<number[]>(Array(20).fill(0));
const netReceived = ref<number[]>(Array(20).fill(0));

function createNetOption(source: number[]) {
  return {
    tooltip: { trigger: "axis" },
    xAxis: { type: "category", data: source.map(() => "") },
    yAxis: { type: "value", show: false },
    grid: { left: 0, right: 0, top: 10, bottom: 10 },
    series: [
      {
        name: "Net",
        type: "line",
        data: source,
        smooth: true,
        lineStyle: { color: "#059669" },
        areaStyle: { color: "rgba(5,150,105,0.08)" },
      },
    ],
  };
}

const netSendOption = computed(() => createNetOption(netSent.value));
const netReceivedOption = computed(() => createNetOption(netReceived.value));

function updateNetwork(io: NetIOInfo) {
  const sent = Math.round(io.sent / 1024);
  const received = Math.round(io.received / 1024);

  if (netSent.value.length >= 20) netSent.value.shift();
  if (netReceived.value.length >= 20) netReceived.value.shift();

  netSent.value.push(sent);
  netReceived.value.push(received);
}

// ---------------------------
// High Usage Servers (static)
// ---------------------------

const pm2Services = ref<PM2Service[]>([]);

// ---------------------------
// High Usage Servers (static)
// ---------------------------
const highServers = [
  { id: "1", name: "web-01", cpu: 92, mem: 78 },
  { id: "2", name: "db-01", cpu: 88, mem: 85 },
  { id: "3", name: "cache-01", cpu: 79, mem: 61 },
];

// ---------------------------
// Data Fetching
// ---------------------------
async function loadSystemUsage() {
  try {
    const data = await systemStore.fetchSystemUsage();
    updateCpu(data.cpuUsage);
    updateMemory(data.memUsage);
    updateNetwork(data.netIO);
    updateDiskUsageChart(data.diskUsage);
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
    });
  }
}

async function loadPm2Services() {
  try {
    const data = await systemStore.fetchPM2Services();
    pm2Services.value = data;
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
    });
  }
}

onMounted(() => {
  loadSystemUsage();
  loadPm2Services();
  setInterval(loadSystemUsage, intervalMS.value);
  setInterval(loadPm2Services, intervalMS.value);
});
</script>
