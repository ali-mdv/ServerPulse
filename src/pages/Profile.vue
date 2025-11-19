<template>
  <div class="w-full">
    <h2 class="text-2xl font-semibold mb-4">Profile</h2>
    <div class="bg-card p-6 rounded shadow">
      <div class="flex items-center gap-4 mb-6">
        <div class="w-16 h-16 bg-primary rounded-full flex items-center justify-center text-white text-xl font-bold">{{ initials }}</div>
        <div>
          <div class="text-lg font-medium">{{ user.name }}</div>
          <div class="text-sm text-muted">{{ user.email }}</div>
        </div>
      </div>

      <form @submit.prevent="onSave" class="grid grid-cols-1 gap-4">
        <label class="flex flex-col">
          <span class="text-sm font-medium mb-1">Full name</span>
          <input v-model="name" class="border rounded px-3 py-2" />
          <ErrorMessage name="name" v-slot="{ message }"><div class="text-sm text-red-600 mt-1">{{ message }}</div></ErrorMessage>
        </label>

        <label class="flex flex-col">
          <span class="text-sm font-medium mb-1">Email</span>
          <input v-model="email" type="email" class="border rounded px-3 py-2" />
          <ErrorMessage name="email" v-slot="{ message }"><div class="text-sm text-red-600 mt-1">{{ message }}</div></ErrorMessage>
        </label>

        <div class="flex gap-3 mt-4">
          <button type="submit" class="px-4 py-2 bg-primary text-white rounded">Save</button>
          <button type="button" @click="reset" class="px-4 py-2 border rounded">Reset</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { reactive, computed } from 'vue';
import { useToast } from 'primevue/usetoast';
import { useForm, useField, ErrorMessage } from 'vee-validate';
import * as yup from 'yup';

const toast = useToast();
const defaultUser = { name: 'Admin User', email: 'admin@example.com', role: 'Admin' };

const stored = localStorage.getItem('currentUser');
const initial = stored ? JSON.parse(stored) : { ...defaultUser };

const schema = yup.object({ name: yup.string().required('Name is required'), email: yup.string().email('Invalid email').required('Email is required') });
const { handleSubmit } = useForm({ validationSchema: schema, initialValues: { name: initial.name, email: initial.email } });
const { value: name } = useField<string>('name');
const { value: email } = useField<string>('email');

const user = reactive({ name: name.value, email: email.value, role: initial.role });

const initials = computed(() => {
  return (user.name || '')
    .split(' ')
    .map((s: string) => s.charAt(0))
    .slice(0, 2)
    .join('')
    .toUpperCase();
});

const onSave = handleSubmit((values) => {
  user.name = values.name;
  user.email = values.email;
  localStorage.setItem('currentUser', JSON.stringify(user));
  const usersRaw = localStorage.getItem('users');
  const users = usersRaw ? JSON.parse(usersRaw) : [];
  const exists = users.find((u: any) => u.email === user.email);
  if (!exists) users.push({ name: user.name, email: user.email, role: user.role });
  localStorage.setItem('users', JSON.stringify(users));
  toast.add({ severity: 'success', summary: 'Saved', detail: 'Profile saved', life: 3000 });
});

function reset() {
  name.value = defaultUser.name;
  email.value = defaultUser.email;
  user.name = defaultUser.name;
  user.email = defaultUser.email;
}
</script>
