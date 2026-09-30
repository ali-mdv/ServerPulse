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
      <div class="flex items-center justify-between mb-3">
        <h3 class="font-semibold">CPU Usage (Total)</h3>
        <router-link to="/system/cpu">
          <Button variant="outline" size="sm">
            <LineChart class="w-4 h-4" aria-hidden="true" />
            History
          </Button>
        </router-link>
      </div>
      <div v-if="usageLoading" class="space-y-2">
        <Skeleton height="1rem" width="60%" />
        <Skeleton height="240px" />
      </div>
      <v-chart v-else :option="cpuOption" :autoresize="true" style="height: 260px; width: 100%" />
    </Card>

    <!-- MEMORY & DISK -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-semibold">Memory Usage</h3>
          <router-link to="/system/memory">
            <Button variant="outline" size="sm">
              <LineChart class="w-4 h-4" aria-hidden="true" />
              History
            </Button>
          </router-link>
        </div>
        <div v-if="usageLoading" class="space-y-2">
          <Skeleton height="1rem" width="60%" />
          <Skeleton height="220px" />
        </div>
        <v-chart v-else :option="memOption" :autoresize="true" style="height: 240px; width: 100%" />
      </Card>

      <Card>
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-semibold">Disk Usage</h3>
          <router-link to="/system/disk">
            <Button variant="outline" size="sm">
              <LineChart class="w-4 h-4" aria-hidden="true" />
              History
            </Button>
          </router-link>
        </div>
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
        <div class="flex items-center justify-between mb-2">
          <h3 class="font-semibold">Network Sent (KB/s)</h3>
          <router-link to="/system/network">
            <Button variant="outline" size="sm">
              <LineChart class="w-4 h-4" aria-hidden="true" />
              History
            </Button>
          </router-link>
        </div>
        <Skeleton v-if="usageLoading" height="120px" />
        <v-chart v-else :option="netSendOption" :autoresize="true" class="h-32 w-full" />
      </Card>

      <Card>
        <div class="flex items-center justify-between mb-2">
          <h3 class="font-semibold">Network Received (KB/s)</h3>
          <router-link to="/system/network">
            <Button variant="outline" size="sm">
              <LineChart class="w-4 h-4" aria-hidden="true" />
              History
            </Button>
          </router-link>
        </div>
        <Skeleton v-if="usageLoading" height="120px" />
        <v-chart v-else :option="netReceivedOption" :autoresize="true" class="h-32 w-full" />
      </Card>
    </div>

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
import { Server, Activity, AlertTriangle, CircleSlash, LineChart } from "lucide-vue-next";
import Card from "@/components/Card.vue";
import { useSystemStore } from "@/stores/system";
import { UsageInfo, NetIOInfo } from "@/types";
import Skeleton from "@/components/ui/Skeleton.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import { useChartTheme } from "@/composables";
import { fetchServers } from "@/api/servers";
import { HistoryRange } from "@/api/history";
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
    formatter: (params: { axisValue: string; data: number }[]) => {
      const p = params[0];
      if (!p) return "";
      const value =
        typeof p.data === "number" && !Number.isNaN(p.data)
          ? `${p.data.toFixed(1)}%`
          : "—";
      return `${p.axisValue}<br/>CPU: <strong>${value}</strong>`;
    },
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
const netSentLabels = ref<string[]>(Array(20).fill(""));
const netReceivedLabels = ref<string[]>(Array(20).fill(""));

