<template>
  <div class="flex w-full">
    <Sidebar
      v-if="!isAuthRoute"
      :collapsed="collapsed"
      @toggle="collapsed = !collapsed"
    />
    <div
      :class="[isAuthRoute ? 'w-full' : 'flex-1 min-h-screen flex flex-col']"
    >
      <TopNav
        v-if="!isAuthRoute"
        :show-controls="!isAuthRoute"
        @toggleSidebar="collapsed = !collapsed"
      />
      <main class="p-6">
        <router-view />
      </main>
      <Toast />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { useRoute } from "vue-router";
import Sidebar from "./Sidebar.vue";
import TopNav from "./TopNav.vue";

const collapsed = ref(false);
const route = useRoute();
const isAuthRoute = computed(() => {
  return ["login", "serverError", "notFound"].includes(String(route.name));
});
</script>
