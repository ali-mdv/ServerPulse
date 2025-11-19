<template>
  <div class="min-h-screen flex items-center justify-center">
    <div class="text-center">
      <div class="text-8xl font-bold text-red-400 mb-4">500</div>
      <h1 class="text-4xl font-semibold mb-2">Server Error</h1>
      <p class="text-muted mb-8 text-lg">
        An unexpected error occurred. Please try again later.
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
