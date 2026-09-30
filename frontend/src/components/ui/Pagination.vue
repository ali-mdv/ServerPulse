<template>
  <nav
    v-if="totalPages > 1"
    :class="cn('flex items-center justify-between gap-2', $attrs.class as string)"
    :aria-label="ariaLabel"
  >
    <Button
      variant="outline"
      size="sm"
      :disabled="page <= 1"
      @click="go(page - 1)"
    >
      <ChevronLeft class="w-4 h-4" aria-hidden="true" />
      Prev
    </Button>
    <span class="text-xs text-muted-foreground">
      Page {{ page }} of {{ totalPages }}
    </span>
    <Button
      variant="outline"
      size="sm"
      :disabled="page >= totalPages"
      @click="go(page + 1)"
    >
      Next
      <ChevronRight class="w-4 h-4" aria-hidden="true" />
    </Button>
  </nav>
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { cn } from "@/lib/utils";
import Button from "./Button.vue";

interface Props {
  page: number;
  total: number;
  pageSize?: number;
  ariaLabel?: string;
}

const props = withDefaults(defineProps<Props>(), {
  pageSize: 10,
  ariaLabel: "Pagination",
});

const emit = defineEmits<{
  "update:page": [value: number];
}>();

defineOptions({ inheritAttrs: false });
const attrs = useAttrs();

const totalPages = computed(() =>
  Math.max(1, Math.ceil(props.total / props.pageSize)),
);

function go(next: number) {
  const clamped = Math.min(Math.max(1, next), totalPages.value);
  if (clamped !== props.page) emit("update:page", clamped);
}
</script>
