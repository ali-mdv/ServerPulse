<template>
  <div :class="['card', padded && 'card-body']">
    <slot />
  </div>
</template>

<script lang="ts" setup>
withDefaults(defineProps<{ padded?: boolean }>(), { padded: true });
</script>
