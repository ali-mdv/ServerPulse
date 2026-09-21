<template>
  <div class="space-y-6">
    <PageHeader
      :title="title"
      :subtitle="subtitle"
    >
      <template #default>
        <Badge v-if="statusLabel" :tone="statusTone">{{ statusLabel }}</Badge>
        <Badge tone="neutral" class="uppercase">{{ provider }}</Badge>
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
      <strong class="font-semibold">Service is not running right now.</strong>
      It may have been stopped, removed, or its provider may be offline. Any
      recorded history is still shown below.
    </div>

    <div
      v-if="loading"
      class="grid gap-3"
      role="status"
      aria-live="polite"
      aria-label="Loading service"
    >
      <Skeleton height="11rem" />
      <Skeleton height="11rem" />
    </div>

    <div v-else class="space-y-4">
      <div v-if="!notFound" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <div class="text-sm text-muted-foreground">Status</div>
          <div class="text-2xl font-bold mt-1">{{ statusLabel || "—" }}</div>
        </Card>
        <Card>
          <div class="text-sm text-muted-foreground">CPU</div>
          <div class="text-2xl font-bold mt-1">
            {{ currentCpu === null ? "—" : `${currentCpu.toFixed(1)}%` }}
          </div>
        </Card>
        <Card>
          <div class="text-sm text-muted-foreground">
            {{ provider === "docker" ? "Memory" : "Memory" }}
          </div>
          <div class="text-2xl font-bold mt-1">{{ currentMemoryLabel }}</div>
        </Card>
        <Card>
          <div class="text-sm text-muted-foreground">Uptime / Restarts</div>
          <div class="text-base font-semibold mt-1">
            {{ uptimeLabel || "—" }}
          </div>
          <div v-if="restartLabel" class="text-xs text-muted-foreground">
            {{ restartLabel }}
          </div>
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

        <div
          v-if="historyError"
          role="alert"
          class="card card-body border-critical/40 bg-critical/10 text-sm"
        >
          {{ historyError }}
        </div>

        <div v-else class="grid gap-6">
          <ServiceHistoryChart
            title="CPU"
            :points="history"
            :loading="historyLoading"
            field="cpu"
            metric-label="CPU"
            color-var="critical"
            y-axis-suffix="%"
            show-high-watermark
          />
          <ServiceHistoryChart
            :title="memoryMetricLabel"
            :points="history"
            :loading="historyLoading"
            field="memory"
            :metric-label="memoryMetricLabel"
            :y-axis-suffix="memorySuffix"
            show-high-watermark
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
  PM2Service,
  PM2ServiceState,
  DockerContainer,
  DockerContainerState,
  HistoryProvider,
  ServiceSnapshot,
} from "@/types/system";
import { HistoryRange } from "@/api/history";
import {
  snapshotToPM2Service,
  snapshotToDockerContainer,
} from "@/lib/service-snapshot";
import Card from "@/components/Card.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Select from "@/components/ui/Select.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import ServiceHistoryChart from "@/components/ServiceHistoryChart.vue";

const route = useRoute();
const router = useRouter();
const toast = useToast();
const systemStore = useSystemStore();

const provider = computed(() => route.params.provider as HistoryProvider);
const serviceId = computed(() => route.params.id as string);
// Absent for the local server; the backend then resolves the local row itself.
const serverId = computed(() =>
  typeof route.query.serverId === "string" ? route.query.serverId : undefined,
);

const range = ref<HistoryRange>("1h");

const pm2Service = ref<PM2Service | null>(null);
const dockerContainer = ref<DockerContainer | null>(null);
const loading = ref(true);
const notFound = ref(false);

const history = ref<ServiceSnapshot[]>([]);
const historyLoading = ref(false);
const historyError = ref<string | null>(null);

const statusLabel = computed(() => {
  if (provider.value === "pm2") return pm2Service.value?.pm2_env.status ?? "";
  return dockerContainer.value?.state ?? "";
});

