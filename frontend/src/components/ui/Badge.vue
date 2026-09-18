<template>
  <span :class="cn(badgeClass, $attrs.class as string)">
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { cn } from "@/lib/utils";

type Tone = "neutral" | "critical" | "warning" | "info" | "success";

interface Props {
  tone?: Tone;
}

const props = withDefaults(defineProps<Props>(), { tone: "neutral" });

const badgeClass = computed(
  () =>
    ({
      neutral: "badge-neutral",
      critical: "badge-critical",
      warning: "badge-warning",
      info: "badge-info",
      success: "badge-success",
    })[props.tone],
);
</script>
