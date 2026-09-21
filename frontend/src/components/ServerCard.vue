<template>
  <div
    class="card card-body flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
  >
    <div class="min-w-0">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="font-semibold text-base truncate">{{ server.name }}</span>
        <Badge :tone="statusTone">{{ server.status }}</Badge>
      </div>
      <div class="text-sm text-muted-foreground mt-1">
        <template v-if="server.host">{{ server.host }}</template>
        <template v-if="server.host && server.port">:</template>
        <template v-if="server.port">{{ server.port }}</template>
        <template v-if="!server.host && !server.port">No host reported yet</template>
        <template v-if="server.usage">
          • CPU: {{ Math.round(server.usage.usage.cpuUsage) }}% • Mem: {{ Math.round(server.usage.usage.memUsage.percent) }}%
        </template>
      </div>
      <div v-if="server.description" class="text-xs text-muted-foreground mt-1">
        {{ server.description }}
      </div>
    </div>
    <div class="flex items-center gap-2 self-start sm:self-auto">
      <router-link :to="`/servers/${server.id}/edit`">
        <Button variant="outline" size="sm">
          <Pencil class="w-4 h-4" aria-hidden="true" />
          Edit
        </Button>
      </router-link>
      <Button
        v-if="server.name !== 'local'"
        variant="outline"
        size="sm"
        @click="emit('delete', server)"
      >
        <Trash2 class="w-4 h-4" aria-hidden="true" />
        Delete
      </Button>
      <router-link :to="`/server/${server.id}`">
        <Button variant="primary" size="sm">
          View
          <ChevronRight class="w-4 h-4" aria-hidden="true" />
        </Button>
      </router-link>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { ChevronRight, Pencil, Trash2 } from "lucide-vue-next";
import type { Server } from "@/api/servers";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const props = defineProps<{ server: Server }>();
const emit = defineEmits<{ delete: [server: Server] }>();

const statusTone = computed<"success" | "critical" | "neutral">(() => {
  if (props.server.status === "online") return "success";
  if (props.server.status === "down") return "critical";
  return "neutral";
});
</script>
