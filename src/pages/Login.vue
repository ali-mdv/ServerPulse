<template>
  <div class="min-h-[70vh] flex items-center justify-center">
    <div class="w-full max-w-md bg-card p-8 rounded shadow">
      <h2 class="text-2xl font-semibold mb-4">Sign in to {{ appName }}</h2>
      <form @submit.prevent="onSubmit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">Email</label>
          <input
            v-model="email"
            type="email"
            class="w-full border p-2 rounded"
          />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Password</label>
          <div class="relative">
            <input
              :type="showPassword ? 'text' : 'password'"
              v-model="password"
              class="w-full border p-2 rounded pr-10"
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
        </div>
        <div class="flex items-center justify-between">
          <label class="flex items-center gap-2"
            ><input type="checkbox" v-model="remember" /> Remember me</label
          >
          <a href="#" class="text-sm text-primary">Forgot?</a>
        </div>
        <button
          type="submit"
          class="w-full py-2 bg-primary text-primary-foreground rounded"
        >
          Sign in
        </button>
      </form>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "primevue/usetoast";

import { useRootStore } from "@/stores/root";
import { useAuthStore } from "@/stores/auth";
import { Eye, EyeOff } from "lucide-vue-next";

const email = ref("");
const password = ref("");
const showPassword = ref(false);
const remember = ref(false);
const toast = useToast();

const rootStore = useRootStore();
const authStore = useAuthStore();
const router = useRouter();

const appName = computed(() => rootStore.appName);

async function onSubmit() {
  try {
    await authStore.login(email.value, password.value, remember.value);
    toast.add({
      severity: "success",
      detail: "Login successfully",
      life: 3000,
    });
    router.push("/dashboard");
  } catch (err) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  }
}
</script>
