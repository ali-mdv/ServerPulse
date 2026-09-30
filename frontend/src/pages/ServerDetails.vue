<template>
  <div class="space-y-6">
    <PageHeader
      :title="server?.name || 'Server details'"
      :subtitle="subtitle"
    >
      <template #default>
        <Badge v-if="server?.status" :tone="statusTone">
          {{ server.status }}
        </Badge>
        <Button
          v-if="server"
          variant="outline"
          size="sm"
          :loading="generatingKey"
          @click="generateKey"
        >
          <Key class="w-4 h-4" aria-hidden="true" />
          API key
        </Button>
        <router-link :to="`/servers/${id}/edit`">
          <Button variant="outline" size="sm">
            <Pencil class="w-4 h-4" aria-hidden="true" />
            Edit
          </Button>
        </router-link>
        <router-link to="/servers">
          <Button variant="outline" size="sm">
            <ChevronLeft class="w-4 h-4" aria-hidden="true" />
            Back
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
      <div
        v-if="apiKey"
        class="card card-body border-warning/40 bg-warning/10 text-sm"
        role="status"
      >
        <strong class="font-semibold">API key generated.</strong>
        Copy it now — it will not be shown again.
        <code class="block mt-1 break-all">{{ apiKey }}</code>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <div class="flex items-center justify-between gap-2">
            <div class="text-sm text-muted-foreground">CPU Usage</div>
            <Button
              variant="ghost"
              size="icon"
              aria-label="View CPU usage history"
              @click="openHistory('cpu')"
            >
              <History class="w-4 h-4" aria-hidden="true" />
            </Button>
          </div>
          <div class="text-3xl font-bold mt-1">{{ cpu }}%</div>
          <div class="w-full bg-muted rounded-full h-2 mt-3">
            <div
              class="bg-info h-2 rounded-full transition-all"
              :style="{ width: `${Math.min(cpu, 100)}%` }"
            />
          </div>
        </Card>

        <Card>
          <div class="flex items-center justify-between gap-2">
            <div class="text-sm text-muted-foreground">Memory Usage</div>
            <Button
              variant="ghost"
              size="icon"
              aria-label="View memory usage history"
              @click="openHistory('memory')"
            >
              <History class="w-4 h-4" aria-hidden="true" />
            </Button>
          </div>
          <div class="text-3xl font-bold mt-1">{{ mem }}%</div>
          <div class="text-xs text-muted-foreground mt-1">
            {{ memUsed }} / {{ memTotal }}
          </div>
          <div class="w-full bg-muted rounded-full h-2 mt-2">
            <div
              class="bg-warning h-2 rounded-full transition-all"
              :style="{ width: `${Math.min(mem, 100)}%` }"
            />
          </div>
        </Card>

        <Card>
          <div class="flex items-center justify-between gap-2">
            <div class="text-sm text-muted-foreground">Disk Usage</div>
            <Button
              variant="ghost"
              size="icon"
              aria-label="View disk usage history"
              @click="openHistory('disk')"
            >
              <History class="w-4 h-4" aria-hidden="true" />
            </Button>
          </div>
          <div class="text-3xl font-bold mt-1">{{ disk }}%</div>
          <div class="text-xs text-muted-foreground mt-1">
            {{ diskUsed }} / {{ diskTotal }}
          </div>
          <div class="w-full bg-muted rounded-full h-2 mt-2">
            <div
              class="bg-success h-2 rounded-full transition-all"
              :style="{ width: `${Math.min(disk, 100)}%` }"
            />
          </div>
        </Card>

        <Card>
          <div class="flex items-center justify-between gap-2">
            <div class="text-sm text-muted-foreground">Network I/O</div>
            <Button
              variant="ghost"
              size="icon"
              aria-label="View network I/O history"
              @click="openHistory('network')"
            >
              <History class="w-4 h-4" aria-hidden="true" />
            </Button>
          </div>
          <div class="text-lg font-bold mt-1">
            ↑ {{ netSent }} KB/s
          </div>
          <div class="text-lg font-bold">
            ↓ {{ netReceived }} KB/s
          </div>
          <div class="text-xs text-muted-foreground mt-1">per second</div>
        </Card>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="lg:col-span-2">
          <ServiceManagerPanel
            :services="pm2Services"
            :pm2-available="pm2Available"
            :containers="dockerContainers"
            :docker-available="dockerAvailable"
            :loading="servicesLoading"
            :server-id="id"
            searchable
          />
        </div>

        <div class="space-y-4">
          <Card>
            <h3 class="font-semibold mb-2">Server Info</h3>
            <dl class="text-sm space-y-2">
              <div class="flex justify-between">
                <dt class="text-muted-foreground">Host</dt>
                <dd class="font-medium">{{ server.host || "—" }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">Port</dt>
                <dd class="font-medium">{{ server.port || "—" }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">Last seen</dt>
                <dd class="font-medium">{{ lastSeen }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">Created</dt>
                <dd class="font-medium">{{ formatDate(server.createdAt) }}</dd>
              </div>
            </dl>
            <p
              v-if="server.description"
              class="text-sm text-muted-foreground mt-3 border-t pt-3"
            >
              {{ server.description }}
            </p>
          </Card>

          <Card>
            <h3 class="font-semibold mb-2">Provider Summary</h3>
            <dl class="text-sm space-y-1">
              <div class="flex justify-between">
                <dt class="text-muted-foreground">PM2</dt>
                <dd>
                  <Badge :tone="pm2Available ? 'success' : 'neutral'">
                    {{ pm2Available ? "Available" : "Unavailable" }}
                  </Badge>
                </dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">Docker</dt>
                <dd>
                  <Badge :tone="dockerAvailable ? 'success' : 'neutral'">
                    {{ dockerAvailable ? "Available" : "Unavailable" }}
                  </Badge>
                </dd>
              </div>
            </dl>
          </Card>
        </div>
      </div>
    </div>

    <MetricHistoryModal
      :visible="historyVisible"
      :metric="historyMetric"
      :server-id="id"
      @update="historyVisible = $event"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, computed } from "vue";
import { useRoute } from "vue-router";
import { useToast } from "primevue/usetoast";
import { ChevronLeft, Pencil, Key, History } from "lucide-vue-next";
import { fetchServer, generateServerApiKey, type Server } from "@/api/servers";
import { fetchServicesState, fetchSystemState } from "@/api/state";
import ServiceManagerPanel from "@/components/ServiceManagerPanel.vue";
import MetricHistoryModal from "@/components/MetricHistoryModal.vue";
import Card from "@/components/Card.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import {
  snapshotsToPM2Services,
  snapshotsToDockerContainers,
} from "@/lib/service-snapshot";
import { PM2Service, DockerContainer, ProviderState, SystemMetricKey } from "@/types";

const toast = useToast();
const route = useRoute();
const id = route.params.id as string;

const server = ref<Server | null>(null);
const loading = ref(true);
const servicesLoading = ref(true);
const apiKey = ref("");
const generatingKey = ref(false);

const historyVisible = ref(false);
const historyMetric = ref<SystemMetricKey | null>(null);

function openHistory(metric: SystemMetricKey) {
  historyMetric.value = metric;
  historyVisible.value = true;
}

const pm2State = ref<ProviderState>({ provider: "pm2", available: false, services: [], updatedAt: "" });
const dockerState = ref<ProviderState>({ provider: "docker", available: false, services: [], updatedAt: "" });

let timer: ReturnType<typeof setInterval> | null = null;

const pm2Services = computed<PM2Service[]>(() =>
  snapshotsToPM2Services(pm2State.value.services ?? []),
);
const pm2Available = computed(() => pm2State.value.available);

const dockerContainers = computed<DockerContainer[]>(() =>
  snapshotsToDockerContainers(dockerState.value.services ?? []),
);
const dockerAvailable = computed(() => dockerState.value.available);

const statusTone = computed<"success" | "critical" | "neutral">(() => {
  const s = server.value?.status;
  if (s === "online") return "success";
  if (s === "down") return "critical";
  return "neutral";
});

const subtitle = computed(() => {
  if (!server.value) return "Loading…";
  const parts: string[] = [];
  if (server.value.host) parts.push(server.value.host);
  if (server.value.port) parts.push(String(server.value.port));
  return parts.length ? parts.join(":") : "Server details";
});

const usage = computed(() => server.value?.usage?.usage);

const cpu = computed(() => Math.round(usage.value?.cpuUsage ?? 0));
const mem = computed(() => Math.round(usage.value?.memUsage.percent ?? 0));
const disk = computed(() => Math.round(usage.value?.diskUsage.percent ?? 0));
const netSent = computed(() => Math.round((usage.value?.netIO.sent ?? 0) / 1024));
const netReceived = computed(() => Math.round((usage.value?.netIO.received ?? 0) / 1024));

function parseBytes(value: number) {
  if (value === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(value) / Math.log(1024));
  const number = Number((value / Math.pow(1024, index)).toFixed(2));
  return `${number} ${units[index]}`;
}

const memUsed = computed(() => parseBytes(usage.value?.memUsage.used ?? 0));
const memTotal = computed(() => parseBytes(usage.value?.memUsage.total ?? 0));
const diskUsed = computed(() => parseBytes(usage.value?.diskUsage.used ?? 0));
const diskTotal = computed(() => parseBytes(usage.value?.diskUsage.total ?? 0));

const lastSeen = computed(() =>
  server.value?.lastSeen ? formatDate(server.value.lastSeen) : "Never",
);

function formatDate(value: string) {
  return new Date(value).toLocaleString();
}

async function loadServer() {
  try {
    const data = await fetchServer(id);
    if (data) server.value = data;
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message, life: 3000 });
  }
}

async function loadState() {
  try {
    const [services, system] = await Promise.all([
      fetchServicesState(id),
      fetchSystemState(id),
    ]);
    pm2State.value = services.pm2;
    dockerState.value = services.docker;
    if (server.value && system.systemUsage) {
      server.value.usage = {
        serverId: id,
        updatedAt: system.updatedAt ?? new Date().toISOString(),
        usage: system.systemUsage,
      };
    }
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message, life: 3000 });
  } finally {
    servicesLoading.value = false;
  }
}

async function generateKey() {
  generatingKey.value = true;
  try {
    apiKey.value = await generateServerApiKey(id);
  } catch (err: any) {
    toast.add({ severity: "error", summary: "Error", detail: err.message, life: 3000 });
  } finally {
    generatingKey.value = false;
  }
}

onMounted(async () => {
  await loadServer();
  loading.value = false;
  await loadState();
  timer = setInterval(async () => {
    await loadServer();
    await loadState();
  }, 10000);
});

onBeforeUnmount(() => {
  if (timer) clearInterval(timer);
});
</script>
