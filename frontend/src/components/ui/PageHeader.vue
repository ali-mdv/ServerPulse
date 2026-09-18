<template>
  <header
    :class="
      cn(
        'flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3',
        $attrs.class as string,
      )
    "
  >
    <div class="min-w-0">
      <h1 class="text-2xl font-semibold truncate">{{ title }}</h1>
      <p v-if="subtitle" class="text-sm text-muted-foreground mt-1">
        {{ subtitle }}
      </p>
    </div>
    <div
      v-if="$slots.default"
      class="flex items-center gap-2 shrink-0 self-start sm:self-auto"
    >
      <slot />
    </div>
  </header>
</template>

<script setup lang="ts">
import { useAttrs } from "vue";
import { cn } from "@/lib/utils";

interface Props {
  title: string;
  subtitle?: string;
}

defineProps<Props>();

defineOptions({ inheritAttrs: false });
useAttrs();
</script>
