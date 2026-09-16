<template>
  <div class="w-full">
    <h2 class="text-2xl font-semibold mb-4">Profile</h2>
    <div class="bg-card p-6 rounded shadow">
      <div class="flex items-center gap-4 mb-6">
        <div
          class="w-16 h-16 bg-primary rounded-full flex items-center justify-center text-white text-xl font-bold"
        >
          initials
        </div>
        <div>
          <div class="text-sm text-muted">{{ user?.email }}</div>
        </div>
      </div>

      <form @submit.prevent="onSave" class="grid grid-cols-1 gap-4">
        <label class="flex flex-col">
          <span class="text-sm font-medium mb-1">Full name</span>
          <ErrorMessage name="name" v-slot="{ message }"
            ><div class="text-sm text-red-600 mt-1">
              {{ message }}
            </div>
          </ErrorMessage>
        </label>

        <label class="flex flex-col">
          <span class="text-sm font-medium mb-1">Email</span>
          <input
            v-model="email"
            type="email"
            class="border rounded px-3 py-2"
          />
          <ErrorMessage name="email" v-slot="{ message }">
            <div class="text-sm text-red-600 mt-1">
              {{ message }}
            </div>
          </ErrorMessage>
        </label>

        <div class="flex gap-3 mt-4">
          <button type="submit" class="px-4 py-2 bg-primary text-white rounded">
            Save
          </button>
          <button type="button" @click="reset" class="px-4 py-2 border rounded">
            Reset
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { useToast } from "primevue/usetoast";
import { useForm, ErrorMessage } from "vee-validate";
import * as yup from "yup";
import { useUsersStore } from "@/stores/users";
import { User } from "@/types";

const usersStore = useUsersStore();
const toast = useToast();

const user = ref<User>(null);
const email = ref<string>("");

const schema = yup.object({
  email: yup.string().email("Invalid email").required("Email is required"),
});

const { handleSubmit } = useForm({
  validationSchema: schema,
  initialValues: { email: email },
});

const onSave = handleSubmit(async (values) => {
  console.log("save");
  try {
    user.value = await usersStore.updateProfile(values.email.value);
    toast.add({
      severity: "success",
      summary: "Saved",
      detail: "Profile saved",
      life: 3000,
    });
  } catch (err) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
    });
  }
});

function reset() {
  console.log("reset");
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