const statusTone = computed<"success" | "warning" | "critical" | "neutral">(
  () => {
    const s = statusLabel.value;
    if (s === PM2ServiceState.ONLINE || s === DockerContainerState.RUNNING) {
      return "success";
    }
    if (s === DockerContainerState.PAUSED) return "warning";
    if (
      s === DockerContainerState.EXITED ||
      s === DockerContainerState.DEAD ||
      s === PM2ServiceState.STOPPED
    ) {
      return "critical";
    }
    return "neutral";
  },
);

const currentCpu = computed<number | null>(() => {
  if (provider.value === "pm2") return pm2Service.value?.monit.cpu ?? null;
  return dockerContainer.value?.usage?.cpuPercent ?? null;
});

const memoryMetricLabel = computed(() =>
  provider.value === "docker" ? "Memory %" : "Memory (MB)",
);

const memorySuffix = computed(() => (provider.value === "docker" ? "%" : " MB"));

const currentMemoryLabel = computed(() => {
  if (provider.value === "docker") {
    const mem = dockerContainer.value?.usage;
    if (!mem) return "—";
    const mb = mem.memUsage / (1024 * 1024);
    return `${mem.memPercent.toFixed(0)}% (${mb.toFixed(0)} MB)`;
  }
  const bytes = pm2Service.value?.monit.memory ?? 0;
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(1)} MB`;
});

const uptimeLabel = computed(() => {
  if (provider.value === "pm2") {
    const env = pm2Service.value?.pm2_env;
    if (!env?.pm_uptime) return "";
    return formatUptime(env.pm_uptime);
  }
  return dockerContainer.value?.upTime ?? "";
});

const restartLabel = computed(() => {
  if (provider.value === "pm2") {
    const r = pm2Service.value?.pm2_env.restart_time ?? 0;
    return r === 0 ? "no restarts" : `${r} restart${r === 1 ? "" : "s"}`;
  }
  return "";
});

const title = computed(() => {
  const name =
    pm2Service.value?.name ??
    dockerContainer.value?.name ??
    // A stopped service has no live state, but its snapshots still carry a name.
    history.value.at(-1)?.meta.name ??
    (notFound.value ? "" : serviceId.value);
  return name || "Service details";
});

const subtitle = computed(() => {
  const id = serviceId.value;
  if (!id) return "No service selected";
  return `id: ${id}`;
});

function formatUptime(ts: number) {
  const diff = Date.now() - ts;
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

function goBack() {
  if (window.history.length > 1) router.back();
  else router.push("/dashboard");
}

async function loadCurrent() {
  loading.value = true;
  notFound.value = false;
  pm2Service.value = null;
  dockerContainer.value = null;

  try {
    const { pm2, docker } = await systemStore.fetchServicesState(
      serverId.value,
    );

    if (provider.value === "pm2") {
      const found = pm2.services.find(
        (s) => s.meta.provider === "pm2" && s.meta.serviceId === serviceId.value,
      );
      if (!found) {
        notFound.value = true;
        return;
      }
      pm2Service.value = snapshotToPM2Service(found);
    } else if (provider.value === "docker") {
      const found = docker.services.find(
        (s) =>
          s.meta.provider === "docker" && s.meta.serviceId === serviceId.value,
      );
      if (!found) {
        notFound.value = true;
        return;
      }
      dockerContainer.value = snapshotToDockerContainer(found);
    } else {
      notFound.value = true;
    }
  } catch (err) {
    notFound.value = true;
    toast.add({
      severity: "error",
      summary: "Error",
      detail: (err as Error).message,
      life: 3000,
    });
  } finally {
    loading.value = false;
  }
}

async function loadHistory() {
  if (!provider.value || !serviceId.value) return;
  historyLoading.value = true;
  historyError.value = null;
  try {
    history.value = await systemStore.fetchServiceHistory(
      provider.value,
      serviceId.value,
      range.value,
      serverId.value,
    );
  } catch (err) {
    history.value = [];
    historyError.value = (err as Error).message;
  } finally {
    historyLoading.value = false;
  }
}

onMounted(() => {
  loadCurrent();
  loadHistory();
});

watch(
  () => [provider.value, serviceId.value, serverId.value],
  () => {
    loadCurrent();
    loadHistory();
  },
);

watch(range, () => {
  loadHistory();
});
</script>