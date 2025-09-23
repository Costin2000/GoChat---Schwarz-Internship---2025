<template>
  <AuthLayout>
    <AuthCard title="Home">
      <p class="mb-3">You are logged in.</p>
      <button class="btn btn-outline-secondary w-100" @click="logout">Log out</button>
    </AuthCard>
  </AuthLayout>
</template>

<script setup lang="ts">
import AuthLayout from '@/components/AuthLayout.vue'
import AuthCard from '@/components/AuthCard.vue'
import { useRouter } from 'vue-router'

const router = useRouter()

function logout() {
  localStorage.removeItem('auth_token')
  localStorage.removeItem('user_id')
  router.replace('/login')
}

function goFriends() {
  router.push('/friends')
}
</script>