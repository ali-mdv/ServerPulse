<template>
  <div class="space-y-6">
    <PageHeader
      :title="server?.name || 'Server details'"
      :subtitle="`Status: ${server?.status || 'unknown'}`"
    >
      <template #default>
        <Badge v-if="server?.status" :tone="statusTone">
          {{ server.status }}
        </Badge>
        <router-link to="/servers">
          <Button variant="outline" size="sm">
            <ChevronLeft class="w-4 h-4" aria-hidden="true" />
            Back to servers
          </Button>
        </router-link>
      </template>
    </PageHeader>

    <div
      v-if="loading"
      class="grid gap-3"
      role="status"
      aria-live="polite"
      aria-label="Loading server"
    >
      <Skeleton height="11rem" />
      <Skeleton height="11rem" />
    </div>

    <EmptyState
      v-else-if="!server"
      title="Server not found"
      description="Check the URL or pick a different server."
    />

    <div v-else class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card class="h-44">
          <h3 class="font-semibold mb-2">CPU Usage</h3>
          <v-chart :option="cpuOption" style="height: 120px; width: 100%" />
        </Card>
        <Card class="h-44">
          <h3 class="font-semibold mb-2">Memory Usage</h3>
          <v-chart :option="memOption" style="height: 120px; width: 100%" />
        </Card>
        <Card class="h-44">
          <h3 class="font-semibold mb-2">Network Traffic (KB/s)</h3>
          <v-chart :option="netOption" style="height: 120px; width: 100%" />
        </Card>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="lg:col-span-2 space-y-4">
          <Card>
            <h3 class="font-semibold mb-2">Overview</h3>
            <dl class="flex flex-wrap gap-x-6 gap-y-2 text-sm">
              <div>
                <dt class="text-muted-foreground">CPU</dt>
                <dd class="font-bold">{{ server.cpu }}%</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">Memory</dt>
                <dd class="font-bold">{{ server.mem }}%</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">Network</dt>
                <dd class="font-bold">{{ latestNet }} KB/s</dd>
              </div>
            </dl>
          </Card>

          <ServiceManagerPanel
            :services="server.pm2 || []"
            :containers="server.docker || []"
          />
        </div>

        <div class="space-y-4">
          <Card>
            <h3 class="font-semibold mb-2">Alerts</h3>
            <EmptyState
              v-if="alerts.length === 0"
              title="No alerts"
              description="Nothing has been triggered for this server."
              class="!border-0 !shadow-none !p-2"
            />
            <ul v-else class="divide-y divide-border">
              <li
                v-for="a in alerts"
                :key="a.id"
                class="py-2 first:pt-0 last:pb-0"
              >
                <div class="font-medium">{{ a.title }}</div>
                <div class="text-xs text-muted-foreground">{{ a.time }}</div>
              </li>
            </ul>
          </Card>

          <Card>
            <h3 class="font-semibold mb-2">Service Summary</h3>
            <dl class="text-sm space-y-1">
              <div class="flex justify-between">
                <dt class="text-muted-foreground">PM2 processes</dt>
                <dd class="font-medium">{{ server.pm2?.length || 0 }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">Docker containers</dt>
                <dd class="font-medium">{{ server.docker?.length || 0 }}</dd>
              </div>
            </dl>
          </Card>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, computed } from "vue";
import { useRoute } from "vue-router";
import { ChevronLeft } from "lucide-vue-next";
import { fetchServer } from "@/api/servers";
import ServiceManagerPanel from "@/components/ServiceManagerPanel.vue";
import Card from "@/components/Card.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import { useChartTheme } from "@/composables";

const chartTheme = useChartTheme();

const route = useRoute();
const id = route.params.id as string;
const server = ref<any | null>(null);
const loading = ref(true);

const alerts = ref([
  { id: "a1", title: "High CPU detected", time: "10:12" },
  { id: "a2", title: "Container restarted", time: "09:50" },
]);

const cpuHistory = ref<number[]>([]);
const memHistory = ref<number[]>([]);
const netHistory = ref<number[]>([]);
let timer: any = null;

