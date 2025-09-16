<template>
  <header class="topbar d-flex align-items-center justify-content-between px-3 py-2">
    <div class="brand fw-bold" role="button" @click="goBrand">GoChat</div>
    <nav class="d-flex gap-3">
      <RouterLink v-if="showRegisterLink" to="/register">Register</RouterLink>
      <!-- Pe Register: arătăm doar Login -->
      <RouterLink v-if="showLoginLink" to="/login">Login</RouterLink>
    </nav>
  </header>

  <RouterView />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const isAuth = computed(() => !!localStorage.getItem('auth_token'))

const showRegisterLink = computed(() => !isAuth.value && route.name !== 'Register')
const showLoginLink = computed(() => !isAuth.value && route.name !== 'Login')

function goBrand() {
  router.push(isAuth.value ? '/home' : '/login')
}
</script>

<style scoped>
.topbar { background: rgba(0,0,0,.05); backdrop-filter: blur(4px); }
.brand { cursor: pointer; }
</style>