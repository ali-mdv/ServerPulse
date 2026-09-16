<template>
  <div class="p-3 bg-card rounded border flex items-center justify-between">
    <div>
      <div class="font-medium">
        {{ props.container.name }}
        <span class="text-sm text-muted">• {{ props.container.state }}</span>
      </div>
      <div class="text-sm text-muted">
        • Status: {{ props.container.upTime }} • Image:{{
          props.container.image
        }}
        • Port:{{ props.container.port }}
      </div>
      <div class="flex gap-2 flex-wrap mt-2">
        <button
          @click="handleAction(ServiceAction.START)"
          :disabled="
            loading || props.container.state === DockerContainerState.RUNNING
          "
          class="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {{ ServiceAction.START }}
        </button>
        <button
          @click="handleAction(ServiceAction.STOP)"
          :disabled="
            loading || props.container.state !== DockerContainerState.RUNNING
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
    <div
      class="text-right"
      v-if="props.container.state === DockerContainerState.RUNNING"
    >
      <div class="text-sm">
        CPU: <span class="font-semibold">{{ cpuPercent }}</span>
      </div>
      <div class="text-sm">
        Mem: <span class="font-semibold">{{ memPercent }} </span>
      </div>
      <div class="text-sm">
        Mem: <span class="font-semibold">{{ memUsage }} </span>
      </div>
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
import LogsModal from "@/components/LogsModal.vue";
import { DockerContainer, DockerContainerState, ServiceAction } from "@/types";
import { useSystemStore } from "@/stores/system";

const props = defineProps<{ container: DockerContainer }>();

const toast = useToast();

const loading = ref(false);
const actionInProgress = ref<ServiceAction | null>(null);

function parseBytes(value: number, base = 1024, decimals = 2) {
  if (value === 0) return { number: 0, unit: "Bytes" };

  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const index = Math.floor(Math.log(value) / Math.log(base));
  const number = Number((value / Math.pow(base, index)).toFixed(decimals));

  return { number, unit: units[index] };
}

const cpuPercent = computed(() => {
  return `${Math.round(props.container.usage.cpuPercent)} %`;
});

const memPercent = computed(() => {
  return `${Math.round(props.container.usage.memPercent)} %`;
});

const memUsage = computed(() => {
  const parsed = parseBytes(props.container.usage.memUsage);
  return `${parsed.number} ${parsed.unit}`;
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
        severity: "info",
        summary: "Info",
        detail: `Successfully started container '${props.container.name}'.`,
        life: 3000,
      });
    } else if (action === ServiceAction.STOP) {
      await systemStore.stopDockerContainer(props.container);
      toast.add({
        severity: "info",
        summary: "Info",
        detail: `Successfully stopped container '${props.container.name}'.`,
        life: 3000,
      });
    } else if (action === ServiceAction.RESTART) {
      await systemStore.restartDockerContainer(props.container);
      toast.add({
        severity: "info",
        summary: "Info",
        detail: `Successfully restarted container '${props.container.name}'.`,
        life: 3000,
      });
    } else if (action === ServiceAction.LOGS) {
      logsLoading.value = true;
      const logs = await systemStore.getContainerLogs(props.container);
      logsContent.value = logs;
      toast.add({
        severity: "info",
        summary: "Info",
        detail: `Successfully retrieved logs for container '${props.container.name}'.`,
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
