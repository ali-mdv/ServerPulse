<template>
  <aside
    :class="[
      'bg-sidebar p-4 h-screen transition-all duration-200 overflow-hidden',
      collapsed ? 'w-20' : 'w-64',
    ]"
  >
    <div class="flex items-center gap-3 mb-6">
      <div
        class="bg-primary h-10 w-10 rounded flex items-center justify-center text-white font-bold"
      >
        SM
      </div>
      <div v-if="!collapsed" class="text-lg font-semibold">{{ appName }}</div>
    </div>

    <nav class="flex flex-col gap-2">
      <router-link to="/dashboard" class="block">
        <div
          :class="[
            'px-3 py-2 rounded flex items-center',
            collapsed ? 'justify-center' : 'gap-3',
            isActive('/dashboard')
              ? 'bg-sidebar-accent text-primary'
              : 'hover:bg-sidebar-accent',
          ]"
          title="Dashboard"
        >
          <Grid class="w-4 h-4" />
          <span v-if="!collapsed">Dashboard</span>
        </div>
      </router-link>

      <router-link to="/servers" class="block">
        <div
          :class="[
            'px-3 py-2 rounded flex items-center',
            collapsed ? 'justify-center' : 'gap-3',
            isActive('/servers')
              ? 'bg-sidebar-accent text-primary'
              : 'hover:bg-sidebar-accent',
          ]"
          title="Servers"
        >
          <Server class="w-4 h-4" />
          <span v-if="!collapsed">Servers</span>
        </div>
      </router-link>

      <router-link to="/users" class="block">
        <div
          :class="[
            'px-3 py-2 rounded flex items-center',
            collapsed ? 'justify-center' : 'gap-3',
            isActive('/users')
              ? 'bg-sidebar-accent text-primary'
              : 'hover:bg-sidebar-accent',
          ]"
          title="Users"
        >
          <User class="w-4 h-4" />
          <span v-if="!collapsed">Users</span>
        </div>
      </router-link>

      <router-link to="/alerts" class="block">
        <div
          :class="[
            'px-3 py-2 rounded flex items-center',
            collapsed ? 'justify-center' : 'gap-3',
            isActive('/alerts')
              ? 'bg-sidebar-accent text-primary'
              : 'hover:bg-sidebar-accent',
          ]"
          title="Alerts"
        >
          <Bell class="w-4 h-4" />
          <span v-if="!collapsed">Alerts</span>
        </div>
      </router-link>

      <router-link to="/settings" class="block">
        <div
          :class="[
            'px-3 py-2 rounded flex items-center',
            collapsed ? 'justify-center' : 'gap-3',
            isActive('/settings')
              ? 'bg-sidebar-accent text-primary'
              : 'hover:bg-sidebar-accent',
          ]"
          title="Settings"
        >
          <SettingsIcon class="w-4 h-4" />
          <span v-if="!collapsed">Settings</span>
        </div>
      </router-link>
    </nav>

    <button
      type="button"
      @click="$emit('toggle')"
      class="mt-auto w-full mt-6 px-3 py-2 bg-muted rounded text-sm flex items-center justify-center"
      :aria-pressed="collapsed"
      :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
    >
      <ChevronRight v-if="collapsed" class="w-4 h-4" />
      <ChevronLeft v-else class="w-4 h-4" />
    </button>
  </aside>
</template>

<script lang="ts" setup>
import { toRef, computed } from "vue";
import { useRoute } from "vue-router";
import {
  Grid,
  Server,
  Bell,
  Settings as SettingsIcon,
  User,
  ChevronLeft,
  ChevronRight,
} from "lucide-vue-next";
import { useRootStore } from "@/stores/root";
const props = defineProps<{ collapsed: boolean }>();
const collapsed = toRef(props, "collapsed");
const rootStore = useRootStore();
const route = useRoute();

const appName = computed(() => rootStore.appName);

function isActive(path: string) {
  if (!route || !("path" in route) || !route.path) return false;
  return route.path === path || route.path.startsWith(path + "/");
}
</script>
