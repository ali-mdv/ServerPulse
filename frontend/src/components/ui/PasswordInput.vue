<template>
  <div class="relative">
    <input
      :type="visible ? 'text' : 'password'"
      :class="cn('input pr-10', $attrs.class as string)"
      :value="modelValue"
      :aria-invalid="invalid || undefined"
      autocomplete="current-password"
      @input="onInput"
      v-bind="remainingAttrs"
    />
    <button
      type="button"
      @click="visible = !visible"
      :aria-pressed="visible"
      :aria-label="visible ? 'Hide password' : 'Show password'"
      class="absolute right-1 top-1/2 -translate-y-1/2 btn btn-ghost btn-icon"
    >
      <Eye v-if="!visible" class="w-4 h-4" aria-hidden="true" />
      <EyeOff v-else class="w-4 h-4" aria-hidden="true" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useAttrs } from "vue";
import { Eye, EyeOff } from "lucide-vue-next";
import { cn } from "@/lib/utils";

interface Props {
  modelValue?: string;
  invalid?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const attrs = useAttrs();
const visible = ref(false);

const remainingAttrs = computed(() => {
  const { class: _c, type: _t, ...rest } = attrs as Record<string, unknown>;
  return rest;
});

function onInput(e: Event) {
  emit("update:modelValue", (e.target as HTMLInputElement).value);
}
</script>
