<template>
  <button
    :type="type"
    :class="cn(buttonClasses, $attrs.class as string)"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    v-bind="$attrs"
  >
    <Spinner v-if="loading" :size="spinnerSize" />
    <slot v-else name="icon-left" />
    <slot />
    <slot name="icon-right" />
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { cn } from "@/lib/utils";
import Spinner from "./Spinner.vue";

type Variant =
  | "primary"
  | "secondary"
  | "outline"
  | "ghost"
  | "destructive"
  | "success"
  | "danger"
  | "info";

type Size = "sm" | "md" | "lg" | "icon";

interface Props {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  disabled?: boolean;
  type?: "button" | "submit" | "reset";
}

const props = withDefaults(defineProps<Props>(), {
  variant: "primary",
  size: "md",
  loading: false,
  disabled: false,
  type: "button",
});

const buttonClasses = computed(() => {
  const variantClass = {
    primary: "btn-primary",
    secondary: "btn-secondary",
    outline: "btn-outline",
    ghost: "btn-ghost",
    destructive: "btn-destructive",
    success: "btn-success",
    danger: "btn-danger",
    info: "btn-info",
  }[props.variant];

  const sizeClass = {
    sm: "btn-sm",
    md: "btn-md",
    lg: "btn-lg",
    icon: "btn-icon",
  }[props.size];

  return cn(variantClass, sizeClass);
});

const spinnerSize = computed(
  () => ({ sm: "sm", md: "sm", lg: "md", icon: "sm" })[props.size],
);
</script>
