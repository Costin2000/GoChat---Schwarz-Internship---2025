<template>
  <div class="auth-wrap" :style="appBg">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const appBg = computed(() => ({
  background: 'linear-gradient(135deg, rgba(15,81,50,1) 0%, rgba(25,135,84,1) 60%, rgba(209,231,221,1) 100%)'
}))
</script>

<style scoped>
.auth-wrap {
  min-height: calc(100vh - 40px);        
  width: 100%;            
  display: grid;          
  place-items: center;
  padding: 24px;          
}
</style>
