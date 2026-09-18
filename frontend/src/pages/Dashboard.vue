<template>
  <div class="space-y-6">
    <!-- Top Stats -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <Card v-for="(item, i) in topStats" :key="i">
        <div class="flex items-center justify-between gap-2">
          <span class="text-sm text-muted-foreground">{{ item.label }}</span>
          <component
            v-if="item.icon"
            :is="item.icon"
            class="w-4 h-4 text-muted-foreground"
            aria-hidden="true"
          />
        </div>
        <div :class="['text-2xl font-bold mt-1', item.color]">
          <template v-if="statsLoading">
            <Skeleton width="3rem" height="1.75rem" />
          </template>
          <template v-else>
            {{ item.value }}
          </template>
        </div>
      </Card>
    </div>

    <!-- CPU -->
    <Card>
      <h3 class="font-semibold mb-3">CPU Usage (Total)</h3>
      <div v-if="usageLoading" class="space-y-2">
        <Skeleton height="1rem" width="60%" />
        <Skeleton height="240px" />
      </div>
      <v-chart v-else :option="cpuOption" :autoresize="true" style="height: 260px; width: 100%" />
    </Card>

    <!-- MEMORY & DISK -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <h3 class="font-semibold mb-3">Memory Usage</h3>
        <div v-if="usageLoading" class="space-y-2">
          <Skeleton height="1rem" width="60%" />
          <Skeleton height="220px" />
        </div>
        <v-chart v-else :option="memOption" :autoresize="true" style="height: 240px; width: 100%" />
      </Card>

      <Card>
        <h3 class="font-semibold mb-3">Disk Usage</h3>
        <div v-if="usageLoading" class="space-y-2">
          <Skeleton height="1rem" width="60%" />
          <Skeleton height="220px" />
        </div>
        <v-chart v-else :option="diskOption" :autoresize="true" style="height: 240px; width: 100%" />
      </Card>
    </div>

    <!-- Network Charts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <h3 class="font-semibold mb-2">Network Sent (KB/s)</h3>
        <Skeleton v-if="usageLoading" height="120px" />
        <v-chart v-else :option="netSendOption" :autoresize="true" class="h-32 w-full" />
      </Card>

      <Card>
        <h3 class="font-semibold mb-2">Network Received (KB/s)</h3>
        <Skeleton v-if="usageLoading" height="120px" />
        <v-chart v-else :option="netReceivedOption" :autoresize="true" class="h-32 w-full" />
      </Card>
    </div>

    <!-- Services List -->
    <ServiceManagerPanel
      :services="pm2Services"
      :containers="dockerContainers"
      :loading="servicesLoading"
    />

    <!-- High Usage Servers -->
    <section v-if="highUsageServers.length">
      <header class="flex items-center justify-between mb-3">
        <h3 class="text-base font-semibold">Servers with High Usage</h3>
        <Badge tone="warning">{{ highUsageServers.length }}</Badge>
      </header>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <router-link
          v-for="s in highUsageServers"
          :key="s.id"
          :to="`/server/${s.id}`"
          class="block"
        >
          <Card class="hover:border-foreground/30 transition-colors">
            <div class="flex items-center justify-between gap-2">
              <div class="font-semibold truncate">{{ s.name }}</div>
              <Badge :tone="s.status === 'online' ? 'success' : 'critical'">
                {{ s.status }}
              </Badge>
            </div>
            <div class="text-sm text-muted-foreground mt-1">
              CPU: {{ s.cpu }}% • Mem: {{ s.mem }}%
            </div>
          </Card>
        </router-link>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, onMounted, ref, watchEffect } from "vue";
import { useToast } from "primevue/usetoast";
import { Server, Activity, AlertTriangle, CircleSlash } from "lucide-vue-next";
import Card from "@/components/Card.vue";
import { useSystemStore } from "@/stores/system";
import { UsageInfo, NetIOInfo, PM2Service, DockerContainer } from "@/types";
import ServiceManagerPanel from "@/components/ServiceManagerPanel.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import Badge from "@/components/ui/Badge.vue";
import { useChartTheme } from "@/composables";
import { fetchServers } from "@/api/servers";
import { withAlpha } from "@/lib/chart-theme";

