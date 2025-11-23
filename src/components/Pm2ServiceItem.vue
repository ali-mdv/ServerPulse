<template>
  <div class="p-3 bg-card rounded border flex items-center justify-between">
    <div>
      <div class="font-medium">
        {{ props.service.name }}
        <span class="text-sm text-muted">
          • {{ props.service.pm2_env.status }}
        </span>
      </div>
      <div class="text-sm text-muted">
        • Up Time: {{ uptime }} • Restart:
        {{ props.service.pm2_env.restart_time }}
      </div>
      <div class="flex gap-2 flex-wrap mt-2">
        <button
          @click="handleAction(ServiceAction.START)"
          :disabled="
            loading || props.service.pm2_env.status === PM2ServiceState.ONLINE
          "
          class="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {{ ServiceAction.START }}
        </button>
        <button
          @click="handleAction(ServiceAction.STOP)"
          :disabled="
            loading || props.service.pm2_env.status !== PM2ServiceState.ONLINE
          "
          class="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {{ ServiceAction.STOP }}
        </button>
        <button
          @click="handleAction(ServiceAction.RESTART)"
          :disabled="loading"
          class="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {{ ServiceAction.RESTART }}
        </button>
        <button
          @click="handleAction(ServiceAction.LOGS)"
          :disabled="loading"
          class="px-3 py-1 text-sm bg-gray-600 text-white rounded hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {{ ServiceAction.LOGS }}
        </button>
      </div>
    </div>
    <div class="text-right">
      <div class="text-sm">
        CPU: <span class="font-semibold">{{ cpuPercent }}</span>
      </div>
      <div class="text-sm">
        Mem:
        <span class="font-semibold">{{ memUsage }}</span>
      </div>
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
import LogsModal from "@/components/LogsModal.vue";
import { useSystemStore } from "@/stores/system";
import { PM2Service, PM2ServiceState, ServiceAction } from "@/types";

const props = defineProps<{ service: PM2Service }>();

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

const cpuPercent = computed(() => {
  const cpu = props.service.monit.cpu;
  return `${cpu} %`;
});

const uptime = computed(() => {
  const parsed = parseUptime(props.service.pm2_env.pm_uptime);
  return parsed;
});

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
        severity: "info",
        summary: "Info",
        detail: `Successfully started service '${props.service.name}'.`,
        life: 3000,
      });
    } else if (action === ServiceAction.STOP) {
      await systemStore.stopPM2Service(props.service);
      toast.add({
        severity: "info",
        summary: "Info",
        detail: `Successfully stopped service '${props.service.name}'.`,
        life: 3000,
      });
    } else if (action === ServiceAction.RESTART) {
      await systemStore.restartPM2Service(props.service);
      toast.add({
        severity: "info",
        summary: "Info",
        detail: `Successfully restarted service '${props.service.name}'.`,
        life: 3000,
      });
    } else if (action === ServiceAction.LOGS) {
      logsLoading.value = true;
      const logs = await systemStore.getPM2ServiceLogs(props.service);
      logsContent.value = logs;
      toast.add({
        severity: "info",
        summary: "Info",
        detail: `Successfully retrieved logs for service '${props.service.name}'.`,
        life: 3000,
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
