<template>
  <div class="min-h-screen flex items-center justify-center">
    <div class="text-center">
      <div class="text-8xl font-bold text-gray-300 mb-4">404</div>
      <h1 class="text-4xl font-semibold mb-2">Page Not Found</h1>
      <p class="text-muted mb-8 text-lg">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <router-link
        to="/dashboard"
        class="inline-block px-6 py-3 bg-primary text-primary-foreground rounded hover:opacity-90 transition"
      >
        Back to Dashboard
      </router-link>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";

defineProps<{}>();

const router = useRouter();
</script>
