<template>
  <div class="min-h-[calc(100vh-2rem)] flex items-center justify-center px-4">
    <div class="w-full max-w-md card card-body">
      <header class="mb-6">
        <h1 class="text-2xl font-semibold">Sign in to {{ appName }}</h1>
        <p class="text-sm text-muted-foreground mt-1">
          Enter your credentials to continue.
        </p>
      </header>

      <form @submit.prevent="onSubmit" class="space-y-4" novalidate>
        <div class="space-y-1.5">
          <label for="login-email" class="text-sm font-medium">Email</label>
          <Input
            id="login-email"
            v-model="email"
            type="email"
            autocomplete="email"
            :invalid="!!emailError"
            :disabled="submitting"
          />
          <ErrorMessage name="email" v-slot="{ message }">
            <p class="text-xs text-critical mt-1">{{ message }}</p>
          </ErrorMessage>
        </div>

        <div class="space-y-1.5">
          <label for="login-password" class="text-sm font-medium">
            Password
          </label>
          <PasswordInput
            id="login-password"
            v-model="password"
            autocomplete="current-password"
            :invalid="!!passwordError"
            :disabled="submitting"
          />
          <ErrorMessage name="password" v-slot="{ message }">
            <p class="text-xs text-critical mt-1">{{ message }}</p>
          </ErrorMessage>
        </div>

        <div class="flex items-center justify-between pt-1">
          <label class="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              v-model="remember"
              class="rounded border-input"
            />
            Remember me
          </label>
          <a
            href="#"
            class="text-sm text-primary hover:underline focus-visible:underline"
            >Forgot?</a
          >
        </div>

        <Button
          type="submit"
          variant="primary"
          :loading="submitting"
          class="w-full"
        >
          {{ submitting ? "Signing in…" : "Sign in" }}
        </Button>
      </form>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "primevue/usetoast";
import { useForm, useField, ErrorMessage } from "vee-validate";
import * as yup from "yup";

import { useRootStore } from "@/stores/root";
import { useAuthStore } from "@/stores/auth";
import Input from "@/components/ui/Input.vue";
import PasswordInput from "@/components/ui/PasswordInput.vue";
import Button from "@/components/ui/Button.vue";

const submitting = ref(false);
const toast = useToast();

const rootStore = useRootStore();
const authStore = useAuthStore();
const router = useRouter();

const schema = yup.object({
  email: yup
    .string()
    .trim()
    .email("Invalid email")
    .required("Email is required"),
  password: yup
    .string()
    .min(6, "Password must be at least 6 characters")
    .required("Password is required"),
});

const { handleSubmit, errors } = useForm({ validationSchema: schema });

const { value: email } = useField<string>("email");
const { value: password } = useField<string>("password");
const remember = ref(false);

const emailError = computed(() => errors.value.email);
const passwordError = computed(() => errors.value.password);

const appName = computed(() => rootStore.appName);

const onSubmit = handleSubmit(async (values) => {
  submitting.value = true;
  try {
    await authStore.login(values.email, values.password, remember.value);
    toast.add({
      severity: "success",
      detail: "Login successfully",
      life: 3000,
    });
    router.push("/dashboard");
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  } finally {
    submitting.value = false;
  }
});
</script>
