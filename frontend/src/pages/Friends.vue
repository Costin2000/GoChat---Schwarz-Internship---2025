<template>
  <AuthLayout>
    <AuthCard title="Your Friends" maxWidth="900px">
      
      <!-- Search Bar -->
      <div class="mb-3">
        <input 
          v-model="searchQuery"
          type="text"
          class="form-control"
          placeholder="Search friends by name..."
        />
      </div>

      <div class="d-flex justify-content-between align-items-center mb-3">
        <div></div>
        <button class="btn btn-outline-secondary" :disabled="loading" @click="refresh">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2" />
          Refresh
        </button>
      </div>

      <div v-if="filteredFriends.length === 0 && !loading" class="text-muted">
        No friends found.
      </div>

      <ul class="list-group list-group-flush">
        <li v-for="f in filteredFriends" :key="f.id"
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
                  @click="openConversation(f)"
                  :disabled="isOpeningConvo === f.id">
            <span v-if="isOpeningConvo === f.id" class="spinner-border spinner-border-sm" />
            <span v-else>Message</span>
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
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, listConversations } from '@/lib/api'
import { getUserId } from '@/lib/auth'
import type { Conversation } from '@/proto/services/conversation-base/proto/conversation'

type Friend = {
  id: string
  first_name?: string
  last_name?: string
  user_name?: string
}

const PAGE_SIZE = 10
const friends = ref<Friend[]>([])
const nextToken = ref<string>("")
const loading = ref(false)
const isOpeningConvo = ref<string | null>(null)
const searchQuery = ref("")
const router = useRouter()

function fullName(f: Friend) {
  const fn = f.first_name?.trim() || ''
  const ln = f.last_name?.trim() || ''
  return `${fn} ${ln}`.trim() || f.user_name || String(f.id)
}

function initials(f: Friend) {
  const fn = (f.first_name || '').trim()
  const ln = (f.last_name || '').trim()
  const a = fn ? fn[0] : ''
  const b = ln ? ln[0] : ''
  return (a + b || (f.user_name?.[0] ?? 'U')).toUpperCase()
}

const filteredFriends = computed(() => {
  if (!searchQuery.value.trim()) return friends.value
  const q = searchQuery.value.toLowerCase()
  return friends.value.filter(f => {
    const name = `${f.first_name || ''} ${f.last_name || ''}`.toLowerCase()
    return name.includes(q)
  })
})

async function fetchFriends(token?: string) {
  loading.value = true
  try {
    const userId = getUserId()
    const body: any = { 
      user_id: userId, 
      page_size: PAGE_SIZE,
      show_friends: true 
    }
    
    if (token) body.next_page_token = token

    const res = await apiFetch<{ users: Friend[]; next_page_token?: string }>('/v1/friends', {
      method: 'POST',
      body,
    })

    if (token) friends.value.push(...(res?.users ?? []))
    else     friends.value = res?.users ?? []

    nextToken.value = res?.next_page_token || ''
  } finally {
    loading.value = false
  }
}

async function openConversation(friend: Friend) {
  isOpeningConvo.value = friend.id;
  try {
    const currentUserId = getUserId();
    if (!currentUserId) {
      router.push('/login');
      return;
    }

    console.log('--- STARTING CONVERSATION SEARCH ---');
    console.log(`Searching for convo between ME (ID: "${currentUserId}") and FRIEND (ID: "${friend.id}")`);
    
    const res = await listConversations(currentUserId);
    const allConversations = (res.conversations || []) as any[];

    console.log('Fetched conversations to search in:', allConversations);

    const conversation = allConversations.find(c => {
      const user1 = c.user1Id || c.user1_id;
      const user2 = c.user2Id || c.user2_id;
      
      console.log(`Checking convo ID ${c.id}: user1="${user1}", user2="${user2}"`);
      const isMatch = (user1 === currentUserId && user2 === friend.id) ||
                      (user2 === currentUserId && user1 === friend.id);
      if (isMatch) {
        console.log(`%cMATCH FOUND!`, 'color: #00ff00; font-weight: bold;', c);
      }
      return isMatch;
    });

    if (conversation) {
      console.log('Navigating with existing conversation ID:', conversation.id);
      router.push({ 
        name: 'Conversations', 
        query: { conversation: conversation.id } 
      });
    } else {
      console.error('NO MATCH FOUND. Falling back to new_with_user.');
      router.push({ name: 'Conversations', query: { new_with_user: friend.id } });
    }
  } catch (error) {
    console.error("Failed to find or open conversation:", error);
  } finally {
    isOpeningConvo.value = null;
  }
}

function loadMore() {
  if (!nextToken.value || loading.value) return
  fetchFriends(nextToken.value)
}
function refresh() {
  if (loading.value) return
  nextToken.value = ''
  friends.value = []
  fetchFriends()
}

onMounted(() => { 
  fetchFriends() 
})
</script>

<style scoped>
.list-group-item {
  background-color: transparent;
}
</style>