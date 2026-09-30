<template>
  <div class="flex items-center gap-3 mb-6 min-w-0">
    <div
      class="bg-primary text-primary-foreground h-10 w-10 rounded flex items-center justify-center font-bold shrink-0"
      aria-hidden="true"
    >
      SP
    </div>
    <div v-if="!collapsed" class="text-base font-semibold truncate">
      {{ appName }}
    </div>
  </div>

  <nav class="flex flex-col gap-1" aria-label="Sections">
    <router-link
      v-for="item in navItems"
      :key="item.to"
      :to="item.to"
      :class="[
        'nav-link',
        collapsed && 'nav-link-icon-only',
        isActive(item.to) && 'nav-link-active',
      ]"
      :aria-current="isActive(item.to) ? 'page' : undefined"
      :title="collapsed ? item.label : undefined"
      @click="$emit('navigate')"
    >
      <component
        :is="item.icon"
        class="w-4 h-4 shrink-0"
        aria-hidden="true"
      />
      <span v-if="!collapsed">{{ item.label }}</span>
    </router-link>
  </nav>

  <button
    v-if="!collapsed || true"
    type="button"
    @click="$emit('toggle')"
    class="mt-auto btn btn-ghost btn-sm justify-center"
    :aria-pressed="collapsed"
    :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
  >
    <ChevronRight v-if="collapsed" class="w-4 h-4" aria-hidden="true" />
    <ChevronLeft v-else class="w-4 h-4" aria-hidden="true" />
    <span v-if="!collapsed">Collapse</span>
  </button>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { useRoute } from "vue-router";
import {
  LayoutDashboard,
  Server,
  Users,
  Bell,
  Settings,
  ChevronLeft,
  ChevronRight,
  type LucideIcon,
} from "lucide-vue-next";
import { useRootStore } from "@/stores/root";

defineProps<{ collapsed: boolean }>();

defineEmits<{
  toggle: [];
  navigate: [];
}>();

const rootStore = useRootStore();
const route = useRoute();

const appName = computed(() => rootStore.appName);

const navItems: { to: string; label: string; icon: LucideIcon }[] = [
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/servers", label: "Servers", icon: Server },
  { to: "/users", label: "Users", icon: Users },
  { to: "/alerts", label: "Alerts", icon: Bell },
  { to: "/settings", label: "Settings", icon: Settings },
];

function isActive(path: string) {
  if (!route || !("path" in route) || !route.path) return false;
  return route.path === path || route.path.startsWith(path + "/");
}
</script>
