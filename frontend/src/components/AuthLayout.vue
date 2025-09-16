<template>
  <div class="auth-wrap">
    <slot />
  </div>
</template>

<style scoped>
.auth-wrap {
  min-height: 100vh;         
  width: 100%;             
  display: grid;            
  place-items: center;
  padding: 24px;            
}
</style>