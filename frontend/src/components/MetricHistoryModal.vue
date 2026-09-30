<template>
  <Dialog
    v-model:visible="isOpen"
    :header="title"
    modal
    :style="{ width: '90vw', maxWidth: '900px' }"
    class="p-fluid"
  >
    <div class="space-y-4">
      <div class="flex items-center justify-between gap-3">
        <p class="text-sm text-muted-foreground">{{ subtitle }}</p>
        <Select v-model="range" aria-label="History range" class="w-auto">
          <option value="1h">Last 1 hour</option>
          <option value="6h">Last 6 hours</option>
          <option value="24h">Last 24 hours</option>
          <option value="7d">Last 7 days</option>
        </Select>
      </div>

      <div
        v-if="error"
        role="alert"
        class="card card-body border-critical/40 bg-critical/10 text-sm"
      >
        {{ error }}
      </div>

      <ServiceHistoryChart
        v-else-if="metric === 'network'"
        :series="networkSeries"
        :loading="loading"
      />
      <ServiceHistoryChart
        v-else
        :points="points"
        :loading="loading"
        :field="field"
        :metric-label="metricLabel"
        :color-var="colorVar"
        y-axis-suffix="%"
        show-high-watermark
      />
    </div>
  </Dialog>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import Dialog from "primevue/dialog";
import { useSystemStore } from "@/stores/system";
import {
  ServiceSnapshot,
  SystemMetricKey,
} from "@/types";
import { HistoryRange } from "@/api/history";
import Select from "@/components/ui/Select.vue";
import ServiceHistoryChart from "@/components/ServiceHistoryChart.vue";

interface Props {
  visible?: boolean;
  serverId?: string;
  metric?: SystemMetricKey | null;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  serverId: undefined,
  metric: null,
});

const emit = defineEmits<{
  update: [visible: boolean];
}>();

const systemStore = useSystemStore();

const METRICS: Record<
  SystemMetricKey,
  {
    label: string;
    field: "cpu" | "memory" | "disk";
    colorVar: "critical" | "info" | "warning";
  }
> = {
  cpu: { label: "CPU Usage", field: "cpu", colorVar: "critical" },
  memory: { label: "Memory Usage", field: "memory", colorVar: "info" },
  disk: { label: "Disk Usage", field: "disk", colorVar: "warning" },
  network: { label: "Network I/O", field: "cpu", colorVar: "info" },
};

const isOpen = ref(props.visible);
const range = ref<HistoryRange>("1h");
const loading = ref(false);
const error = ref<string | null>(null);

const points = ref<ServiceSnapshot[]>([]);
const netSent = ref<ServiceSnapshot[]>([]);
const netRecv = ref<ServiceSnapshot[]>([]);

const title = computed(() =>
  props.metric ? `${METRICS[props.metric].label} History` : "History",
);

const subtitle = computed(() => {
  if (!props.metric) return "";
  if (props.metric === "network") {
    return "Sent and received traffic for this server over time.";
  }
  return `${METRICS[props.metric].label} for this server over time.`;
});

const field = computed(() => METRICS[props.metric ?? "cpu"].field);
const metricLabel = computed(() =>
  props.metric ? METRICS[props.metric].label : "",
);
const colorVar = computed(() =>
  props.metric ? METRICS[props.metric].colorVar : "info",
);

const networkSeries = computed(() => [
  {
    field: "netSent" as const,
    label: "Sent",
    colorVar: "info" as const,
    yAxisFormatter: formatBytes,
    points: netSent.value,
  },
  {
    field: "netRecv" as const,
    label: "Received",
    colorVar: "success" as const,
    yAxisFormatter: formatBytes,
    points: netRecv.value,
  },
]);

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

async function load() {
  const metric = props.metric;
  if (!metric || !isOpen.value) return;

  loading.value = true;
  error.value = null;
  points.value = [];
  netSent.value = [];
  netRecv.value = [];

  try {
    if (metric === "network") {
      const [sent, recv] = await Promise.all([
        systemStore.fetchServiceHistory(
          "system",
          "network_sent",
          range.value,
          props.serverId,
        ),
        systemStore.fetchServiceHistory(
          "system",
          "network_recv",
          range.value,
          props.serverId,
        ),
      ]);
      netSent.value = sent;
      netRecv.value = recv;
    } else {
      points.value = await systemStore.fetchServiceHistory(
        "system",
        metric,
        range.value,
        props.serverId,
      );
    }
  } catch (err) {
    error.value = (err as Error).message;
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.visible,
  (visible) => {
    isOpen.value = visible;
  },
);

watch(isOpen, (open) => {
  emit("update", open);
  if (open) void load();
});

watch([range, () => props.metric], () => {
  if (isOpen.value) void load();
});
</script>
