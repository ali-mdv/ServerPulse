<template>
  <Dialog
    v-model:visible="isOpen"
    :header="`${serviceName} Logs`"
    modal
    :style="{ width: '90vw', maxWidth: '900px' }"
    class="p-fluid"
    @show="scrollToBottom"
  >
    <div class="space-y-4">
      <div v-if="loading" class="flex items-center justify-center p-8">
        <div class="text-center">
          <div class="mb-2">Loading logs...</div>
          <div class="text-sm text-muted">Please wait</div>
        </div>
      </div>

      <div
        v-else
        ref="logsContainer"
        class="bg-gray-900 text-gray-100 p-4 rounded text-sm overflow-auto max-h-96 border border-gray-700"
        style="
          font-family: 'Menlo', 'Monaco', 'Courier New', 'Courier', monospace;
          white-space: pre-wrap;
          word-break: break-word;
        "
      >
        <div
          v-for="(line, index) in logLines"
          :key="index"
          class="py-0"
          v-html="line"
        ></div>
        <div v-if="logLines.length === 0" class="text-gray-500">
          No logs available
        </div>
      </div>

      <div class="flex gap-2 justify-end">
        <button
          @click="refreshLogs"
          :disabled="loading"
          class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {{ loading ? "Refreshing..." : "Refresh" }}
        </button>
        <button
          @click="copyToClipboard"
          class="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 transition"
        >
          Copy
        </button>
        <button
          @click="isOpen = false"
          class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600 transition"
        >
          Close
        </button>
      </div>
    </div>
  </Dialog>
</template>

<script lang="ts" setup>
import { ref, computed, watch, nextTick } from "vue";
import Dialog from "primevue/dialog";
import { useToast } from "primevue/usetoast";
import { ansiToHtml } from "@/lib/ansi-to-html";
import { LogsModalProps } from "@/types";

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

const logLines = computed(() => {
  return props.content.split("\n").map((line) => ansiToHtml(line));
});

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
    // When loading finishes, scroll to bottom after content renders
    if (!newLoading) {
      scrollToBottomAfterRender();
    }
  },
);

watch(
  () => props.content,
  () => {
    if (!props.loading) {
      scrollToBottomAfterRender();
    }
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
  } catch (err) {
    try {
      fallbackCopy(props.content);
      toast.add({
        severity: "success",
        summary: "Success",
        detail: "Logs copied to clipboard",
        life: 2000,
      });
    } catch (err) {
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