function createNetOption(
  source: number[],
  labels: string[],
  label: string,
) {
  return {
    animation: false,
    tooltip: {
      trigger: "axis",
      backgroundColor: chartTheme.value.popover,
      borderColor: chartTheme.value.border,
      textStyle: { color: chartTheme.value.text },
      formatter: (params: { axisValue: string; data: number }[]) => {
        const p = params[0];
        if (!p) return "";
        const value =
          typeof p.data === "number" && !Number.isNaN(p.data)
            ? `${p.data} KB/s`
            : "—";
        return `${p.axisValue}<br/>${label}: <strong>${value}</strong>`;
      },
    },
    xAxis: {
      type: "category",
      data: labels,
      axisLine: { lineStyle: { color: chartTheme.value.border } },
      axisTick: { show: false },
      axisLabel: {
        show: true,
        color: chartTheme.value.textMuted,
        fontSize: 10,
        interval: 4,
        formatter: (v: string) => v,
      },
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

const netSendOption = ref(
  createNetOption(netSent.value, netSentLabels.value, "Sent"),
);
const netReceivedOption = ref(
  createNetOption(netReceived.value, netReceivedLabels.value, "Received"),
);

watchEffect(() => {
  const theme = chartTheme.value;
  const sendOpt = netSendOption.value;
  sendOpt.tooltip.backgroundColor = theme.popover;
  sendOpt.tooltip.borderColor = theme.border;
  sendOpt.tooltip.textStyle.color = theme.text;
  sendOpt.xAxis.data = netSentLabels.value;
  sendOpt.xAxis.axisLine.lineStyle.color = theme.border;
  sendOpt.xAxis.axisLabel.color = theme.textMuted;
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
  recvOpt.xAxis.data = netReceivedLabels.value;
  recvOpt.xAxis.axisLine.lineStyle.color = theme.border;
  recvOpt.xAxis.axisLabel.color = theme.textMuted;
  recvOpt.series[0].data = netReceived.value;
  recvOpt.series[0].lineStyle.color = theme.success;
  recvOpt.series[0].areaStyle.color.colorStops[0].color = withAlpha(theme.success, 0.33);
  recvOpt.series[0].areaStyle.color.colorStops[1].color = withAlpha(theme.success, 0);
});

function updateNetwork(io: NetIOInfo) {
  const sent = Math.round(io.sent / 1024);
  const received = Math.round(io.received / 1024);
  const time = new Date().toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  if (netSent.value.length >= 20) {
    netSent.value.shift();
    netSentLabels.value.shift();
  }
  if (netReceived.value.length >= 20) {
    netReceived.value.shift();
    netReceivedLabels.value.shift();
  }

  netSent.value.push(sent);
  netSentLabels.value.push(time);
  netReceived.value.push(received);
  netReceivedLabels.value.push(time);
}

const highUsageServers = computed(() =>
  allServers.value
    .filter((s) => s.cpu >= 80 || s.mem >= 80)
    .sort((a, b) => b.cpu + b.mem - (a.cpu + a.mem))
    .slice(0, 6),
);

async function loadSystemUsage() {
  try {
    const data = await systemStore.fetchSystemState();
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

async function loadServers() {
  try {
    allServers.value = await fetchServers();
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message });
  } finally {
    statsLoading.value = false;
  }
}

async function loadChartHistory() {
  try {
    const [cpuPoints, sentPoints, recvPoints] = await Promise.all([
      systemStore.fetchServiceHistory("system", "cpu", "1h"),
      systemStore.fetchServiceHistory("system", "network_sent", "1h"),
      systemStore.fetchServiceHistory("system", "network_recv", "1h"),
    ]);

    const recentCpu = cpuPoints.slice(-12);
    cpuData.labels = recentCpu.map((p) =>
      new Date(p.ts).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      }),
    );
    cpuData.values = recentCpu.map((p) => Math.round(p.cpu ?? 0));

    const recentSent = sentPoints.slice(-20);
    const recentRecv = recvPoints.slice(-20);
    netSent.value = recentSent.map((p) => Math.round((p.netSent ?? 0) / 1024));
    netSentLabels.value = recentSent.map((p) =>
      new Date(p.ts).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    );
    netReceived.value = recentRecv.map((p) =>
      Math.round((p.netRecv ?? 0) / 1024),
    );
    netReceivedLabels.value = recentRecv.map((p) =>
      new Date(p.ts).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    );
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message });
  }
}

onMounted(async () => {
  loadServers();
  await loadChartHistory();
  loadSystemUsage();
  setInterval(loadServers, systemStore.intervalMS * 6);
  setInterval(loadSystemUsage, systemStore.intervalMS);
});
</script>
