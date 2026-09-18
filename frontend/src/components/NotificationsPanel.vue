<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="panel.isOpen.value"
        class="fixed inset-0 bg-black/40 z-40"
        aria-hidden="true"
        @click="panel.close()"
      />
    </Transition>

    <Transition
      enter-active-class="transition-transform duration-200 ease-out"
      leave-active-class="transition-transform duration-200 ease-in"
      enter-from-class="translate-x-full"
      leave-to-class="translate-x-full"
    >
      <aside
        v-if="panel.isOpen.value"
        class="fixed inset-y-0 right-0 z-50 w-full sm:w-96 bg-card text-card-foreground border-l border-border shadow-2xl flex flex-col"
        role="dialog"
        aria-modal="true"
        aria-labelledby="notifications-title"
      >
        <header
          class="flex items-center justify-between px-4 py-3 border-b border-border"
        >
          <div>
            <h2
              id="notifications-title"
              class="text-base font-semibold flex items-center gap-2"
            >
              <Bell class="w-4 h-4" aria-hidden="true" />
              Notifications
            </h2>
            <p class="text-xs text-muted-foreground mt-0.5">
              {{ unreadCount }} unread
            </p>
          </div>
          <Button
            variant="ghost"
            size="icon"
            aria-label="Close notifications"
            @click="panel.close()"
          >
            <X class="w-5 h-5" aria-hidden="true" />
          </Button>
        </header>

        <div class="flex items-center gap-2 px-4 py-2 border-b border-border">
          <Select v-model="filter" aria-label="Filter by severity" class="flex-1">
            <option value="all">All</option>
            <option value="critical">Critical</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </Select>
          <Button
            variant="ghost"
            size="sm"
            :disabled="!hasUnread"
            @click="markAllRead"
          >
            <CheckCheck class="w-4 h-4" aria-hidden="true" />
            Mark all read
          </Button>
        </div>

        <div class="flex-1 overflow-y-auto" role="list">
          <EmptyState
            v-if="filteredAlerts.length === 0"
            title="No notifications"
            description="You're all caught up."
          />
          <ul v-else class="divide-y divide-border">
            <li
              v-for="alert in filteredAlerts"
              :key="alert.id"
              role="listitem"
              class="p-4 hover:bg-muted/40 transition-colors"
            >
              <div class="flex items-start gap-3">
                <span
                  :class="[
                    'mt-1.5 w-2 h-2 rounded-full shrink-0',
                    alert.severity === 'critical' && 'bg-critical',
                    alert.severity === 'warning' && 'bg-warning',
                    alert.severity === 'info' && 'bg-info',
                  ]"
                  aria-hidden="true"
                />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="text-sm font-medium truncate">
                      {{ alert.title }}
                    </span>
                    <Badge :tone="severityTone(alert.severity)">
                      {{ alert.severity }}
                    </Badge>
                  </div>
                  <div class="text-xs text-muted-foreground mt-1">
                    {{ alert.server }} • {{ alert.time }}
                  </div>
                  <div class="mt-2 flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      @click="acknowledge(alert.id)"
                    >
                      <Check class="w-3.5 h-3.5" aria-hidden="true" />
                      Acknowledge
                    </Button>
                    <router-link
                      to="/alerts"
                      class="text-xs text-primary hover:underline focus-visible:underline"
                      @click="panel.close()"
                    >
                      View details
                    </router-link>
                  </div>
                </div>
              </div>
            </li>
          </ul>
        </div>

        <footer
          class="px-4 py-3 border-t border-border flex items-center justify-between"
        >
          <router-link
            to="/alerts"
            class="text-sm text-primary hover:underline focus-visible:underline"
            @click="panel.close()"
          >
            View all alerts
          </router-link>
          <span class="text-xs text-muted-foreground">
            Showing {{ filteredAlerts.length }} of {{ alerts.length }}
          </span>
        </footer>
      </aside>
    </Transition>
  </Teleport>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { Bell, X, Check, CheckCheck } from "lucide-vue-next";
import { useNotificationsPanel } from "@/composables/useNotificationsPanel";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Select from "@/components/ui/Select.vue";
import EmptyState from "@/components/ui/EmptyState.vue";

const panel = useNotificationsPanel();

const filter = ref<"all" | "critical" | "warning" | "info">("all");

const alerts = ref([
  {
    id: "a1",
    title: "High CPU on web-01",
    server: "web-01",
    time: "10:12",
    severity: "critical",
    acknowledged: false,
  },
  {
    id: "a2",
    title: "Disk near full on db-01",
    server: "db-01",
    time: "09:50",
    severity: "warning",
    acknowledged: false,
  },
  {
    id: "a3",
    title: "Container restart on cache-01",
    server: "cache-01",
    time: "22:05",
    severity: "info",
    acknowledged: true,
  },
  {
    id: "a4",
    title: "Memory pressure on api-01",
    server: "api-01",
    time: "08:31",
    severity: "warning",
    acknowledged: false,
  },
]);

const filteredAlerts = computed(() =>
  filter.value === "all"
    ? alerts.value
    : alerts.value.filter((a) => a.severity === filter.value),
);

const unreadCount = computed(
  () => alerts.value.filter((a) => !a.acknowledged).length,
);

const hasUnread = computed(() => unreadCount.value > 0);

function severityTone(s: string): "critical" | "warning" | "info" | "neutral" {
  if (s === "critical") return "critical";
  if (s === "warning") return "warning";
  if (s === "info") return "info";
  return "neutral";
}

function acknowledge(id: string) {
  const found = alerts.value.find((a) => a.id === id);
  if (found) found.acknowledged = true;
}

function markAllRead() {
  for (const a of alerts.value) a.acknowledged = true;
}
</script>
