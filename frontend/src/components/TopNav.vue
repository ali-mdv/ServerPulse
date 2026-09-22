<template>
  <header
    class="flex items-center justify-between gap-3 px-4 sm:px-6 py-3 border-b border-sidebar-border bg-background"
  >
    <slot name="before"></slot>

    <div v-if="showControls" class="flex items-center gap-2 sm:gap-3 min-w-0">
      <Button
        v-if="!hideSidebarToggle"
        variant="ghost"
        size="icon"
        aria-label="Toggle navigation"
        @click="$emit('toggleSidebar')"
      >
        <Menu class="w-5 h-5" aria-hidden="true" />
      </Button>

      <h1 class="text-lg sm:text-xl font-semibold truncate">
        {{ pageTitle }}
      </h1>
    </div>

    <div v-if="showControls" class="flex items-center gap-1 sm:gap-2">
      <Button
        variant="ghost"
        size="icon"
        class="relative"
        aria-label="Notifications"
        :aria-haspopup="true"
        :aria-expanded="notifications.isOpen.value"
        @click="notifications.toggle()"
      >
        <Bell class="w-5 h-5" aria-hidden="true" />
        <span
          v-if="unreadCount > 0"
          class="absolute top-1 right-1 inline-flex h-2 w-2 rounded-full bg-critical"
          aria-hidden="true"
        />
      </Button>

      <Button
        variant="ghost"
        size="icon"
        :aria-pressed="isDark"
        :aria-label="`Switch to ${isDark ? 'light' : 'dark'} theme`"
        @click="toggle"
      >
        <Sun v-if="!isDark" class="w-5 h-5" aria-hidden="true" />
        <Moon v-else class="w-5 h-5" aria-hidden="true" />
      </Button>

      <div ref="userRef" class="relative">
        <Button
          variant="ghost"
          class="gap-2 px-2 h-9"
          aria-haspopup="true"
          :aria-expanded="openUser"
          @click="openUser = !openUser"
        >
          <span
            class="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-semibold"
            aria-hidden="true"
          >
            {{ userInitial }}
          </span>
          <span class="hidden md:inline text-sm font-medium">
            {{ userName }}
          </span>
        </Button>

        <Transition
          enter-active-class="transition duration-100 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition duration-75 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
        >
          <div v-if="openUser" class="menu w-48" role="menu">
            <button class="menu-item" role="menuitem" @click="goToProfile">
              <UserIcon class="w-4 h-4" aria-hidden="true" /> Profile
            </button>
            <button class="menu-item" role="menuitem" @click="goToSettings">
              <SettingsIcon class="w-4 h-4" aria-hidden="true" /> Settings
            </button>
            <div class="border-t border-border my-1" />
            <button
              class="menu-item text-destructive focus-visible:text-destructive"
              role="menuitem"
              @click="logout"
            >
              <LogOut class="w-4 h-4" aria-hidden="true" /> Sign out
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { onClickOutside } from "@vueuse/core";
import {
  Bell,
  Sun,
  Moon,
  User as UserIcon,
  Settings as SettingsIcon,
  LogOut,
  Menu,
} from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";
import { useNotificationsStore } from "@/stores/notifications";
import { useSettingsStore } from "@/stores/settings";
import { useTheme, useNotificationsPanel } from "@/composables";
import Button from "@/components/ui/Button.vue";

const auth = useAuthStore();
const notificationsStore = useNotificationsStore();
const settingsStore = useSettingsStore();
const router = useRouter();
const route = useRoute();

defineProps<{
  showControls?: boolean;
  hideSidebarToggle?: boolean;
  sidebarCollapsed?: boolean;
}>();

defineEmits<{
  toggleSidebar: [];
}>();

const { isDark, theme, toggle: toggleTheme } = useTheme();
const notifications = useNotificationsPanel();

// Toggle locally, then persist the new appearance so it survives reloads
// and reaches other sessions.
async function toggle() {
  toggleTheme();
  try {
    await settingsStore.save({ appearance: theme.value });
  } catch {
    /* local theme is already applied; keep going even if the save fails */
  }
}

const openUser = ref(false);
const userRef = ref<HTMLElement | null>(null);

onClickOutside(userRef, () => (openUser.value = false));

const unreadCount = computed(() => notificationsStore.unreadCount);

const pageTitle = computed(() => {
  const map: Record<string, string> = {
    dashboard: "Dashboard",
    serversList: "Servers",
    serversDetail: "Server details",
    usersList: "Users",
    usersAdd: "Add user",
    profile: "Profile",
    alerts: "Alerts",
    settings: "Settings",
  };
  return map[String(route.name)] ?? "ServerPulse";
});

const userName = computed(() => auth.user?.email?.split("@")[0] ?? "Admin");
const userInitial = computed(() => userName.value.charAt(0).toUpperCase());

function goToSettings() {
  openUser.value = false;
  router.push("/settings");
}
function goToProfile() {
  openUser.value = false;
  router.push("/profile");
}
function logout() {
  openUser.value = false;
  auth.logout();
  router.push("/login");
}
</script>
