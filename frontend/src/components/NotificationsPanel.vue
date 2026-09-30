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
              v-for="alert in pagedAlerts"
              :key="alert.id"
              role="listitem"
              :class="[
                'p-4 transition-colors',
                alert.acknowledged
                  ? 'opacity-60 hover:bg-muted/40'
                  : 'bg-primary/5 hover:bg-primary/10',
              ]"
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
                    <span
                      :class="[
                        'text-sm truncate',
                        alert.acknowledged
                          ? 'font-medium text-muted-foreground'
                          : 'font-semibold',
                      ]"
                    >
                      {{ alert.title }}
                    </span>
                    <Badge :tone="severityTone(alert.severity)">
                      {{ alert.severity }}
                    </Badge>
                    <span
                      v-if="!alert.acknowledged"
                      class="inline-flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wide text-primary"
                    >
                      <span
                        class="w-1.5 h-1.5 rounded-full bg-primary"
                        aria-hidden="true"
                      />
                      <span class="sr-only">Unread</span>
                    </span>
                  </div>
                  <div class="text-xs text-muted-foreground mt-1">
                    {{ alert.server }} • {{ alert.time }}
                  </div>
                  <div class="mt-2 flex items-center gap-2">
                    <Button
                      v-if="!alert.acknowledged"
                      variant="outline"
                      size="sm"
                      @click="acknowledge(alert.id)"
                    >
                      <Check class="w-3.5 h-3.5" aria-hidden="true" />
                      Acknowledge
                    </Button>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 text-xs text-muted-foreground italic"
                    >
                      <Check class="w-3.5 h-3.5" aria-hidden="true" />
                      Read
                    </span>
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

        <Pagination
          :page="page"
          :total="filteredAlerts.length"
          :page-size="pageSize"
          aria-label="Notifications pagination"
          class="px-4 py-2 border-t border-border"
          @update:page="page = $event"
        />

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
            Showing {{ rangeStart }}–{{ rangeEnd }} of {{ filteredAlerts.length }}
          </span>
        </footer>
      </aside>
    </Transition>
  </Teleport>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from "vue";
import { Bell, X, Check, CheckCheck } from "lucide-vue-next";
import { useNotificationsPanel } from "@/composables/useNotificationsPanel";
import { useNotificationsStore } from "@/stores/notifications";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Select from "@/components/ui/Select.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import Pagination from "@/components/ui/Pagination.vue";

const panel = useNotificationsPanel();
const store = useNotificationsStore();

const filter = ref<"all" | "critical" | "warning" | "info">("all");

const pageSize = 10;
const page = ref(1);

const alerts = computed(() =>
  store.items.map((n) => ({
    id: n.id,
    title: n.title,
    server: n.serverId,
    time: formatTime(n.createdAt),
    severity: n.severity,
    acknowledged: n.read,
  })),
);

const filteredAlerts = computed(() =>
  filter.value === "all"
    ? alerts.value
    : alerts.value.filter((a) => a.severity === filter.value),
);

const pagedAlerts = computed(() => {
  const start = (page.value - 1) * pageSize;
  return filteredAlerts.value.slice(start, start + pageSize);
});

const rangeStart = computed(() =>
  filteredAlerts.value.length === 0 ? 0 : (page.value - 1) * pageSize + 1,
);
const rangeEnd = computed(() =>
  Math.min(page.value * pageSize, filteredAlerts.value.length),
);

const unreadCount = computed(() => store.unreadCount);
const hasUnread = computed(() => unreadCount.value > 0);

function formatTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function severityTone(s: string): "critical" | "warning" | "info" | "neutral" {
  if (s === "critical") return "critical";
  if (s === "warning") return "warning";
  if (s === "info") return "info";
  return "neutral";
}

async function acknowledge(id: string) {
  try {
    await store.markRead(id);
  } catch {
    /* error surfaced via store.error */
  }
}

async function markAllRead() {
  try {
    await store.markAllRead();
  } catch {
    /* error surfaced via store.error */
  }
}

// Refresh from the API whenever the panel opens, so a reconnecting tab
// catches up on anything missed while the socket was down.
watch(panel.isOpen, (open) => {
  if (open) void store.load();
});

// Reset to the first page when the severity filter changes, and clamp the
// current page when the underlying list shrinks.
watch(filter, () => {
  page.value = 1;
});

watch(
  () => filteredAlerts.value.length,
  (length) => {
    const maxPage = Math.max(1, Math.ceil(length / pageSize));
    if (page.value > maxPage) page.value = maxPage;
  },
);
</script>
