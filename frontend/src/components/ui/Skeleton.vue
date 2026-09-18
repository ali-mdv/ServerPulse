<template>
  <div
    :class="
      cn(
        'animate-pulse rounded-md bg-muted',
        $attrs.class as string,
      )
    "
    :style="style"
    aria-hidden="true"
  />
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { cn } from "@/lib/utils";

interface Props {
  width?: string;
  height?: string;
}

const props = withDefaults(defineProps<Props>(), {
  width: "100%",
  height: "1rem",
});

defineOptions({ inheritAttrs: false });
useAttrs();

const style = computed(() => ({
  width: props.width,
  height: props.height,
}));
</script>
