<template>
  <div class="space-y-6 max-w-xl">
    <PageHeader
      title="Edit Server"
      :subtitle="server?.name ?? 'Update server details'"
    />

    <Card>
      <div v-if="loading" class="space-y-3">
        <Skeleton height="2.5rem" />
        <Skeleton height="2.5rem" />
        <Skeleton height="2.5rem" />
      </div>
      <ServerForm v-else :server="server" @saved="onSaved" />
    </Card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useToast } from "primevue/usetoast";
import ServerForm from "@/components/ServerForm.vue";
import Card from "@/components/Card.vue";
import Skeleton from "@/components/ui/Skeleton.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import { fetchServer, type Server } from "@/api/servers";

const route = useRoute();
const router = useRouter();
const toast = useToast();

const id = route.params.id as string;
const server = ref<Server | null>(null);
const loading = ref(true);

onMounted(async () => {
  try {
    const data = await fetchServer(id);
    if (!data) {
      toast.add({
        severity: "error",
        summary: "Not found",
        detail: "Server not found.",
        life: 3000,
      });
      router.push("/servers");
      return;
    }
    server.value = data;
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
    router.push("/servers");
  } finally {
    loading.value = false;
  }
});

function onSaved(saved: Server) {
  toast.add({
    severity: "success",
    summary: "Saved",
    detail: `Server '${saved.name}' updated.`,
    life: 3000,
  });
  router.push("/servers");
}
</script>
