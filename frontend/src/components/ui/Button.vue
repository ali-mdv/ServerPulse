<template>
  <button
    :type="type"
    :class="mergedClasses"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    v-bind="extraAttrs"
  >
    <Spinner v-if="loading" :size="spinnerSize" />
    <template v-else>
      <slot name="icon-left" />
      <slot />
      <slot name="icon-right" />
    </template>
  </button>
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
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

defineOptions({ inheritAttrs: false });
const attrs = useAttrs();

const variantClass = computed(
  () =>
    ({
      primary: "btn-primary",
      secondary: "btn-secondary",
      outline: "btn-outline",
      ghost: "btn-ghost",
      destructive: "btn-destructive",
      success: "btn-success",
      danger: "btn-danger",
      info: "btn-info",
    })[props.variant],
);

const sizeClass = computed(
  () =>
    ({
      sm: "btn-sm",
      md: "btn-md",
      lg: "btn-lg",
      icon: "btn-icon",
    })[props.size],
);

const mergedClasses = computed(() =>
  cn(variantClass.value, sizeClass.value, attrs.class as string | undefined),
);

const extraAttrs = computed(() => {
  const { class: _c, ...rest } = attrs as Record<string, unknown>;
  return rest;
});

const spinnerSize = computed(
  () => ({ sm: "sm", md: "sm", lg: "md", icon: "sm" })[props.size],
);
</script>
