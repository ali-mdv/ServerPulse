<template>
  <div class="p-3 bg-card rounded border flex items-center justify-between">
    <div>
      <div class="font-medium">
        {{ props.container.name }}
        <span class="text-sm text-muted">• {{ props.container.state }}</span>
      </div>
      <div class="text-sm text-muted">
        • Up Time: {{ props.container.upTime }} • Image:{{
          props.container.image
        }}
        • Port:{{ props.container.port }}
      </div>
    </div>
    <div class="text-right">
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
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { DockerContainer } from "@/types";

const props = defineProps<{ container: DockerContainer }>();

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
</script>