const toast = useToast();
const systemStore = useSystemStore();
const chartTheme = useChartTheme();

function parseBytes(value: number, base = 1024, decimals = 2) {
  if (value === 0) return { number: 0, unit: "Bytes" };

  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(value) / Math.log(base));
  const number = Number((value / Math.pow(base, index)).toFixed(decimals));

  return { number, unit: units[index] };
}

const allServers = ref<Array<{ id: string; name: string; status: string; cpu: number; mem: number }>>([]);

const topStats = computed(() => {
  const total = allServers.value.length;
  const online = allServers.value.filter((s) => s.status === "online").length;
  const down = allServers.value.filter((s) => s.status === "down").length;
  const highLoad = allServers.value.filter(
    (s) => s.cpu >= 80 || s.mem >= 80,
  ).length;
  return [
    { label: "Total Servers", value: total, icon: Server, color: "text-foreground" },
    { label: "Online", value: online, icon: Activity, color: "text-success" },
    { label: "Down", value: down, icon: CircleSlash, color: "text-critical" },
    { label: "High Load", value: highLoad, icon: AlertTriangle, color: "text-warning" },
  ];
});

const statsLoading = ref(true);
const usageLoading = ref(true);
const servicesLoading = ref(true);

const cpuData = reactive({
  labels: Array(12).fill("-"),
  values: Array(12).fill(0),
});

const cpuOption = ref({
  animation: false,
  tooltip: {
    trigger: "axis",
    backgroundColor: chartTheme.value.popover,
    borderColor: chartTheme.value.border,
    borderWidth: 1,
    textStyle: { color: chartTheme.value.text },
    extraCssText:
      "box-shadow: 0 4px 12px rgba(0,0,0,0.25); border-radius: 6px;",
  },
  grid: { left: 40, right: 16, top: 16, bottom: 28 },
  xAxis: {
    type: "category",
    data: cpuData.labels,
    axisLine: { lineStyle: { color: chartTheme.value.border } },
    axisLabel: { color: chartTheme.value.textMuted, fontSize: 11 },
    axisTick: { lineStyle: { color: chartTheme.value.border } },
  },
  yAxis: {
    type: "value",
    splitLine: { lineStyle: { color: chartTheme.value.border, type: "dashed" } },
    axisLabel: { color: chartTheme.value.textMuted, fontSize: 11 },
  },
  series: [
    {
      name: "CPU %",
      type: "line",
      data: cpuData.values,
      smooth: true,
      symbol: "circle",
      symbolSize: 6,
      showSymbol: cpuData.values.some((v) => v > 0),
      lineStyle: { color: chartTheme.value.info, width: 2.5 },
      itemStyle: { color: chartTheme.value.info, borderColor: chartTheme.value.popover, borderWidth: 2 },
      areaStyle: {
        color: {
          type: "linear",
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: withAlpha(chartTheme.value.info, 0.33) },
            { offset: 1, color: withAlpha(chartTheme.value.info, 0) },
          ],
        },
      },
    },
  ],
});

watchEffect(() => {
  const theme = chartTheme.value;
  const opt = cpuOption.value;
  opt.tooltip.backgroundColor = theme.popover;
  opt.tooltip.borderColor = theme.border;
  opt.tooltip.textStyle.color = theme.text;
  opt.xAxis.data = cpuData.labels;
  opt.xAxis.axisLine.lineStyle.color = theme.border;
  opt.xAxis.axisLabel.color = theme.textMuted;
  opt.xAxis.axisTick.lineStyle.color = theme.border;
  opt.yAxis.splitLine.lineStyle.color = theme.border;
  opt.yAxis.axisLabel.color = theme.textMuted;
  opt.series[0].data = cpuData.values;
  opt.series[0].showSymbol = cpuData.values.some((v) => v > 0);
  opt.series[0].lineStyle.color = theme.info;
  opt.series[0].itemStyle.color = theme.info;
  opt.series[0].itemStyle.borderColor = theme.popover;
  opt.series[0].areaStyle.color.colorStops[0].color = withAlpha(theme.info, 0.33);
  opt.series[0].areaStyle.color.colorStops[1].color = withAlpha(theme.info, 0);
});

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

const memData = reactive({
  title: "Total: 0 GB",
  used: 0,
  free: 0,
});

