<template>
  <div class="space-y-6 max-w-xl">
    <PageHeader
      title="Add User"
      subtitle="Create a new account for someone on your team."
    />

    <Card>
      <form @submit.prevent="onSubmit" class="space-y-4" novalidate>
        <div class="space-y-1.5">
          <label for="add-email" class="text-sm font-medium">Email</label>
          <Input
            id="add-email"
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
          <label for="add-password" class="text-sm font-medium">Password</label>
          <PasswordInput
            id="add-password"
            v-model="password"
            autocomplete="new-password"
            :invalid="!!passwordError"
            :disabled="submitting"
          />
          <ErrorMessage name="password" v-slot="{ message }">
            <p class="text-xs text-critical mt-1">{{ message }}</p>
          </ErrorMessage>
        </div>

        <div class="flex gap-3 pt-2">
          <Button type="submit" variant="primary" :loading="submitting">
            {{ submitting ? "Creating…" : "Create user" }}
          </Button>
          <router-link to="/users">
            <Button variant="outline" type="button">Cancel</Button>
          </router-link>
        </div>
      </form>
    </Card>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "primevue/usetoast";
import { useForm, useField, ErrorMessage } from "vee-validate";
import * as yup from "yup";
import { useUsersStore } from "@/stores/users";
import Card from "@/components/Card.vue";
import Input from "@/components/ui/Input.vue";
import PasswordInput from "@/components/ui/PasswordInput.vue";
import Button from "@/components/ui/Button.vue";
import PageHeader from "@/components/ui/PageHeader.vue";

const toast = useToast();
const router = useRouter();

const schema = yup.object({
  email: yup.string().email("Invalid email").required("Email is required"),
  password: yup
    .string()
    .min(6, "Password must be at least 6 characters")
    .required("Password is required"),
});

const { handleSubmit, errors } = useForm({ validationSchema: schema });

const { value: email } = useField<string>("email");
const { value: password } = useField<string>("password");
const submitting = ref(false);

const emailError = computed(() => errors.value.email);
const passwordError = computed(() => errors.value.password);

const userStore = useUsersStore();

const onSubmit = handleSubmit(() => {
  submitForm();
});

async function submitForm() {
  submitting.value = true;
  try {
    const user = await userStore.addUser(email.value, password.value);
    toast.add({
      severity: "success",
      summary: "Created",
      detail: `User '${user.email}' created`,
      life: 3000,
    });
    router.push("/users");
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
}
</script>
