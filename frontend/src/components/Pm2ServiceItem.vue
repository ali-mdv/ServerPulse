<template>
  <div
    class="card card-body flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
  >
    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="font-medium truncate">{{ props.service.name }}</span>
        <Badge :tone="stateTone">{{ props.service.pm2_env.status }}</Badge>
      </div>
      <div class="text-xs text-muted-foreground mt-1">
        Up Time: {{ uptime }} • Restarts: {{ props.service.pm2_env.restart_time }}
      </div>
      <div class="flex gap-2 flex-wrap mt-3">
        <Button
          variant="success"
          size="sm"
          :disabled="
            loading ||
            actionInProgress === ServiceAction.START ||
            props.service.pm2_env.status === PM2ServiceState.ONLINE
          "
          :loading="actionInProgress === ServiceAction.START"
          @click="handleAction(ServiceAction.START)"
        >
          <template v-if="actionInProgress !== ServiceAction.START" #icon-left>
            <Play class="w-3.5 h-3.5" aria-hidden="true" />
          </template>
          Start
        </Button>
        <Button
          variant="danger"
          size="sm"
          :disabled="
            loading ||
            actionInProgress === ServiceAction.STOP ||
            props.service.pm2_env.status !== PM2ServiceState.ONLINE
          "
          :loading="actionInProgress === ServiceAction.STOP"
          @click="handleAction(ServiceAction.STOP)"
        >
          <template v-if="actionInProgress !== ServiceAction.STOP" #icon-left>
            <Square class="w-3.5 h-3.5" aria-hidden="true" />
          </template>
          Stop
        </Button>
        <Button
          variant="info"
          size="sm"
          :disabled="loading || actionInProgress === ServiceAction.RESTART"
          :loading="actionInProgress === ServiceAction.RESTART"
          @click="handleAction(ServiceAction.RESTART)"
        >
          <template v-if="actionInProgress !== ServiceAction.RESTART" #icon-left>
            <RotateCw class="w-3.5 h-3.5" aria-hidden="true" />
          </template>
          Restart
        </Button>
        <Button
          variant="secondary"
          size="sm"
          :disabled="loading || actionInProgress === ServiceAction.LOGS"
          :loading="actionInProgress === ServiceAction.LOGS"
          @click="handleAction(ServiceAction.LOGS)"
        >
          <template #icon-left>
            <ScrollText class="w-3.5 h-3.5" aria-hidden="true" />
          </template>
          Logs
        </Button>
        <router-link :to="historyLink">
          <Button variant="outline" size="sm">
            <template #icon-left>
              <LineChart class="w-3.5 h-3.5" aria-hidden="true" />
            </template>
            History
          </Button>
        </router-link>
      </div>
    </div>
    <div class="text-right text-sm shrink-0">
      <div>CPU: <span class="font-semibold">{{ cpuPercent }}</span></div>
      <div>Mem: <span class="font-semibold">{{ memUsage }}</span></div>
    </div>
  </div>
  <LogsModal
    :visible="showLogsModal"
    :service-name="props.service.name"
    :content="logsContent"
    :loading="logsLoading"
    @refresh="refreshLogs"
    @update="updateModal"
  />
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { useToast } from "primevue/usetoast";
import { Play, Square, RotateCw, ScrollText, LineChart } from "lucide-vue-next";
import LogsModal from "@/components/LogsModal.vue";
import { useSystemStore } from "@/stores/system";
import { PM2Service, PM2ServiceState, ServiceAction } from "@/types";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const props = defineProps<{
  service: PM2Service;
  /** Scopes the History link so the detail page queries the right server. */
  serverId?: string;
}>();

const historyLink = computed(() => ({
  name: "serviceDetail",
  params: { provider: "pm2", id: props.service.pm_id },
  query: props.serverId ? { serverId: props.serverId } : {},
}));

function parseBytes(value: number, base = 1024, decimals = 2) {
  if (value === 0) return { number: 0, unit: "Bytes" };
  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(value) / Math.log(base));
  const number = Number((value / Math.pow(base, index)).toFixed(decimals));
  return { number, unit: units[index] };
}

function parseUptime(value: number) {
  const now = Date.now();
  const uptimeMs = now - value;
  const seconds = Math.floor(uptimeMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

const memUsage = computed(() => {
  const parsed = parseBytes(props.service.monit.memory);
  return `${parsed.number} ${parsed.unit}`;
});

const cpuPercent = computed(() => `${props.service.monit.cpu} %`);

const uptime = computed(() => parseUptime(props.service.pm2_env.pm_uptime));

const stateTone = computed<"success" | "neutral">(() =>
  props.service.pm2_env.status === PM2ServiceState.ONLINE ? "success" : "neutral",
);

const toast = useToast();
const loading = ref(false);
const actionInProgress = ref<ServiceAction | null>(null);

const showLogsModal = ref(false);
const logsLoading = ref(false);
const logsContent = ref("");

function refreshLogs() {
  handleAction(ServiceAction.LOGS);
}

function updateModal(isOpen: boolean) {
  showLogsModal.value = isOpen;
}

const systemStore = useSystemStore();

async function handleAction(action: ServiceAction) {
  loading.value = true;
  actionInProgress.value = action;

  try {
    if (action === ServiceAction.START) {
      await systemStore.startPM2Service(props.service);
      toast.add({
        severity: "success",
        summary: "Started",
        detail: `Service '${props.service.name}' is now online.`,
        life: 3000,
      });
    } else if (action === ServiceAction.STOP) {
      await systemStore.stopPM2Service(props.service);
      toast.add({
        severity: "success",
        summary: "Stopped",
        detail: `Service '${props.service.name}' has stopped.`,
        life: 3000,
      });
    } else if (action === ServiceAction.RESTART) {
      await systemStore.restartPM2Service(props.service);
      toast.add({
        severity: "success",
        summary: "Restarted",
        detail: `Service '${props.service.name}' has restarted.`,
        life: 3000,
      });
    } else if (action === ServiceAction.LOGS) {
      logsLoading.value = true;
      const logs = await systemStore.getPM2ServiceLogs(props.service);
      logsContent.value = logs;
      toast.add({
        severity: "success",
        summary: "Logs loaded",
        detail: `Showing latest logs for '${props.service.name}'.`,
        life: 2000,
      });
      showLogsModal.value = true;
      logsLoading.value = false;
    }
  } catch (error: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: error.message,
      life: 3000,
    });
  } finally {
    loading.value = false;
    actionInProgress.value = null;
  }
}
</script>