const cpuChart = computed(() => ({
  labels: cpuHistory.value.map(() => ""),
  datasets: [
    {
      label: "CPU",
      data: cpuHistory.value.slice(),
      borderColor: "hsl(var(--critical))",
      backgroundColor: "hsla(var(--critical) / 0.08)",
      tension: 0.3,
    },
  ],
}));
const memChart = computed(() => ({
  labels: memHistory.value.map(() => ""),
  datasets: [
    {
      label: "Mem",
      data: memHistory.value.slice(),
      borderColor: "hsl(var(--info))",
      backgroundColor: "hsla(var(--info) / 0.08)",
      tension: 0.3,
    },
  ],
}));
const netChart = computed(() => ({
  labels: netHistory.value.map(() => ""),
  datasets: [
    {
      label: "Net",
      data: netHistory.value.slice(),
      borderColor: "hsl(var(--success))",
      backgroundColor: "hsla(var(--success) / 0.08)",
      tension: 0.3,
    },
  ],
}));

const cpuOption = computed(() => ({
  tooltip: {
    trigger: "axis",
    backgroundColor: chartTheme.value.popover,
    borderColor: chartTheme.value.border,
    textStyle: { color: chartTheme.value.text },
  },
  xAxis: {
    type: "category",
    data: cpuChart.value.labels,
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { show: false },
  },
  yAxis: { type: "value", show: false, splitLine: { show: false } },
  grid: { left: 0, right: 0, top: 10, bottom: 10 },
  series: [
    {
      name: cpuChart.value.datasets[0].label,
      type: "line",
      data: cpuChart.value.datasets[0].data,
      smooth: true,
      symbol: "none",
      lineStyle: { color: cpuChart.value.datasets[0].borderColor, width: 2 },
      areaStyle: { color: cpuChart.value.datasets[0].backgroundColor },
    },
  ],
}));

const memOption = computed(() => ({
  tooltip: {
    trigger: "axis",
    backgroundColor: chartTheme.value.popover,
    borderColor: chartTheme.value.border,
    textStyle: { color: chartTheme.value.text },
  },
  xAxis: {
    type: "category",
    data: memChart.value.labels,
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { show: false },
  },
  yAxis: { type: "value", show: false, splitLine: { show: false } },
  grid: { left: 0, right: 0, top: 10, bottom: 10 },
  series: [
    {
      name: memChart.value.datasets[0].label,
      type: "line",
      data: memChart.value.datasets[0].data,
      smooth: true,
      symbol: "none",
      lineStyle: { color: memChart.value.datasets[0].borderColor, width: 2 },
      areaStyle: { color: memChart.value.datasets[0].backgroundColor },
    },
  ],
}));

const netOption = computed(() => ({
  tooltip: {
    trigger: "axis",
    backgroundColor: chartTheme.value.popover,
    borderColor: chartTheme.value.border,
    textStyle: { color: chartTheme.value.text },
  },
  xAxis: {
    type: "category",
    data: netChart.value.labels,
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { show: false },
  },
  yAxis: { type: "value", show: false, splitLine: { show: false } },
  grid: { left: 0, right: 0, top: 10, bottom: 10 },
  series: [
    {
      name: netChart.value.datasets[0].label,
      type: "line",
      data: netChart.value.datasets[0].data,
      smooth: true,
      symbol: "none",
      lineStyle: { color: netChart.value.datasets[0].borderColor, width: 2 },
      areaStyle: { color: netChart.value.datasets[0].backgroundColor },
    },
  ],
}));

const latestNet = computed(() =>
  netHistory.value.length ? netHistory.value[netHistory.value.length - 1] : 0,
);

const statusTone = computed<"success" | "critical" | "neutral">(() => {
  const s = server.value?.status;
  if (s === "online") return "success";
  if (s === "down") return "critical";
  return "neutral";
});

onMounted(async () => {
  loading.value = true;
  const data = await fetchServer(id);
  server.value = data;

  const initialCpu = server.value?.cpu ?? 0;
  const initialMem = server.value?.mem ?? 0;
  const initialNet = 0;
  cpuHistory.value = Array.from({ length: 20 }, () => initialCpu);
  memHistory.value = Array.from({ length: 20 }, () => initialMem);
  netHistory.value = Array.from({ length: 20 }, () => initialNet);

  timer = setInterval(() => {
    if (!server.value) return;
    const cpu = Math.max(
      0,
      Math.min(100, (server.value.cpu ?? initialCpu) + Math.round((Math.random() - 0.5) * 8)),
    );
    const mem = Math.max(
      0,
      Math.min(100, (server.value.mem ?? initialMem) + Math.round((Math.random() - 0.5) * 6)),
    );
    const net = Math.max(0, Math.round(Math.random() * 200));
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

onBeforeUnmount(() => {
  if (timer) clearInterval(timer);
});
</script>
