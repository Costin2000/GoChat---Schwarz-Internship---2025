<template>
  <AuthLayout>
    <AuthCard title="Your Friends">
      
      <!-- All the specific content for the friends list goes here -->
      <div class="d-flex justify-content-between align-items-center mb-3">
        <!-- The title is now passed as a prop, so we can remove the h3 -->
        <div></div>
        <button class="btn btn-outline-secondary" :disabled="loading" @click="refresh">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2" />
          Refresh
        </button>
      </div>

      <div v-if="friends.length === 0 && !loading" class="text-muted">
        No friends yet.
      </div>

      <ul class="list-group list-group-flush">
        <li v-for="f in friends" :key="f.id"
            class="list-group-item d-flex align-items-center justify-content-between">
          <div class="d-flex align-items-center">
            <div class="rounded-circle d-flex align-items-center justify-content-center me-3"
                 style="width:44px;height:44px;background:#d1e7dd;color:#0f5132;font-weight:700;">
              {{ initials(f) }}
            </div>
            <div>
              <div class="fw-semibold">{{ fullName(f) }}</div>
              <div class="text-muted small">@{{ f.user_name }}</div>
            </div>
          </div>

          <button class="btn btn-success"
                  @click="openConversation(f)">
            Message
          </button>
        </li>
      </ul>

      <div class="text-center mt-4">
        <button v-if="nextToken" class="btn btn-outline-success px-4"
                :disabled="loading" @click="loadMore">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2" />
          Load more
        </button>
      </div>

    </AuthCard>
  </AuthLayout>
</template>

<script setup lang="ts">
import AuthLayout from '@/components/AuthLayout.vue'
import AuthCard from '@/components/AuthCard.vue'

import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch } from '@/lib/api'

type Friend = {
  id: number | string
  first_name?: string
  last_name?: string
  user_name?: string
  email?: string
  conversation_id?: number | string | null
}

const PAGE_SIZE = 10
const friends = ref<Friend[]>([])
const nextToken = ref<string>("")
const loading = ref(false)
const router = useRouter()

function fullName(f: Friend) {
  const fn = f.first_name?.trim() || ''
  const ln = f.last_name?.trim() || ''
  return `${fn} ${ln}`.trim() || f.user_name || f.email || String(f.id)
}

function initials(f: Friend) {
  const fn = (f.first_name || '').trim()
  const ln = (f.last_name || '').trim()
  const a = fn ? fn[0] : ''
  const b = ln ? ln[0] : ''
  return (a + b || (f.user_name?.[0] ?? 'U')).toUpperCase()
}

async function fetchFriends(token?: string) {
  loading.value = true
  try {
    const userId = localStorage.getItem('user_id')
    const body: any = { 
      user_id: userId, 
      page_size: PAGE_SIZE,
      show_friends: true 
    }
    
    if (token) body.next_page_token = token

    console.log("SENDING REQUEST BODY:", body);
    const res = await apiFetch<{ users: Friend[]; next_page_token?: string }>('/v1/friends', {
      method: 'POST',
      body,
    })
    console.log("RECEIVED RESPONSE:", res);

    if (token) friends.value.push(...(res?.users ?? []))
    else     friends.value = res?.users ?? []

    nextToken.value = res?.next_page_token || ''
  } finally {
    loading.value = false
  }
}

function loadMore() {
  if (!nextToken.value || loading.value) return
  fetchFriends(nextToken.value)
}
function refresh() {
  if (loading.value) return
  nextToken.value = ''
  fetchFriends()
}

function openConversation(f: Friend) {
  if (f.conversation_id) router.push({ name: 'Conversation', params: { id: String(f.conversation_id) } })
  else router.push({ name: 'Conversation', params: { id: 'new' }, query: { friend_id: String(f.id) } })
}

onMounted(() => { fetchFriends() })
</script>

<style scoped>
.list-group-item {
  background-color: transparent; /* Makes list items blend with the card background */
}
</style>

