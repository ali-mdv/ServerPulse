<template>
  <div class="space-y-6">
    <PageHeader
      title="Alerts & Events"
      :subtitle="`${filtered.length} of ${alerts.length} shown`"
    />

    <div class="flex flex-wrap items-center gap-2">
      <Select v-model="severity" aria-label="Filter by severity" class="w-auto">
        <option value="all">All severities</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </Select>
      <Input
        type="date"
        v-model="from"
        aria-label="From date"
        class="w-auto"
      />
      <Input type="date" v-model="to" aria-label="To date" class="w-auto" />
    </div>

    <EmptyState
      v-if="filtered.length === 0"
      title="No alerts"
      description="Nothing matches the current filters."
    />

    <div v-else class="grid gap-3">
      <Card v-for="alert in filtered" :key="alert.id" class="!p-4">
        <div
          class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
        >
          <div class="min-w-0">
            <div class="font-semibold truncate">{{ alert.title }}</div>
            <div class="text-xs text-muted-foreground mt-1">
              {{ alert.server }} • {{ alert.time }}
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <Badge :tone="severityTone(alert.severity)">
              {{ alert.severity }}
            </Badge>
            <Button
              v-if="!alert.resolved"
              variant="primary"
              size="sm"
              @click="resolve(alert.id)"
            >
              <Check class="w-3.5 h-3.5" aria-hidden="true" />
              Resolve
            </Button>
            <span v-else class="text-xs text-muted-foreground italic">
              Resolved
            </span>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { Check } from "lucide-vue-next";
import Card from "@/components/Card.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";

const severity = ref("all");
const from = ref("");
const to = ref("");

const alerts = ref([
  {
    id: "a1",
    title: "High CPU on web-01",
    server: "web-01",
    time: "2025-10-24 10:12",
    severity: "critical",
    resolved: false,
  },
  {
    id: "a2",
    title: "Disk near full on db-01",
    server: "db-01",
    time: "2025-10-24 09:50",
    severity: "warning",
    resolved: false,
  },
  {
    id: "a3",
    title: "Container restart on cache-01",
    server: "cache-01",
    time: "2025-10-23 22:05",
    severity: "info",
    resolved: true,
  },
]);

const filtered = computed(() => {
  return alerts.value.filter((a) => {
    if (severity.value !== "all" && a.severity !== severity.value) return false;
    return true;
  });
});

function severityTone(s: string): "critical" | "warning" | "info" | "neutral" {
  if (s === "critical") return "critical";
  if (s === "warning") return "warning";
  if (s === "info") return "info";
  return "neutral";
}

function resolve(id: string) {
  const found = alerts.value.find((a) => a.id === id);
  if (found) found.resolved = true;
}
</script>
