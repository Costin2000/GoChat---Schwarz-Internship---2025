<template>
  <!-- The max-width is now dynamically controlled by the 'maxWidth' prop -->
  <div class="auth-card shadow-lg border-0 rounded-4" :style="{ maxWidth: maxWidth }">
    <h2 v-if="title" class="h4 mb-3 text-center">{{ title }}</h2>
    <slot />
  </div>
</template>

<script setup lang="ts">
// We use withDefaults to set a default max-width for the card.
// This will apply to Login, Register, etc., unless specified otherwise.
withDefaults(defineProps<{ 
  title?: string,
  maxWidth?: string 
}>(), {
  maxWidth: '480px' // A good default width for forms
})
</script>

<style scoped>
.auth-card {
  width: 100%;
  padding: 24px;
  background: rgba(255,255,255,0.9);
}
</style>

