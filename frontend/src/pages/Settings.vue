<template>
  <div class="space-y-6 max-w-2xl">
    <PageHeader
      title="Settings"
      subtitle="Manage your preferences and connections."
    />

    <form @submit.prevent="onSave" novalidate>
      <div class="space-y-6">
        <Card>
          <h3 class="font-semibold mb-3">Appearance</h3>
          <fieldset class="space-y-2">
            <legend class="sr-only">Theme preference</legend>
            <label
              v-for="opt in themeOptions"
              :key="opt.value"
              class="flex items-center gap-2 text-sm cursor-pointer"
            >
              <input
                type="radio"
                :value="opt.value"
                v-model="themeValue"
                name="theme"
              />
              {{ opt.label }}
            </label>
          </fieldset>
        </Card>

        <Card>
          <h3 class="font-semibold mb-3">Preferences</h3>
          <div
            class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
          >
            <div>
              <div class="text-sm font-medium">Metrics refresh interval</div>
              <div class="text-xs text-muted-foreground">
                How frequently metrics auto-refresh.
              </div>
            </div>
            <Select
              v-model.number="refreshInterval"
              aria-label="Refresh interval"
              class="w-auto"
            >
              <option :value="5">5 seconds</option>
              <option :value="10">10 seconds</option>
              <option :value="30">30 seconds</option>
            </Select>
          </div>
          <ErrorMessage name="refreshInterval" v-slot="{ message }">
            <p class="text-xs text-critical mt-2">{{ message }}</p>
          </ErrorMessage>
        </Card>

        <Card>
          <h3 class="font-semibold mb-3">Connections</h3>
          <div class="space-y-4">
            <div class="space-y-1.5">
              <label for="ws-endpoint" class="text-sm font-medium">
                WebSocket endpoint
              </label>
              <Input
                id="ws-endpoint"
                v-model="wsEndpoint"
                type="url"
                :invalid="!!wsEndpointError"
              />
              <ErrorMessage name="wsEndpoint" v-slot="{ message }">
                <p class="text-xs text-critical mt-1">{{ message }}</p>
              </ErrorMessage>
            </div>
            <div class="space-y-1.5">
              <label for="api-base" class="text-sm font-medium">
                REST API base
              </label>
              <Input
                id="api-base"
                v-model="apiBase"
                type="url"
                :invalid="!!apiBaseError"
              />
              <ErrorMessage name="apiBase" v-slot="{ message }">
                <p class="text-xs text-critical mt-1">{{ message }}</p>
              </ErrorMessage>
            </div>
          </div>
        </Card>

        <div class="flex gap-3">
          <Button type="submit" variant="primary" :loading="saving">
            {{ saving ? "Saving…" : "Save changes" }}
          </Button>
          <Button type="button" variant="outline" @click="reset">
            Reset
          </Button>
        </div>
      </div>
    </form>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { useToast } from "primevue/usetoast";
import { useForm, useField, ErrorMessage } from "vee-validate";
import * as yup from "yup";
import { useTheme, type Theme } from "@/composables";
import Card from "@/components/Card.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import PageHeader from "@/components/ui/PageHeader.vue";

const toast = useToast();
const { theme: themeValue, setTheme } = useTheme();
const saving = ref(false);

const themeOptions = [
  { value: "system", label: "System" },
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
] as const;

const schema = yup.object({
  refreshInterval: yup
    .number()
    .oneOf([5, 10, 30], "Pick 5, 10 or 30 seconds")
    .required(),
  wsEndpoint: yup
    .string()
    .trim()
    .url("Must be a valid URL")
    .required("WebSocket endpoint is required"),
  apiBase: yup
    .string()
    .trim()
    .url("Must be a valid URL")
    .required("REST API base is required"),
});

const { handleSubmit, errors, setValues } = useForm({
  validationSchema: schema,
  initialValues: {
    refreshInterval: 10,
    wsEndpoint: "wss://example.com/ws",
    apiBase: "https://api.example.com",
  },
});

const { value: refreshInterval } = useField<number>("refreshInterval");
const { value: wsEndpoint } = useField<string>("wsEndpoint");
const { value: apiBase } = useField<string>("apiBase");

const wsEndpointError = computed(() => errors.value.wsEndpoint);
const apiBaseError = computed(() => errors.value.apiBase);

function loadFromStorage() {
  if (typeof localStorage === "undefined") return;
  try {
    const raw = localStorage.getItem("settings");
    if (!raw) return;
    const s = JSON.parse(raw);
    setValues({
      refreshInterval: s.refreshInterval ?? 10,
      wsEndpoint: s.wsEndpoint ?? "wss://example.com/ws",
      apiBase: s.apiBase ?? "https://api.example.com",
    });
  } catch {
    /* ignore */
  }
}

loadFromStorage();

const onSave = handleSubmit((values) => {
  saving.value = true;
  const payload = {
    theme: themeValue.value as Theme,
    refreshInterval: values.refreshInterval,
    wsEndpoint: values.wsEndpoint,
    apiBase: values.apiBase,
  };
  localStorage.setItem("settings", JSON.stringify(payload));
  setTheme(payload.theme);
  window.dispatchEvent(new CustomEvent("settings-updated", { detail: payload }));
  toast.add({
    severity: "success",
    summary: "Saved",
    detail: "Settings saved",
    life: 3000,
  });
  saving.value = false;
});

function reset() {
  themeValue.value = "system";
  setValues({
    refreshInterval: 10,
    wsEndpoint: "wss://example.com/ws",
    apiBase: "https://api.example.com",
  });
}
</script>
