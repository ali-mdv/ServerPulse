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

    <div
      v-if="store.loading && alerts.length === 0"
      class="flex justify-center py-10"
    >
      <Spinner size="lg" />
    </div>

    <EmptyState
      v-else-if="filtered.length === 0"
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
import { ref, computed, onMounted } from "vue";
import { Check } from "lucide-vue-next";
import Card from "@/components/Card.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import Spinner from "@/components/ui/Spinner.vue";
import { useNotificationsStore } from "@/stores/notifications";

const store = useNotificationsStore();

const severity = ref("all");
const from = ref("");
const to = ref("");

const alerts = computed(() =>
  store.items.map((n) => ({
    id: n.id,
    title: n.title,
    server: n.serverId,
    createdAt: n.createdAt,
    time: formatDateTime(n.createdAt),
    severity: n.severity,
    resolved: n.read,
  })),
);

const filtered = computed(() =>
  alerts.value.filter((a) => {
    if (severity.value !== "all" && a.severity !== severity.value) return false;

    const ts = new Date(a.createdAt).getTime();
    if (Number.isNaN(ts)) return true;
    if (from.value && ts < startOfDay(from.value)) return false;
    if (to.value && ts > endOfDay(to.value)) return false;
    return true;
  }),
);

onMounted(() => {
  void store.load();
});

function startOfDay(date: string): number {
  return new Date(`${date}T00:00:00`).getTime();
}

function endOfDay(date: string): number {
  return new Date(`${date}T23:59:59.999`).getTime();
}

function formatDateTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString([], {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function severityTone(s: string): "critical" | "warning" | "info" | "neutral" {
  if (s === "critical") return "critical";
  if (s === "warning") return "warning";
  if (s === "info") return "info";
  return "neutral";
}

async function resolve(id: string) {
  try {
    await store.markRead(id);
  } catch {
    /* error surfaced via store.error */
  }
}
</script>