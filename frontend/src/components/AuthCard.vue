<template>
  <div class="auth-card shadow-lg">
    <h2 class="h4 mb-3 text-center">{{ title }}</h2>
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 420px;
  padding: 24px;
  border-radius: 16px;

  background:
    linear-gradient(180deg, rgba(255,255,255,0.18) 0%, rgba(255,255,255,0.10) 100%),
    rgba(17, 84, 63, 0.18);   
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border: 1px solid rgba(255,255,255,0.28);
  box-shadow: 0 10px 30px rgba(0,0,0,0.25);
}

.auth-card :deep(.form-control) {
  background: rgba(255,255,255,0.7);
  border: 1px solid rgba(0,0,0,0.1);
}
.auth-card :deep(.form-control:focus) {
  background: rgba(255,255,255,0.85);
  border-color: #2fbf7a;
  box-shadow: 0 0 0 0.2rem rgba(47,191,122,0.25);
}
</style>