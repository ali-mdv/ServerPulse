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
</template>

<script lang="ts" setup>
import { computed } from "vue";
import type { PM2Service } from "@/types";

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
</script>
