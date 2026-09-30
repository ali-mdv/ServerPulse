<template>
  <div class="min-h-[60vh] flex items-center justify-center px-4">
    <div class="text-center max-w-md">
      <div
        class="text-7xl sm:text-8xl font-bold text-muted-foreground mb-2"
        aria-hidden="true"
      >
        404
      </div>
      <h1 class="text-2xl sm:text-3xl font-semibold mb-2">Page not found</h1>
      <p class="text-muted-foreground mb-8">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <router-link to="/dashboard">
        <Button variant="primary">
          <ArrowLeft class="w-4 h-4" aria-hidden="true" />
          Back to dashboard
        </Button>
      </router-link>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ArrowLeft } from "lucide-vue-next";
import Button from "@/components/ui/Button.vue";
</script>
