<template>
  <Dialog
    v-model:visible="isOpen"
    :header="title"
    modal
    :style="{ width: '90vw', maxWidth: '420px' }"
    class="p-fluid"
  >
    <p class="text-sm text-muted-foreground">{{ message }}</p>
    <div class="flex justify-end gap-2 mt-6">
      <Button variant="outline" :disabled="loading" @click="isOpen = false">
        Cancel
      </Button>
      <Button
        variant="danger"
        :loading="loading"
        @click="confirm"
      >
        {{ confirmLabel }}
      </Button>
    </div>
  </Dialog>
</template>

<script lang="ts" setup>
import { ref, watch } from "vue";
import Dialog from "primevue/dialog";
import Button from "@/components/ui/Button.vue";

interface Props {
  visible: boolean;
  title?: string;
  message?: string;
  confirmLabel?: string;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  title: "Confirm",
  message: "Are you sure?",
  confirmLabel: "Confirm",
  loading: false,
});

const emit = defineEmits<{
  update: [visible: boolean];
  confirm: [];
}>();

const isOpen = ref(props.visible);

watch(
  () => props.visible,
  (newVal) => {
    isOpen.value = newVal;
  },
);

watch(isOpen, (newVal) => {
  emit("update", newVal);
});

function confirm() {
  emit("confirm");
}
</script>
