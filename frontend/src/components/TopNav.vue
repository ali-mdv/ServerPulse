<template>
  <header
    class="flex items-center justify-between px-6 py-3 border-b border-sidebar-border bg-transparent"
  >
    <slot name="before"></slot>
    <div v-if="showControls" class="flex items-center gap-4">
      <button
        @click="$emit('toggleSidebar')"
        class="p-2 rounded hover:bg-muted"
        aria-label="Toggle sidebar"
      >
        ☰
      </button>
      <div class="hidden md:flex items-center gap-3">
        <h1 class="text-xl font-semibold">Dashboard</h1>
        <div class="relative">
          <input
            v-model="q"
            aria-label="Search"
            placeholder=""
            class="border rounded pl-9 pr-3 py-1 w-64 bg-card"
          />
          <Search
            class="w-4 h-4 absolute left-2 top-1/2 -translate-y-1/2 text-muted"
          />
        </div>
      </div>
    </div>

    <div v-if="showControls" class="flex items-center gap-4">
      <button
        @click="toggleTheme"
        class="p-2 rounded hover:bg-muted"
        :aria-pressed="isDark"
        aria-label="Toggle theme"
      >
        <Sun v-if="!isDark" class="w-5 h-5" />
        <Moon v-else class="w-5 h-5" />
      </button>

      <div class="relative">
        <button
          @click="openNotifications = !openNotifications"
          class="p-2 rounded hover:bg-muted"
          aria-label="Notifications"
        >
          <Bell class="w-5 h-5" />
        </button>

        <div
          v-if="openNotifications"
          class="absolute right-0 mt-2 w-80 bg-card rounded shadow p-2 z-50"
        >
          <div class="font-semibold px-2 pb-2">Recent alerts</div>
          <div
            v-for="a in recentAlerts"
            :key="a.id"
            class="p-2 border-b last:border-b-0 flex justify-between"
          >
            <div>
              <div class="font-medium">{{ a.title }}</div>
              <div class="text-sm text-muted">
                {{ a.time }} • {{ a.server }}
              </div>
            </div>
            <div
              :class="[
                'text-sm px-2 py-1 rounded',
                a.severity === 'critical'
                  ? 'bg-red-100 text-red-700'
                  : a.severity === 'warning'
                  ? 'bg-yellow-100 text-yellow-800'
                  : 'bg-gray-100 text-gray-800',
              ]"
            >
              {{ a.severity }}
            </div>
          </div>
          <div class="text-center mt-2">
            <router-link to="/alerts" class="text-sm text-primary"
              >View all alerts</router-link
            >
          </div>
        </div>
      </div>

      <div class="relative">
        <button
          @click="openUser = !openUser"
          class="flex items-center gap-2 p-1 rounded hover:bg-muted"
        >
          <div
            class="w-8 h-8 bg-primary rounded-full flex items-center justify-center text-white"
          >
            A
          </div>
          <span class="hidden md:inline">Admin</span>
        </button>
        <div
          v-if="openUser"
          class="absolute right-0 mt-2 w-48 bg-card rounded shadow p-2"
        >
          <button
            @click="goToProfile"
            class="w-full text-left px-2 py-1 rounded hover:bg-muted"
          >
            Profile
          </button>
          <button
            @click="goToSettings"
            class="w-full text-left px-2 py-1 rounded hover:bg-muted"
          >
            Settings
          </button>
          <button
            @click="logout"
            class="w-full text-left px-2 py-1 rounded hover:bg-muted"
          >
            Sign out
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<script lang="ts" setup>
import { ref, onMounted, toRef } from "vue";
import { useRouter } from "vue-router";
import { Search, Sun, Moon, Bell } from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();

const props = withDefaults(defineProps<{ showControls?: boolean }>(), {
  showControls: true,
});
const showControls = toRef(props, "showControls");

const q = ref("");
const isDark = ref(document.documentElement.classList.contains("dark"));
const openUser = ref(false);
const openNotifications = ref(false);

const recentAlerts = ref([
  {
    id: "a1",
    title: "High CPU on web-01",
    server: "web-01",
    time: "10:12",
    severity: "critical",
  },
  {
    id: "a2",
    title: "Disk near full on db-01",
    server: "db-01",
    time: "09:50",
    severity: "warning",
  },
  {
    id: "a3",
    title: "Container restart on cache-01",
    server: "cache-01",
    time: "22:05",
    severity: "info",
  },
]);

const router = useRouter();

function goToSettings() {
  router.push("/settings");
}
function goToProfile() {
  openUser.value = false;
  router.push("/profile");
}
function logout() {
  auth.logout();
  router.push("/login");
}

function toggleTheme() {
  isDark.value = !isDark.value;
  if (isDark.value) document.documentElement.classList.add("dark");
  else document.documentElement.classList.remove("dark");
  localStorage.setItem("theme", isDark.value ? "dark" : "light");
}

onMounted(() => {
  const t = localStorage.getItem("theme");
  if (t === "dark") {
    isDark.value = true;
    document.documentElement.classList.add("dark");
  } else if (t === "light") {
    isDark.value = false;
    document.documentElement.classList.remove("dark");
  }
});
</script>
