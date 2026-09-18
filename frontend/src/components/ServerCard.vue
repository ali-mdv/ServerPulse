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
        CPU: {{ server.cpu }}% • Mem: {{ server.mem }}%
      </div>
    </div>
    <router-link :to="`/server/${server.id}`">
      <Button variant="outline" size="sm" class="self-start sm:self-auto">
        View
        <ChevronRight class="w-4 h-4" aria-hidden="true" />
      </Button>
    </router-link>
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { ChevronRight } from "lucide-vue-next";
import type { Server } from "@/api/servers";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const props = defineProps<{ server: Server }>();

const statusTone = computed<"success" | "critical" | "neutral">(() => {
  if (props.server.status === "online") return "success";
  if (props.server.status === "down") return "critical";
  return "neutral";
});
</script>
