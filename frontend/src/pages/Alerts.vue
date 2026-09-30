<template>
  <div class="space-y-6">
    <PageHeader
      title="Alerts & Events"
      :subtitle="`${filtered.length} of ${alerts.length} shown`"
    />

    <div class="flex flex-wrap items-center gap-2">
      <div class="relative">
        <Search
          class="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
          aria-hidden="true"
        />
        <Input
          type="search"
          v-model="q"
          aria-label="Search alerts"
          placeholder="Search alerts"
          class="pl-9 w-64"
        />
      </div>
      <Select v-model="severity" aria-label="Filter by severity" class="w-auto">
        <option value="all">All severities</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </Select>
      <Select v-model="status" aria-label="Filter by status" class="w-auto">
        <option value="all">All statuses</option>
        <option value="unresolved">Unresolved</option>
        <option value="resolved">Resolved</option>
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
import { Check, Search } from "lucide-vue-next";
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

const q = ref("");
const severity = ref("all");
const status = ref("all");
const from = ref("");
const to = ref("");

const alerts = computed(() =>
  store.items.map((n) => ({
    id: n.id,
    title: n.title,
    message: n.message ?? "",
    server: n.serverId,
    createdAt: n.createdAt,
    time: formatDateTime(n.createdAt),
    severity: n.severity,
    resolved: n.read,
  })),
);

const filtered = computed(() => {
  const term = q.value.trim().toLowerCase();
  return alerts.value.filter((a) => {
    if (term) {
      const haystack = `${a.title} ${a.message} ${a.server}`.toLowerCase();
      if (!haystack.includes(term)) return false;
    }
    if (severity.value !== "all" && a.severity !== severity.value) return false;
    if (status.value === "unresolved" && a.resolved) return false;
    if (status.value === "resolved" && !a.resolved) return false;

    const ts = new Date(a.createdAt).getTime();
    if (Number.isNaN(ts)) return true;
    if (from.value && ts < startOfDay(from.value)) return false;
    if (to.value && ts > endOfDay(to.value)) return false;
    return true;
  });
});

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