<template>
  <div
    class="card card-body flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
  >
    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="font-medium truncate">{{ props.container.name }}</span>
        <Badge :tone="stateTone">{{ props.container.state }}</Badge>
      </div>
      <div class="text-xs text-muted-foreground mt-1 break-words">
        <template v-if="props.container.upTime">Status: {{ props.container.upTime }}</template>
        <template v-if="props.container.image">
          <template v-if="props.container.upTime"> • </template>Image: {{ props.container.image }}
        </template>
        <template v-if="props.container.port">
          <template v-if="props.container.upTime || props.container.image"> • </template>Port: {{ props.container.port }}
        </template>
        <template v-if="!props.container.upTime && !props.container.image && !props.container.port">
          State snapshot — live details unavailable
        </template>
      </div>
      <div class="flex gap-2 flex-wrap mt-3">
        <Button
          variant="success"
          size="sm"
          :disabled="
            loading ||
            actionInProgress === ServiceAction.START ||
            props.container.state === DockerContainerState.RUNNING
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
            props.container.state !== DockerContainerState.RUNNING
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
    <div
      v-if="props.container.state === DockerContainerState.RUNNING && props.container.usage"
      class="text-right text-sm shrink-0"
    >
      <div>CPU: <span class="font-semibold">{{ cpuPercent }}</span></div>
      <div>Mem: <span class="font-semibold">{{ memPercent }}</span></div>
      <div class="text-xs text-muted-foreground">{{ memUsage }}</div>
    </div>
  </div>
  <LogsModal
    :visible="showLogsModal"
    :service-name="props.container.name"
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
import { DockerContainer, DockerContainerState, ServiceAction } from "@/types";
import { useSystemStore } from "@/stores/system";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const props = defineProps<{
  container: DockerContainer;
  /** Scopes the History link so the detail page queries the right server. */
  serverId?: string;
}>();

const toast = useToast();

const historyLink = computed(() => ({
  name: "serviceDetail",
  params: { provider: "docker", id: props.container.id },
  query: props.serverId ? { serverId: props.serverId } : {},
}));

const loading = ref(false);
const actionInProgress = ref<ServiceAction | null>(null);

function parseBytes(value: number, base = 1024, decimals = 2) {
  if (value === 0) return { number: 0, unit: "Bytes" };
  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(value) / Math.log(base));
  const number = Number((value / Math.pow(base, index)).toFixed(decimals));
  return { number, unit: units[index] };
}

const cpuPercent = computed(
  () => `${Math.round(props.container.usage?.cpuPercent ?? 0)} %`,
);
const memPercent = computed(
  () => `${Math.round(props.container.usage?.memPercent ?? 0)} %`,
);
const memUsage = computed(() => {
  const parsed = parseBytes(props.container.usage?.memUsage ?? 0);
  return `${parsed.number} ${parsed.unit}`;
});

const stateTone = computed<
  "success" | "warning" | "critical" | "neutral"
>(() => {
  switch (props.container.state) {
    case DockerContainerState.RUNNING:
      return "success";
    case DockerContainerState.PAUSED:
      return "warning";
    case DockerContainerState.EXITED:
    case DockerContainerState.DEAD:
      return "critical";
    default:
      return "neutral";
  }
});

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
      await systemStore.startDockerContainer(props.container);
      toast.add({
        severity: "success",
        summary: "Started",
        detail: `Container '${props.container.name}' is now running.`,
        life: 3000,
      });
    } else if (action === ServiceAction.STOP) {
      await systemStore.stopDockerContainer(props.container);
      toast.add({
        severity: "success",
        summary: "Stopped",
        detail: `Container '${props.container.name}' has stopped.`,
        life: 3000,
      });
    } else if (action === ServiceAction.RESTART) {
      await systemStore.restartDockerContainer(props.container);
      toast.add({
        severity: "success",
        summary: "Restarted",
        detail: `Container '${props.container.name}' has restarted.`,
        life: 3000,
      });
    } else if (action === ServiceAction.LOGS) {
      logsLoading.value = true;
      const logs = await systemStore.getContainerLogs(props.container);
      logsContent.value = logs;
      toast.add({
        severity: "success",
        summary: "Logs loaded",
        detail: `Showing latest logs for '${props.container.name}'.`,
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
