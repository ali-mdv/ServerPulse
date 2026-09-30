<template>
  <div class="space-y-6">
    <PageHeader :title="title" :subtitle="subtitle">
      <template #default>
        <Badge tone="neutral" class="uppercase">system</Badge>
        <Button variant="outline" size="sm" @click="goBack">
          <ChevronLeft class="w-4 h-4" aria-hidden="true" />
          Back
        </Button>
      </template>
    </PageHeader>

    <div
      v-if="notFound"
      role="alert"
      class="card card-body border-warning/40 bg-warning/10 text-sm"
    >
      <strong class="font-semibold">Unknown metric.</strong>
      Available metrics are CPU, Memory, Disk, and Network.
    </div>

    <div
      v-else-if="historyError"
      role="alert"
      class="card card-body border-critical/40 bg-critical/10 text-sm"
    >
      {{ historyError }}
    </div>

    <div v-else class="space-y-4">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <div class="text-sm text-muted-foreground">Latest</div>
          <div class="text-2xl font-bold mt-1">{{ latestLabel }}</div>
          <div v-if="latestSub" class="text-xs text-muted-foreground mt-1">
            {{ latestSub }}
          </div>
        </Card>
        <Card>
          <div class="text-sm text-muted-foreground">Average</div>
          <div class="text-2xl font-bold mt-1">{{ averageLabel }}</div>
        </Card>
        <Card>
          <div class="text-sm text-muted-foreground">Peak</div>
          <div class="text-2xl font-bold mt-1">{{ peakLabel }}</div>
          <div v-if="peakSub" class="text-xs text-muted-foreground mt-1">
            {{ peakSub }}
          </div>
        </Card>
        <Card>
          <div class="text-sm text-muted-foreground">Data points</div>
          <div class="text-2xl font-bold mt-1">{{ points.length }}</div>
        </Card>
      </div>

      <Card>
        <div class="flex items-center justify-between gap-3 mb-3">
          <h3 class="font-semibold">History</h3>
          <Select v-model="range" aria-label="History range" class="w-auto">
            <option value="1h">Last 1 hour</option>
            <option value="6h">Last 6 hours</option>
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
          </Select>
        </div>

        <div class="grid gap-4">
          <ServiceHistoryChart
            v-if="metricParam === 'network'"
            :series="networkChartSeries"
            :loading="historyLoading"
          />
          <ServiceHistoryChart
            v-for="series in chartSeries"
            v-else
            :key="series.field"
            :points="series.points"
            :loading="historyLoading"
            :field="series.field"
            :metric-label="series.label"
            :color-var="series.colorVar"
            :y-axis-suffix="series.yAxisSuffix"
            :y-axis-formatter="series.yAxisFormatter"
            :show-high-watermark="series.showHighWatermark"
          />
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useToast } from "primevue/usetoast";
import { ChevronLeft } from "lucide-vue-next";
import { useSystemStore } from "@/stores/system";
import {
  ServiceSnapshot,
  SystemMetricKey,
  SystemUsage,
} from "@/types/system";
import { HistoryRange } from "@/api/history";
import Card from "@/components/Card.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Select from "@/components/ui/Select.vue";
import ServiceHistoryChart from "@/components/ServiceHistoryChart.vue";

const route = useRoute();
const router = useRouter();
const toast = useToast();
const systemStore = useSystemStore();

const metricParam = computed(() =>
  String(route.params.metric ?? "").toLowerCase(),
);

const METRIC_IDS: Record<SystemMetricKey, string> = {
  cpu: "cpu",
  memory: "memory",
  disk: "disk",
  network: "network",
};

const range = ref<HistoryRange>("1h");

const networkSent = ref<ServiceSnapshot[]>([]);
const networkRecv = ref<ServiceSnapshot[]>([]);
const single = ref<ServiceSnapshot[]>([]);
const systemUsage = ref<SystemUsage | null>(null);

const historyLoading = ref(false);
const historyError = ref<string | null>(null);

const notFound = computed(
  () => !(metricParam.value in METRIC_IDS),
);

const title = computed(() => {
  switch (metricParam.value) {
    case "cpu":
      return "CPU Usage";
    case "memory":
      return "Memory Usage";
    case "disk":
      return "Disk Usage";
    case "network":
      return "Network Traffic";
    default:
      return "System metric";
  }
});

const subtitle = computed(() =>
  notFound.value
    ? "Unknown metric"
    : `system-wide ${title.value.toLowerCase()} over time`,
);

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B/s";
  const units = ["B/s", "KB/s", "MB/s", "GB/s"];
  const i = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1,
  );
  const v = bytes / Math.pow(1024, i);
  return `${v.toFixed(v >= 100 ? 0 : 1)} ${units[i]}`;
}

function formatPercent(v: number): string {
  return Number.isFinite(v) ? `${v.toFixed(1)}%` : "—";
}

interface ChartSeries {
  field: "cpu" | "memory" | "disk" | "netSent" | "netRecv";
  label: string;
  colorVar: "critical" | "info" | "warning" | "success";
  yAxisSuffix?: string;
  yAxisFormatter?: (v: number) => string;
  showHighWatermark?: boolean;
  points: ServiceSnapshot[];
}

