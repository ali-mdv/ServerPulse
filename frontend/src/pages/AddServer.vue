<template>
  <div class="space-y-6 max-w-xl">
    <PageHeader
      title="Add Server"
      subtitle="Register a new agent-monitored server."
    />

    <Card>
      <ServerForm @saved="onSaved" />
    </Card>

    <Card v-if="createdServer">
      <h3 class="font-semibold mb-2">API Key</h3>
      <p class="text-sm text-muted-foreground mb-3">
        Generate an API key for <strong>{{ createdServer.name }}</strong> so the remote agent can authenticate.
      </p>
      <div class="flex gap-2">
        <Button
          variant="primary"
          :loading="generatingKey"
          @click="generateKey"
        >
          Generate API key
        </Button>
      </div>
      <div v-if="apiKey" class="mt-4 p-3 bg-muted rounded-md">
        <div class="text-xs text-muted-foreground mb-1">Copy this key now. It will not be shown again.</div>
        <code class="text-sm break-all">{{ apiKey }}</code>
      </div>
    </Card>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useToast } from "primevue/usetoast";
import ServerForm from "@/components/ServerForm.vue";
import Card from "@/components/Card.vue";
import Button from "@/components/ui/Button.vue";
import PageHeader from "@/components/ui/PageHeader.vue";
import { generateServerApiKey, type Server } from "@/api/servers";

const router = useRouter();
const toast = useToast();

const createdServer = ref<Server | null>(null);
const apiKey = ref("");
const generatingKey = ref(false);

function onSaved(server: Server) {
  createdServer.value = server;
  toast.add({
    severity: "success",
    summary: "Created",
    detail: `Server '${server.name}' registered.`,
    life: 3000,
  });
}

async function generateKey() {
  if (!createdServer.value) return;
  generatingKey.value = true;
  try {
    apiKey.value = await generateServerApiKey(createdServer.value.id);
  } catch (err: any) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  } finally {
    generatingKey.value = false;
  }
}
</script>
