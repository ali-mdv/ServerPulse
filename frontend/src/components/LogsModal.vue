<template>
  <Dialog
    v-model:visible="isOpen"
    :header="`${serviceName} — Logs`"
    modal
    :style="{ width: '90vw', maxWidth: '900px' }"
    class="p-fluid"
    @show="scrollToBottom"
  >
    <div class="space-y-4">
      <div
        v-if="loading"
        class="flex items-center justify-center p-8 text-sm text-muted-foreground"
        role="status"
        aria-live="polite"
      >
        <Spinner size="md" class="mr-2" />
        Loading logs…
      </div>

      <div
        v-else
        ref="logsContainer"
        class="rounded-md border border-border bg-zinc-950 text-zinc-100 p-4 text-sm overflow-auto max-h-96 font-mono whitespace-pre-wrap break-words"
      >
        <div
          v-for="(line, index) in logLines"
          :key="index"
          class="py-0"
          v-html="line"
        ></div>
        <div v-if="logLines.length === 0" class="text-zinc-500">
          No logs available
        </div>
      </div>

      <div class="flex flex-wrap gap-2 justify-end">
        <Button
          variant="secondary"
          :disabled="loading || logLines.length === 0"
          @click="copyToClipboard"
        >
          <template #icon-left>
            <Copy class="w-4 h-4" aria-hidden="true" />
          </template>
          Copy
        </Button>
        <Button variant="info" :loading="loading" @click="refreshLogs">
          <template v-if="!loading" #icon-left>
            <RotateCw class="w-4 h-4" aria-hidden="true" />
          </template>
          {{ loading ? "Refreshing…" : "Refresh" }}
        </Button>
        <Button variant="outline" @click="isOpen = false">Close</Button>
      </div>
    </div>
  </Dialog>
</template>

<script lang="ts" setup>
import { ref, computed, watch, nextTick } from "vue";
import Dialog from "primevue/dialog";
import { useToast } from "primevue/usetoast";
import { Copy, RotateCw } from "lucide-vue-next";
import { ansiToHtml } from "@/lib/ansi-to-html";
import { LogsModalProps } from "@/types";
import Button from "@/components/ui/Button.vue";
import Spinner from "@/components/ui/Spinner.vue";

const props = withDefaults(defineProps<LogsModalProps>(), {
  visible: false,
  loading: false,
});

const emit = defineEmits<{
  update: [visible: boolean];
  refresh: [];
}>();

const isOpen = ref(props.visible);
const toast = useToast();
const logsContainer = ref<HTMLDivElement | null>(null);

function scrollToBottom() {
  nextTick(() => {
    if (logsContainer.value) {
      logsContainer.value.scrollTop = logsContainer.value.scrollHeight;
    }
  });
}

const logLines = computed(() =>
  props.content.split("\n").map((line) => ansiToHtml(line)),
);

watch(
  () => props.visible,
  (newVal) => {
    isOpen.value = newVal;
  },
);

watch(isOpen, (newVal) => {
  emit("update", newVal);
});

watch(
  () => props.loading,
  (newLoading) => {
    if (!newLoading) scrollToBottomAfterRender();
  },
);

watch(
  () => props.content,
  () => {
    if (!props.loading) scrollToBottomAfterRender();
  },
  { flush: "post" },
);

function scrollToBottomAfterRender() {
  requestAnimationFrame(() => {
    if (logsContainer.value) {
      logsContainer.value.scrollTop = logsContainer.value.scrollHeight;
    }
  });
}

function refreshLogs() {
  emit("refresh");
  toast.add({
    severity: "success",
    summary: "Success",
    detail: "Logs refreshed",
    life: 2000,
  });
}

function fallbackCopy(text: string) {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  try {
    document.execCommand("copy");
  } finally {
    document.body.removeChild(textarea);
  }
}

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(props.content);
    toast.add({
      severity: "success",
      summary: "Success",
      detail: "Logs copied to clipboard",
      life: 2000,
    });
  } catch {
    try {
      fallbackCopy(props.content);
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Logs copied to clipboard",
        life: 2000,
      });
    } catch {
      toast.add({
        severity: "error",
        summary: "Error",
        detail: "Failed to copy logs",
        life: 2000,
      });
    }
  }
}
</script>
