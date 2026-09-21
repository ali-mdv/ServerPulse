<template>
  <form @submit.prevent="onSubmit" class="space-y-4" novalidate>
    <div class="space-y-1.5">
      <label for="server-name" class="text-sm font-medium">Name</label>
      <Input
        id="server-name"
        v-model="name"
        placeholder="web-01"
        :invalid="!!nameError"
        :disabled="submitting"
      />
      <ErrorMessage name="name" v-slot="{ message }">
        <p class="text-xs text-critical mt-1">{{ message }}</p>
      </ErrorMessage>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div class="space-y-1.5">
        <label for="server-host" class="text-sm font-medium">Host</label>
        <Input
          id="server-host"
          v-model="host"
          placeholder="192.168.1.10"
          :disabled="submitting"
        />
        <p class="text-xs text-muted-foreground">Optional. The agent can report this later.</p>
      </div>

      <div class="space-y-1.5">
        <label for="server-port" class="text-sm font-medium">Port</label>
        <Input
          id="server-port"
          v-model="port"
          type="number"
          placeholder="0"
          :disabled="submitting"
        />
      </div>
    </div>

    <div class="space-y-1.5">
      <label for="server-description" class="text-sm font-medium">Description</label>
      <Input
        id="server-description"
        v-model="description"
        placeholder="Frontend host"
        :disabled="submitting"
      />
    </div>

    <div class="flex gap-3 pt-2">
      <Button type="submit" variant="primary" :loading="submitting">
        {{ submitting ? "Saving…" : submitLabel }}
      </Button>
      <router-link :to="cancelTo">
        <Button variant="outline" type="button">Cancel</Button>
      </router-link>
    </div>
  </form>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import { useForm, useField, ErrorMessage } from "vee-validate";
import * as yup from "yup";
import { createServer, updateServer, type Server, type CreateServerData, type UpdateServerData } from "@/api/servers";
import Input from "@/components/ui/Input.vue";
import Button from "@/components/ui/Button.vue";

const props = defineProps<{
  server?: Server | null;
}>();

const emit = defineEmits<{
  saved: [server: Server];
}>();

const isEdit = computed(() => !!props.server);
const submitLabel = computed(() => (isEdit.value ? "Save changes" : "Create server"));
const cancelTo = computed(() => "/servers");

const schema = yup.object({
  name: yup.string().required("Name is required").max(64, "Name is too long"),
  host: yup.string().max(255, "Host is too long"),
  port: yup
    .number()
    .transform((value, originalValue) =>
      originalValue === "" ? undefined : value,
    )
    .min(0, "Port must be at least 0")
    .max(65535, "Port must be at most 65535"),
  description: yup.string().max(255, "Description is too long"),
});

const { handleSubmit, errors, resetForm } = useForm({
  validationSchema: schema,
  initialValues: {
    name: props.server?.name ?? "",
    host: props.server?.host ?? "",
    port: props.server?.port ?? undefined,
    description: props.server?.description ?? "",
  },
});

const { value: name } = useField<string>("name");
const { value: host } = useField<string>("host");
const { value: port } = useField<number | undefined>("port");
const { value: description } = useField<string>("description");

watch(
  () => props.server,
  (s) => {
    if (s) {
      resetForm({
        values: {
          name: s.name ?? "",
          host: s.host ?? "",
          port: s.port ?? undefined,
          description: s.description ?? "",
        },
      });
    }
  },
  { immediate: true },
);

const nameError = computed(() => errors.value.name);
const submitting = ref(false);

const onSubmit = handleSubmit(async () => {
  submitting.value = true;
  try {
    const payload: CreateServerData & UpdateServerData = {
      name: name.value,
      host: host.value || undefined,
      port: port.value === undefined ? undefined : Number(port.value),
      description: description.value || undefined,
    };

    const server = isEdit.value
      ? await updateServer(props.server!.id, payload)
      : await createServer(payload);

    emit("saved", server);
  } finally {
    submitting.value = false;
  }
});
</script>