const chartSeries = computed<ChartSeries[]>(() => {
  switch (metricParam.value) {
    case "cpu":
      return [
        {
          field: "cpu",
          label: "CPU",
          colorVar: "critical",
          yAxisSuffix: "%",
          yAxisFormatter: (v) => `${v.toFixed(0)}%`,
          showHighWatermark: true,
          points: single.value,
        },
      ];
    case "memory":
      return [
        {
          field: "memory",
          label: "Memory",
          colorVar: "info",
          yAxisSuffix: "%",
          yAxisFormatter: (v) => `${v.toFixed(0)}%`,
          showHighWatermark: true,
          points: single.value,
        },
      ];
    case "disk":
      return [
        {
          field: "disk",
          label: "Disk",
          colorVar: "warning",
          yAxisSuffix: "%",
          yAxisFormatter: (v) => `${v.toFixed(0)}%`,
          showHighWatermark: true,
          points: single.value,
        },
      ];
    default:
      return [];
  }
});

const networkChartSeries = computed<ChartSeries[]>(() => [
  {
    field: "netSent",
    label: "Sent",
    colorVar: "info",
    yAxisFormatter: formatBytes,
    points: networkSent.value,
  },
  {
    field: "netRecv",
    label: "Received",
    colorVar: "success",
    yAxisFormatter: formatBytes,
    points: networkRecv.value,
  },
]);

const points = computed(() => {
  const m = metricParam.value;
  if (m === "network") return [...networkSent.value, ...networkRecv.value];
  return single.value;
});

const currentValues = computed<number[]>(() => {
  switch (metricParam.value) {
    case "cpu":
    case "memory":
    case "disk":
      return single.value
        .map((p) => p[metricParam.value === "cpu" ? "cpu" : metricParam.value === "memory" ? "memory" : "disk"])
        .filter((v): v is number => typeof v === "number" && !Number.isNaN(v));
    case "network":
      return [
        ...networkSent.value
          .map((p) => p.netSent)
          .filter((v): v is number => typeof v === "number"),
        ...networkRecv.value
          .map((p) => p.netRecv)
          .filter((v): v is number => typeof v === "number"),
      ];
    default:
      return [];
  }
});

const latestLabel = computed(() => {
  const m = metricParam.value;
  if (m === "cpu") {
    const v = systemUsage.value?.cpuUsage;
    return typeof v === "number" ? formatPercent(v) : "—";
  }
  if (m === "memory") {
    const v = systemUsage.value?.memUsage.percent;
    return typeof v === "number" ? formatPercent(v) : "—";
  }
  if (m === "disk") {
    const v = systemUsage.value?.diskUsage.percent;
    return typeof v === "number" ? formatPercent(v) : "—";
  }
  if (m === "network") {
    const sent = systemUsage.value?.netIO.sent ?? 0;
    return `${formatBytes(sent)} ↑`;
  }
  return "—";
});

const latestSub = computed(() => {
  if (metricParam.value !== "network") return "";
  const recv = systemUsage.value?.netIO.received ?? 0;
  return `${formatBytes(recv)} ↓`;
});

const averageLabel = computed(() => {
  const m = metricParam.value;
  if (!currentValues.value.length) return "—";
  const avg =
    currentValues.value.reduce((a, b) => a + b, 0) /
    currentValues.value.length;
  if (m === "network") return formatBytes(avg);
  return formatPercent(avg);
});

const peakLabel = computed(() => {
  const m = metricParam.value;
  if (!currentValues.value.length) return "—";
  const peak = Math.max(...currentValues.value);
  if (m === "network") return formatBytes(peak);
  return formatPercent(peak);
});

const peakSub = computed(() => {
  if (metricParam.value !== "network") return "";
  const sentPeak = Math.max(
    0,
    ...networkSent.value
      .map((p) => p.netSent ?? 0)
      .filter((v) => typeof v === "number"),
  );
  const recvPeak = Math.max(
    0,
    ...networkRecv.value
      .map((p) => p.netRecv ?? 0)
      .filter((v) => typeof v === "number"),
  );
  return `↑ ${formatBytes(sentPeak)} • ↓ ${formatBytes(recvPeak)}`;
});

function goBack() {
  if (window.history.length > 1) router.back();
  else router.push("/dashboard");
}

async function loadCurrent() {
  try {
    systemUsage.value = await systemStore.fetchSystemState();
  } catch {
    // State is optional; history is the main content.
  }
}

async function loadHistory() {
  if (notFound.value) return;
  historyLoading.value = true;
  historyError.value = null;
  networkSent.value = [];
  networkRecv.value = [];
  single.value = [];
  try {
    const m = metricParam.value;
    if (m === "network") {
      const [sent, recv] = await Promise.all([
        systemStore.fetchServiceHistory("system", "network_sent", range.value),
        systemStore.fetchServiceHistory("system", "network_recv", range.value),
      ]);
      networkSent.value = sent;
      networkRecv.value = recv;
    } else {
      single.value = await systemStore.fetchServiceHistory(
        "system",
        m,
        range.value,
      );
    }
  } catch (err) {
    historyError.value = (err as Error).message;
    toast.add({
      severity: "error",
      summary: "Error",
      detail: historyError.value,
      life: 3000,
    });
  } finally {
    historyLoading.value = false;
  }
}

onMounted(() => {
  loadCurrent();
  loadHistory();
});

watch(metricParam, () => {
  loadCurrent();
  loadHistory();
});
watch(range, loadHistory);
</script>