const memOption = ref({
  animation: false,
  tooltip: {
    trigger: "item",
    backgroundColor: chartTheme.value.popover,
    borderColor: chartTheme.value.border,
    textStyle: { color: chartTheme.value.text },
  },
  legend: {
    bottom: 0,
    textStyle: { color: chartTheme.value.textMuted },
  },
  title: {
    text: memData.title,
    left: "center",
    bottom: 15,
    textStyle: { color: chartTheme.value.text, fontSize: 12, fontWeight: 500 },
  },
  series: [
    {
      name: "Memory",
      type: "pie",
      radius: "75%",
      center: ["50%", "45%"],
      data: [
        {
          value: memData.used,
          name: "Used",
          itemStyle: { color: chartTheme.value.info },
        },
        {
          value: memData.free,
          name: "Free",
          itemStyle: { color: chartTheme.value.free },
        },
      ],
      label: { color: chartTheme.value.text },
    },
  ],
});

watchEffect(() => {
  const theme = chartTheme.value;
  const opt = memOption.value;
  opt.tooltip.backgroundColor = theme.popover;
  opt.tooltip.borderColor = theme.border;
  opt.tooltip.textStyle.color = theme.text;
  opt.legend.textStyle.color = theme.textMuted;
  opt.title.text = memData.title;
  opt.title.textStyle.color = theme.text;
  opt.series[0].data[0].value = memData.used;
  opt.series[0].data[0].itemStyle.color = theme.info;
  opt.series[0].data[1].value = memData.free;
  opt.series[0].data[1].itemStyle.color = theme.free;
  opt.series[0].label.color = theme.text;
});

function updateMemory(mem: UsageInfo) {
  const total = parseBytes(mem.total);
  const used = parseBytes(mem.used);
  const free = parseBytes(mem.total - mem.used);

  memData.used = used.number;
  memData.free = free.number;
  memData.title = `Total: ${total.number} ${total.unit}`;
}

const diskData = reactive({
  title: "Total: 0 GB",
  used: 0,
  free: 0,
});

const diskOption = ref({
  animation: false,
  tooltip: {
    trigger: "item",
    backgroundColor: chartTheme.value.popover,
    borderColor: chartTheme.value.border,
    textStyle: { color: chartTheme.value.text },
  },
  legend: {
    bottom: 0,
    textStyle: { color: chartTheme.value.textMuted },
  },
  title: {
    text: diskData.title,
    left: "center",
    bottom: 15,
    textStyle: { color: chartTheme.value.text, fontSize: 12, fontWeight: 500 },
  },
  series: [
    {
      name: "Disk",
      type: "pie",
      radius: "75%",
      center: ["50%", "45%"],
      data: [
        {
          value: diskData.used,
          name: "Used",
          itemStyle: { color: chartTheme.value.warning },
        },
        {
          value: diskData.free,
          name: "Free",
          itemStyle: { color: chartTheme.value.free },
        },
      ],
      label: { color: chartTheme.value.text },
    },
  ],
});

watchEffect(() => {
  const theme = chartTheme.value;
  const opt = diskOption.value;
  opt.tooltip.backgroundColor = theme.popover;
  opt.tooltip.borderColor = theme.border;
  opt.tooltip.textStyle.color = theme.text;
  opt.legend.textStyle.color = theme.textMuted;
  opt.title.text = diskData.title;
  opt.title.textStyle.color = theme.text;
  opt.series[0].data[0].value = diskData.used;
  opt.series[0].data[0].itemStyle.color = theme.warning;
  opt.series[0].data[1].value = diskData.free;
  opt.series[0].data[1].itemStyle.color = theme.free;
  opt.series[0].label.color = theme.text;
});

function updateDiskUsageChart(diskUsage: UsageInfo) {
  const total = parseBytes(diskUsage.total, 1000);
  const used = parseBytes(diskUsage.used, 1000);
  const free = parseBytes(diskUsage.total - diskUsage.used, 1000);

  diskData.used = used.number;
  diskData.free = free.number;
  diskData.title = `Total: ${total.number} ${total.unit}`;
}

const netSent = ref<number[]>(Array(20).fill(0));
const netReceived = ref<number[]>(Array(20).fill(0));

