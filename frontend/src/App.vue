<template>
  <header class="topbar d-flex align-items-center justify-content-between px-3 py-2">
    <div class="d-flex align-items-center gap-3">
      <div class="brand fw-bold" role="button" @click="goBrand">GoChat</div>
      <RouterLink v-if="token" to='/conversations' class="nav-link">Conversations</RouterLink>
      <RouterLink v-if="token" to="/find-friends" class="nav-link">Find Friends</RouterLink>
      <RouterLink v-if="token" to="/friend-requests" class="nav-link">Friend Requests</RouterLink>
      <RouterLink v-if="token" to="/friends" class="nav-link">Friends</RouterLink>
    </div>
    <nav class="d-flex gap-3 align-items-center">
      <!-- A logout button that only appears when logged in -->
      <button v-if="token" @click="logout" class="logout-button">Logout</button>
    </nav>
  </header>

  <RouterView />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getToken, clearAuth } from '@/lib/auth'

const route = useRoute()
const router = useRouter()

const token = ref(getToken())

watch(() => route.fullPath, () => {
  token.value = getToken()
}, {
  immediate: true 
})

function goBrand() {
  router.push(token.value ? '/home' : '/login')
}

function logout() {
  clearAuth()
  token.value = null 
  router.replace('/login')
}
</script>

<style scoped>
.topbar { background: rgba(0,0,0,.05); backdrop-filter: blur(4px); }
.brand { cursor: pointer; color: aquamarine;}
.nav-link {
  color: rgb(191, 234, 233);
  text-decoration: none;
  transition: color 0.2s ease-in-out;
}
.nav-link.router-link-active,
.nav-link.router-link-exact-active { color: rgb(191, 234, 233); text-decoration: underline }
.nav-link:hover { color: rgb(171, 206, 205); }

.logout-button {
  background: none;
  border: none;
  color: rgb(191, 234, 233);
  cursor: pointer;
  padding: 0;
  font-size: 1rem;
}
.logout-button:hover {
  text-decoration: underline;
}
</style>

