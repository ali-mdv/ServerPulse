<template>
  <div class="space-y-6 max-w-2xl">
    <PageHeader
      title="Settings"
      subtitle="Manage your preferences and connections."
    />

    <div v-if="store.loading && !store.settings" class="flex justify-center py-10">
      <Spinner size="lg" />
    </div>

    <form v-else @submit.prevent="onSave" novalidate>
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
                v-model="appearance"
                name="theme"
              />
              {{ opt.label }}
            </label>
          </fieldset>
        </Card>

        <Card>
          <h3 class="font-semibold mb-3">History</h3>
          <div class="space-y-4">
            <div
              class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
            >
              <div>
                <div class="text-sm font-medium">Poll interval</div>
                <div class="text-xs text-muted-foreground">
                  How often PM2, Docker and system metrics are snapshotted.
                </div>
              </div>
              <Select
                v-model.number="pollSeconds"
                aria-label="History poll interval"
                class="w-auto"
              >
                <option v-for="s in pollOptions" :key="s" :value="s">
                  {{ durationLabel(s) }}
                </option>
              </Select>
            </div>

            <div
              class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
            >
              <div>
                <div class="text-sm font-medium">Retention</div>
                <div class="text-xs text-muted-foreground">
                  How long snapshots are kept before they expire.
                </div>
              </div>
              <Select
                v-model.number="retentionDays"
                aria-label="History retention"
                class="w-auto"
              >
                <option v-for="d in retentionOptions" :key="d" :value="d">
                  {{ d }} {{ d === 1 ? "day" : "days" }}
                </option>
              </Select>
            </div>
          </div>
          <p v-if="store.error" class="text-xs text-critical mt-3">
            {{ store.error }}
          </p>
        </Card>

        <div class="flex gap-3">
          <Button type="submit" variant="primary" :loading="store.saving">
            {{ store.saving ? "Saving…" : "Save changes" }}
          </Button>
          <Button
            type="button"
            variant="outline"
            :disabled="store.saving"
            @click="reset"
          >
            Reset
          </Button>
        </div>
      </div>
    </form>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted } from "vue";
import { useToast } from "primevue/usetoast";
import { useForm, useField } from "vee-validate";
import * as yup from "yup";
import { useTheme, type Theme } from "@/composables";
import { useSettingsStore } from "@/stores/settings";
import Card from "@/components/Card.vue";
import Button from "@/components/ui/Button.vue";
import Select from "@/components/ui/Select.vue";
import Spinner from "@/components/ui/Spinner.vue";
import PageHeader from "@/components/ui/PageHeader.vue";

const toast = useToast();
const store = useSettingsStore();
const { setTheme } = useTheme();

const themeOptions = [
  { value: "system", label: "System" },
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
] as const;

const POLL_PRESETS = [10, 30, 60, 300];
const RETENTION_PRESETS = [1, 7, 30, 90];

const schema = yup.object({
  appearance: yup
    .string()
    .oneOf(["system", "light", "dark"], "Pick system, light or dark")
    .required(),
  historyPollIntervalSeconds: yup
    .number()
    .min(5, "Must be at least 5 seconds")
    .max(86400, "Must be at most 24 hours")
    .required(),
  historyRetentionDays: yup
    .number()
    .min(1, "Must be at least 1 day")
    .max(365, "Must be at most 365 days")
    .required(),
});

const { handleSubmit, setValues } = useForm({
  validationSchema: schema,
  initialValues: {
    appearance: "system" as Theme,
    historyPollIntervalSeconds: 30,
    historyRetentionDays: 7,
  },
});

const { value: appearance } = useField<Theme>("appearance");
const { value: pollSeconds } = useField<number>("historyPollIntervalSeconds");
const { value: retentionDays } = useField<number>("historyRetentionDays");

const pollOptions = computed(() => {
  const options = [...POLL_PRESETS];
  if (pollSeconds.value && !options.includes(pollSeconds.value)) {
    options.push(pollSeconds.value);
  }
  return options.sort((a, b) => a - b);
});

const retentionOptions = computed(() => {
  const options = [...RETENTION_PRESETS];
  if (retentionDays.value && !options.includes(retentionDays.value)) {
    options.push(retentionDays.value);
  }
  return options.sort((a, b) => a - b);
});

function durationLabel(seconds: number): string {
  if (seconds >= 60 && seconds % 60 === 0) {
    const minutes = seconds / 60;
    return `${minutes} ${minutes === 1 ? "minute" : "minutes"}`;
  }
  return `${seconds} seconds`;
}

function applySettingsToForm() {
  const settings = store.settings;
  if (!settings) return;
  setValues({
    appearance: settings.appearance,
    historyPollIntervalSeconds: settings.historyPollIntervalSeconds,
    historyRetentionDays: Math.max(1, Math.round(settings.historyRetentionSeconds / 86400)),
  });
  setTheme(settings.appearance);
}

onMounted(async () => {
  await store.load();
  applySettingsToForm();
});

const onSave = handleSubmit(async (values) => {
  try {
    await store.save({
      appearance: values.appearance as Theme,
      historyPollIntervalSeconds: values.historyPollIntervalSeconds,
      historyRetentionSeconds: values.historyRetentionDays * 86400,
    });
    // Apply immediately; otherwise the new theme only lands on reload.
    setTheme(values.appearance as Theme);
    toast.add({
      severity: "success",
      summary: "Saved",
      detail: "Settings saved",
      life: 3000,
    });
  } catch {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: store.error ?? "Failed to save settings",
      life: 4000,
    });
  }
});

async function reset() {
  await store.load();
  applySettingsToForm();
}
</script>