<template>
  <div class="flex w-full min-h-screen">
    <Sidebar
      v-if="!isAuthRoute"
      :collapsed="collapsed"
      :mobile-open="mobileOpen"
      @toggle="toggleCollapsed"
      @close="mobileOpen = false"
    />
    <div
      :class="[
        'flex flex-col min-w-0 flex-1 min-h-screen',
        isAuthRoute && 'w-full',
      ]"
    >
      <TopNav
        v-if="!isAuthRoute"
        :show-controls="!isAuthRoute"
        :hide-sidebar-toggle="false"
        @toggleSidebar="onTopNavToggle"
      />
      <a
        v-if="!isAuthRoute"
        href="#main-content"
        class="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[100] focus:rounded-md focus:bg-popover focus:px-3 focus:py-2 focus:text-sm focus:shadow-lg"
      >
        Skip to content
      </a>
      <main id="main-content" tabindex="-1" class="flex-1 p-4 sm:p-6">
        <router-view v-slot="{ Component, route }">
          <Transition
            enter-active-class="transition-all duration-200 ease-out"
            enter-from-class="opacity-0 translate-y-2"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition-all duration-150 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
            mode="out-in"
          >
            <component :is="Component" :key="route.fullPath" />
          </Transition>
        </router-view>
      </main>
      <Toast />
    </div>
    <NotificationsPanel />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, watch, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import { useMediaQuery, useNotificationsPanel } from "@/composables";
import { useScrollLock } from "@vueuse/core";
import Sidebar from "./Sidebar.vue";
import TopNav from "./TopNav.vue";
import NotificationsPanel from "./NotificationsPanel.vue";

const collapsed = ref(false);
const mobileOpen = ref(false);
const route = useRoute();
const notifications = useNotificationsPanel();

const isDesktop = useMediaQuery("(min-width: 768px)");

const isAuthRoute = computed(() =>
  ["login", "serverError", "notFound"].includes(String(route.name)),
);

const sidebarScrollLocked = useScrollLock(document.body);
const notificationsScrollLocked = useScrollLock(document.body);

watch(mobileOpen, (open) => {
  sidebarScrollLocked.value = open && !isDesktop.value;
});

watch(
  () => route.fullPath,
  () => {
    if (!isDesktop.value) mobileOpen.value = false;
  },
);

watch(isDesktop, (desktop) => {
  if (desktop) {
    mobileOpen.value = false;
    sidebarScrollLocked.value = false;
  }
});

watch(notifications.isOpen, (open) => {
  notificationsScrollLocked.value = open;
});

function toggleCollapsed() {
  if (isDesktop.value) {
    collapsed.value = !collapsed.value;
  } else {
    mobileOpen.value = !mobileOpen.value;
  }
}

function onTopNavToggle() {
  if (isDesktop.value) {
    collapsed.value = !collapsed.value;
  } else {
    mobileOpen.value = true;
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    if (mobileOpen.value) mobileOpen.value = false;
    else if (notifications.isOpen.value) notifications.close();
  }
}

window.addEventListener("keydown", onKeydown);
onUnmounted(() => window.removeEventListener("keydown", onKeydown));
</script>
