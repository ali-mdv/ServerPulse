<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="mobileOpen"
        class="md:hidden fixed inset-0 bg-black/50 z-40"
        @click="$emit('close')"
        aria-hidden="true"
      />
    </Transition>

    <Transition
      enter-active-class="transition-transform duration-200 ease-out"
      leave-active-class="transition-transform duration-200 ease-in"
      enter-from-class="-translate-x-full"
      leave-to-class="-translate-x-full"
    >
      <aside
        v-if="mobileOpen"
        class="md:hidden fixed inset-y-0 left-0 z-50 bg-sidebar text-sidebar-foreground w-72 p-4 flex flex-col shadow-xl"
        aria-label="Primary navigation"
        role="dialog"
        aria-modal="true"
      >
        <SidebarBody
          :collapsed="false"
          @navigate="$emit('close')"
          @toggle="$emit('toggle')"
        />
      </aside>
    </Transition>

    <aside
      :class="[
        'hidden md:flex sticky top-0 h-screen bg-sidebar text-sidebar-foreground p-4 flex-col transition-all duration-200 overflow-hidden',
        collapsed ? 'md:w-20' : 'md:w-64',
      ]"
      aria-label="Primary navigation"
    >
      <SidebarBody :collapsed="collapsed" @toggle="$emit('toggle')" />
    </aside>
  </Teleport>
</template>

<script lang="ts" setup>
import { onMounted, onUnmounted } from "vue";
import SidebarBody from "./SidebarBody.vue";

defineProps<{
  collapsed: boolean;
  mobileOpen: boolean;
}>();

defineEmits<{
  toggle: [];
  close: [];
}>();

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    document.dispatchEvent(new CustomEvent("sidebar-close"));
  }
}

onMounted(() => {
  document.addEventListener("keydown", onKeydown);
});
onUnmounted(() => {
  document.removeEventListener("keydown", onKeydown);
});
</script>
