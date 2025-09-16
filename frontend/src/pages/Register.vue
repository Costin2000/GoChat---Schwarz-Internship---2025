<template>
  <AuthLayout>
    <AuthCard title="Create account">
      <form @submit.prevent="onSubmit" novalidate>
        <div class="row">
          <div class="col-12 col-md-6 mb-3">
            <label class="form-label">First name</label>
            <input
              v-model.trim="first_name"
              class="form-control"
              required
              autocomplete="given-name"
            />
          </div>
          <div class="col-12 col-md-6 mb-3">
            <label class="form-label">Last name</label>
            <input
              v-model.trim="last_name"
              class="form-control"
              required
              autocomplete="family-name"
            />
          </div>
        </div>

        <div class="mb-3">
          <label class="form-label">Username</label>
          <input
            v-model.trim="user_name"
            class="form-control"
            required
            autocomplete="username"
          />
        </div>

        <div class="mb-3">
          <label class="form-label">Email</label>
          <input
            v-model.trim="email"
            type="email"
            class="form-control"
            required
            autocomplete="email"
          />
        </div>

        <div class="mb-2">
          <label class="form-label">Password</label>
          <input
            v-model="password"
            type="password"
            class="form-control"
            required
            minlength="6"
            autocomplete="new-password"
          />
        </div>

        <p v-if="error" class="text-danger small mb-2">{{ error }}</p>

        <!-- spacing peste buton -->
        <button class="btn btn-success w-100 mt-3" :disabled="loading">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2" />
          Create account
        </button>
      </form>

      <hr class="my-3" />
      <div class="text-center">
        <span class="me-1">Already have an account?</span>
        <RouterLink to="/login">Log in</RouterLink>
      </div>
    </AuthCard>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AuthLayout from '@/components/AuthLayout.vue'
import AuthCard from '@/components/AuthCard.vue'
import { apiFetch } from '@/lib/api'
import { saveAuth } from '@/lib/auth'

import type { CreateUserRequest } from '@/proto/services/user-base/proto/userbase'

const router = useRouter()

const first_name = ref('')
const last_name  = ref('')
const user_name  = ref('')
const email      = ref('')
const password   = ref('')

const loading = ref(false)
const error   = ref<string | null>(null)

type CreateUserResponse = {
  token?: string
  user?: { id?: number | string } | null
  user_id?: number | string
}

async function onSubmit() {
  error.value = null
  loading.value = true
  try {
    const payload: CreateUserRequest = {
      user: {
        first_name: first_name.value,
        last_name:  last_name.value,
        user_name:  user_name.value,
        email:      email.value,
        password:   password.value,
      }
    }

    const data = await apiFetch<CreateUserResponse>('/v1/user', {
      method: 'POST',
      body: payload,
    })

    let token = data?.token
    let uid   = data?.user?.id ?? data?.user_id

    if (!token || uid == null) {
      const loginResp = await apiFetch<{ token: string; user_id: number | string }>('/v1/auth/login', {
        method: 'POST',
        body: { email: email.value, password: password.value },
      })
      token = loginResp.token
      uid   = loginResp.user_id
    }

    if (token && uid != null) {
      saveAuth(token, uid)
      localStorage.setItem('token', String(token))
      router.push('/home')
      return
    }

    router.push('/login')
  } catch (e: any) {
    error.value = e?.message || 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>