<template>
  <select
    :class="cn('input', $attrs.class as string)"
    :value="modelValue"
    @change="onChange"
    v-bind="remainingAttrs"
  >
    <slot />
  </select>
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { cn } from "@/lib/utils";

interface Props {
  modelValue?: string | number;
}
defineProps<Props>();

const emit = defineEmits<{
  "update:modelValue": [value: string | number];
}>();

const attrs = useAttrs();

const remainingAttrs = computed(() => {
  const { class: _c, ...rest } = attrs as Record<string, unknown>;
  return rest;
});

function onChange(e: Event) {
  const target = e.target as HTMLSelectElement;
  const value = target.value;
  emit("update:modelValue", value);
}
</script>
