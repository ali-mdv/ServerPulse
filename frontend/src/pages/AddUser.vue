<template>
  <div class="w-full">
    <h2 class="text-2xl font-semibold mb-4">Add User</h2>
    <div class="bg-card p-6 rounded shadow">
      <form @submit.prevent="onSubmit" class="grid gap-4">
        <label class="flex flex-col">
          <span class="text-sm font-medium mb-1">Email</span>
          <input
            v-model="email"
            type="email"
            class="border rounded px-3 py-2"
          />
          <ErrorMessage name="email" v-slot="{ message }"
            ><div class="text-sm text-red-600 mt-1">
              {{ message }}
            </div></ErrorMessage
          >
        </label>

        <label class="flex flex-col">
          <span class="text-sm font-medium mb-1">Password</span>
          <div class="relative">
            <input
              :type="showPassword ? 'text' : 'password'"
              v-model="password"
              class="border rounded px-3 py-2 w-full pr-10"
              minlength="6"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              :aria-pressed="showPassword"
              aria-label="Toggle password visibility"
              class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted hover:text-primary"
            >
              <Eye v-if="!showPassword" class="w-5 h-5" />
              <EyeOff v-else class="w-5 h-5" />
            </button>
          </div>
          <ErrorMessage name="password" v-slot="{ message }"
            ><div class="text-sm text-red-600 mt-1">
              {{ message }}
            </div></ErrorMessage
          >
        </label>

        <div class="flex gap-3 mt-4">
          <button type="submit" class="px-4 py-2 bg-primary text-white rounded">
            Create
          </button>
          <router-link to="/users" class="px-4 py-2 border rounded">
            Cancel
          </router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { Eye, EyeOff } from "lucide-vue-next";
import { useToast } from "primevue/usetoast";
import { useForm, useField, ErrorMessage } from "vee-validate";
import * as yup from "yup";
import { useUsersStore } from "@/stores/users";

const toast = useToast();

const router = useRouter();

const schema = yup.object({
  email: yup.string().email("Invalid email").required("Email is required"),
  password: yup
    .string()
    .min(6, "Password must be at least 6 characters")
    .required("Password is required"),
});

const { handleSubmit } = useForm({ validationSchema: schema });

const { value: email } = useField<string>("email");
const { value: password } = useField<string>("password");
const showPassword = ref(false);

const userStore = useUsersStore();

const onSubmit = handleSubmit(() => {
  submitForm();
});

async function submitForm() {
  try {
    const user = await userStore.addUser(email.value, password.value);
    toast.add({
      severity: "success",
      summary: "Created",
      detail: `User '${user.email}' created`,
      life: 3000,
    });
    router.push("/users");
  } catch (err) {
    console.log(err);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  }
}
</script>