function createNetOption(source: number[]) {
  return {
    animation: false,
    tooltip: {
      trigger: "axis",
      backgroundColor: chartTheme.value.popover,
      borderColor: chartTheme.value.border,
      textStyle: { color: chartTheme.value.text },
    },
    xAxis: {
      type: "category",
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { show: false },
    },
    yAxis: {
      type: "value",
      show: false,
      splitLine: { show: false },
    },
    grid: { left: 0, right: 0, top: 10, bottom: 10 },
    series: [
      {
        name: "Net",
        type: "line",
        data: source,
        smooth: true,
        symbol: "none",
        lineStyle: { color: chartTheme.value.success, width: 2 },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: withAlpha(chartTheme.value.success, 0.33) },
              { offset: 1, color: withAlpha(chartTheme.value.success, 0) },
            ],
          },
        },
      },
    ],
  };
}

const netSendOption = ref(createNetOption(netSent.value));
const netReceivedOption = ref(createNetOption(netReceived.value));

watchEffect(() => {
  const theme = chartTheme.value;
  const sendOpt = netSendOption.value;
  sendOpt.tooltip.backgroundColor = theme.popover;
  sendOpt.tooltip.borderColor = theme.border;
  sendOpt.tooltip.textStyle.color = theme.text;
  sendOpt.series[0].data = netSent.value;
  sendOpt.series[0].lineStyle.color = theme.success;
  sendOpt.series[0].areaStyle.color.colorStops[0].color = withAlpha(theme.success, 0.33);
  sendOpt.series[0].areaStyle.color.colorStops[1].color = withAlpha(theme.success, 0);
});

watchEffect(() => {
  const theme = chartTheme.value;
  const recvOpt = netReceivedOption.value;
  recvOpt.tooltip.backgroundColor = theme.popover;
  recvOpt.tooltip.borderColor = theme.border;
  recvOpt.tooltip.textStyle.color = theme.text;
  recvOpt.series[0].data = netReceived.value;
  recvOpt.series[0].lineStyle.color = theme.success;
  recvOpt.series[0].areaStyle.color.colorStops[0].color = withAlpha(theme.success, 0.33);
  recvOpt.series[0].areaStyle.color.colorStops[1].color = withAlpha(theme.success, 0);
});

function updateNetwork(io: NetIOInfo) {
  const sent = Math.round(io.sent / 1024);
  const received = Math.round(io.received / 1024);

  if (netSent.value.length >= 20) netSent.value.shift();
  if (netReceived.value.length >= 20) netReceived.value.shift();

  netSent.value.push(sent);
  netReceived.value.push(received);
}

const pm2Services = ref<PM2Service[]>([]);
const dockerContainers = ref<DockerContainer[]>([]);

const highUsageServers = computed(() =>
  allServers.value
    .filter((s) => s.cpu >= 80 || s.mem >= 80)
    .sort((a, b) => b.cpu + b.mem - (a.cpu + a.mem))
    .slice(0, 6),
);

async function loadSystemUsage() {
  try {
    const data = await systemStore.fetchSystemUsage();
    updateCpu(data.cpuUsage);
    updateMemory(data.memUsage);
    updateNetwork(data.netIO);
    updateDiskUsageChart(data.diskUsage);
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message });
  } finally {
    usageLoading.value = false;
    statsLoading.value = false;
  }
}

async function loadPm2Services() {
  try {
    const data = await systemStore.fetchPM2Services();
    pm2Services.value = data;
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message });
  } finally {
    servicesLoading.value = false;
  }
}

async function loadDockerContainers() {
  try {
    const data = await systemStore.fetchDockerContainers();
    dockerContainers.value = data;
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message });
  }
}

async function loadServers() {
  try {
    allServers.value = await fetchServers();
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message });
  } finally {
    statsLoading.value = false;
  }
}

onMounted(() => {
  loadServers();
  loadSystemUsage();
  loadPm2Services();
  loadDockerContainers();
  setInterval(loadServers, systemStore.intervalMS * 6);
  setInterval(loadSystemUsage, systemStore.intervalMS);
  setInterval(loadPm2Services, systemStore.intervalMS);
  setInterval(loadDockerContainers, systemStore.intervalMS);
});
</script>
