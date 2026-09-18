<template>
  <input
    :type="type"
    :class="cn('input', $attrs.class as string)"
    :value="modelValue"
    :aria-invalid="invalid || undefined"
    @input="onInput"
    v-bind="remainingAttrs"
  />
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { cn } from "@/lib/utils";

interface Props {
  modelValue?: string | number;
  type?: string;
  invalid?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const attrs = useAttrs();

const remainingAttrs = computed(() => {
  const { class: _c, ...rest } = attrs as Record<string, unknown>;
  return rest;
});

function onInput(e: Event) {
  emit("update:modelValue", (e.target as HTMLInputElement).value);
}
</script>
