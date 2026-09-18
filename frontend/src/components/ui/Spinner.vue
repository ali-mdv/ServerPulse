<template>
  <span
    :class="
      cn(
        'inline-block animate-spin rounded-full border-2 border-current border-r-transparent',
        sizeClass,
        $attrs.class as string,
      )
    "
    role="status"
    aria-hidden="true"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { cn } from "@/lib/utils";

interface Props {
  size?: "xs" | "sm" | "md" | "lg";
}
const props = withDefaults(defineProps<Props>(), { size: "md" });

const sizeClass = computed(
  () =>
    ({
      xs: "w-3 h-3 border",
      sm: "w-3.5 h-3.5 border-2",
      md: "w-4 h-4 border-2",
      lg: "w-6 h-6 border-[3px]",
    })[props.size],
);
</script>
