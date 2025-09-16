<template>
  <AuthLayout>
    <AuthCard title="Log in">
      <form @submit.prevent="onSubmit" novalidate>
        <div class="mb-3">
          <label class="form-label">Email</label>
          <input v-model.trim="email" type="email" class="form-control" required />
        </div>

        <div class="mb-2">
          <label class="form-label">Password</label>
          <input v-model="password" type="password" class="form-control" required minlength="6" />
        </div>

        <p v-if="error" class="text-danger small mb-2">{{ error }}</p>

        <!-- added mt-3 for spacing above the button -->
        <button class="btn btn-success w-100 mt-3" :disabled="loading">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2" /> Log in
        </button>
      </form>

      <hr class="my-3" />
      <div class="text-center">
        <span class="me-1">Don’t have an account?</span>
        <RouterLink to="/register">Create one</RouterLink>
      </div>
    </AuthCard>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AuthLayout from '@/components/AuthLayout.vue'
import AuthCard from '@/components/AuthCard.vue'
import { apiFetch } from '@/lib/api'

const router = useRouter()
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

onMounted(() => {
  const token = localStorage.getItem('auth_token')
  if (token) router.replace('/home')
})

async function onSubmit() {
  error.value = null
  loading.value = true
  try {
    const res = await apiFetch('/v1/auth/login', {
      method: 'POST',
      body: { email: email.value, password: password.value },
    })
    localStorage.setItem('auth_token', res.token)
    localStorage.setItem('user_id', String(res.user_id))
    router.push('/home')
  } catch (e: any) {
    error.value = e?.message || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>