<template>
  <div
    :class="
      cn(
        'card card-body text-center flex flex-col items-center gap-2',
        $attrs.class as string,
      )
    "
  >
    <div
      v-if="$slots.icon"
      class="w-12 h-12 rounded-full bg-muted flex items-center justify-center text-muted-foreground"
      aria-hidden="true"
    >
      <slot name="icon" />
    </div>
    <h3 v-if="title" class="text-base font-semibold text-foreground">
      {{ title }}
    </h3>
    <p v-if="description" class="text-sm text-muted-foreground max-w-sm">
      {{ description }}
    </p>
    <div v-if="$slots.action" class="mt-2">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { cn } from "@/lib/utils";

interface Props {
  title?: string;
  description?: string;
}

defineProps<Props>();
</script>
