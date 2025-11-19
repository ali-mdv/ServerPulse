<template>
  <div class="space-y-6">
    <div class="text-sm text-muted">
      Window: {{ Math.round(width) }} × {{ Math.round(height) }}
    </div>
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card>
        <div>
          <div class="text-sm text-muted">Total Servers</div>
          <div class="text-2xl font-bold">12</div>
        </div>
      </Card>
      <Card>
        <div>
          <div class="text-sm text-muted">Online</div>
          <div class="text-2xl font-bold text-green-600">10</div>
        </div>
      </Card>
      <Card>
        <div>
          <div class="text-sm text-muted">Down</div>
          <div class="text-2xl font-bold text-red-600">2</div>
        </div>
      </Card>
      <Card>
        <div>
          <div class="text-sm text-muted">High Load</div>
          <div class="text-2xl font-bold text-yellow-600">3</div>
        </div>
      </Card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <Card class="col-span-2">
        <h3 class="font-semibold mb-3">CPU Usage (Total)</h3>
        <v-chart :option="cpuOption" style="height: 260px; width: 100%" />
      </Card>

      <Card>
        <h3 class="font-semibold mb-3">Memory Usage</h3>
        <v-chart :option="memOption" style="height: 220px; width: 100%" />
      </Card>
    </div>

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

<script lang="ts" setup>
import Card from "@/components/Card.vue";
import { reactive, computed, onMounted } from "vue";
import { useToast } from "primevue/usetoast";
import { useWindowSize } from "@vueuse/core";
import { useSystemStore } from "@/stores/system";
import { UsageInfo } from "@/types";

const { width, height } = useWindowSize();

const systemStore = useSystemStore();

const toast = useToast();

const cpuData = reactive({
  labels: [],
  datasets: [
    {
      label: "CPU %",
      data: [],
      borderColor: "#2563EB",
      backgroundColor: "rgba(37,99,235,0.1)",
    },
  ],
});

const cpuOption = computed(() => ({
  tooltip: { trigger: "axis" },
  xAxis: { type: "category", data: cpuData.labels },
  yAxis: { type: "value" },
  series: [
    {
      name: cpuData.datasets[0].label,
      type: "line",
      data: cpuData.datasets[0].data,
      smooth: true,
      lineStyle: { color: cpuData.datasets[0].borderColor },
      areaStyle: { color: cpuData.datasets[0].backgroundColor },
    },
  ],
}));

const memOption = reactive({
  tooltip: { trigger: "item" },
  legend: { bottom: 0 },
  title: {
    text: "Total: 0 GB",
    left: "center",
    bottom: 15,
    textStyle: {
      fontSize: 14,
      fontWeight: "normal",
    },
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
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: "rgba(0,0,0,0.5)",
        },
      },
    },
  ],
});

const highServers = [
  { id: "1", name: "web-01", cpu: 92, mem: 78 },
  { id: "2", name: "db-01", cpu: 88, mem: 85 },
  { id: "3", name: "cache-01", cpu: 79, mem: 61 },
];

function parseBytes(bytes: number, decimals = 2) {
  if (bytes === 0) {
    return {
      number: 0,
      unit: "Bytes",
    };
  }

  const k = 1024;
  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  const number = parseFloat((bytes / Math.pow(k, i)).toFixed(decimals));
  const unit = units[i];

  return {
    number,
    unit,
  };
}

function updateCpuUsageChart(cpuUsage: number) {
  const timeLabel = new Date().toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });

  if (cpuData.labels.length >= 12) {
    cpuData.labels.shift();
    cpuData.datasets[0].data.shift();
  }

  cpuData.labels.push(timeLabel);
  cpuData.datasets[0].data.push(Math.round(cpuUsage));
}

function updateMemUsageChart(memUsage: UsageInfo) {
  const total = parseBytes(memUsage.total);
  const used = parseBytes(memUsage.used);
  const free = parseBytes(memUsage.total - memUsage.used);

  memOption.series[0].data[0].value = used.number;
  memOption.series[0].data[1].value = free.number;
  memOption.title.text = `Total: ${total.number} ${total.unit}`;
}

async function getSystemUsage() {
  try {
    const systemUsage = await systemStore.fetchSystemUsage();
    updateCpuUsageChart(systemUsage.cpuUsage);
    updateMemUsageChart(systemUsage.memUsage);
  } catch (err) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  }
}

onMounted(() => {
  getSystemUsage();
  setInterval(getSystemUsage, 10000);
});
</script>
