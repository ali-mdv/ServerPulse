<template>
  <div class="space-y-6 max-w-xl">
    <PageHeader title="Profile" subtitle="Manage your account details." />

    <Card>
      <div class="flex items-center gap-4 mb-6">
        <span
          class="w-16 h-16 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xl font-bold"
          aria-hidden="true"
        >
          {{ initials }}
        </span>
        <div class="min-w-0">
          <div class="text-sm text-muted-foreground truncate">
            {{ user?.email }}
          </div>
        </div>
      </div>

      <form @submit.prevent="onSave" class="space-y-4" novalidate>
        <div class="space-y-1.5">
          <label for="profile-name" class="text-sm font-medium">
            Full name
          </label>
          <Input
            id="profile-name"
            v-model="name"
            type="text"
            autocomplete="name"
            :invalid="!!nameError"
            :disabled="saving"
          />
          <ErrorMessage name="name" v-slot="{ message }">
            <p class="text-xs text-critical mt-1">{{ message }}</p>
          </ErrorMessage>
        </div>

        <div class="space-y-1.5">
          <label for="profile-email" class="text-sm font-medium">Email</label>
          <Input
            id="profile-email"
            v-model="email"
            type="email"
            autocomplete="email"
            :invalid="!!emailError"
            :disabled="saving"
          />
          <ErrorMessage name="email" v-slot="{ message }">
            <p class="text-xs text-critical mt-1">{{ message }}</p>
          </ErrorMessage>
        </div>

        <div class="flex gap-3 pt-2">
          <Button type="submit" variant="primary" :loading="saving">
            {{ saving ? "Saving…" : "Save changes" }}
          </Button>
          <Button
            type="button"
            variant="outline"
            :disabled="saving"
            @click="reset"
          >
            Reset
          </Button>
        </div>
      </form>
    </Card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import { useToast } from "primevue/usetoast";
import { useForm, useField, ErrorMessage } from "vee-validate";
import * as yup from "yup";
import { useUsersStore } from "@/stores/users";
import { User } from "@/types";
import Card from "@/components/Card.vue";
import Input from "@/components/ui/Input.vue";
import Button from "@/components/ui/Button.vue";
import PageHeader from "@/components/ui/PageHeader.vue";

const usersStore = useUsersStore();
const toast = useToast();

const user = ref<User | null>(null);
const saving = ref(false);

const schema = yup.object({
  name: yup.string().trim().min(2, "Name must be at least 2 characters"),
  email: yup.string().email("Invalid email").required("Email is required"),
});

const { handleSubmit, resetForm, errors } = useForm({
  validationSchema: schema,
  initialValues: { name: "", email: "" },
});

const { value: name } = useField<string>("name");
const { value: email } = useField<string>("email");

const nameError = computed(() => errors.value.name);
const emailError = computed(() => errors.value.email);

const initials = computed(() => {
  const n = (name.value || user.value?.email || "?").trim();
  return n.charAt(0).toUpperCase();
});

const onSave = handleSubmit(async (values) => {
  saving.value = true;
  try {
    user.value = await usersStore.updateProfile(values.email);
    toast.add({
      severity: "success",
      summary: "Saved",
      detail: "Profile saved",
      life: 3000,
    });
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
    });
  } finally {
    saving.value = false;
  }
});

function reset() {
  resetForm({ values: { name: name.value, email: email.value } });
}

onMounted(async () => {
  try {
    user.value = await usersStore.getProfile();
    email.value = user.value.email;
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
    });
  }
});
</script>
