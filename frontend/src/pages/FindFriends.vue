<template>
  <AuthLayout>
    <AuthCard title="Find Friends" maxWidth="900px" class="find-friends-card">
      
      <div class="content-wrapper">
        
        <div class="header">
          <input 
            v-model="searchQuery"
            type="text"
            class="form-control"
            placeholder="Search users by name..."
          />
        </div>

        <div class="user-list-container">
          <div v-if="loading" class="text-muted text-center py-5">Loading...</div>
          <div v-else-if="error" class="text-danger text-center py-5">{{ error }}</div>
          <div v-else-if="filteredUsers.length === 0" class="text-muted text-center py-5">
            No matching users found.
          </div>

          <ul v-else class="list-group list-group-flush">
            <li
              v-for="u in filteredUsers"
              :key="u.id"
              class="list-group-item d-flex align-items-center justify-content-between"
            >
              <div class="d-flex align-items-center">
                <div class="user-avatar">
                  {{ initials(u) }}
                </div>
                <div>
                  <div class="user-name">{{ u.first_name }} {{ u.last_name }}</div>
                  <p class="user-handle">@{{ u.user_name }}</p>
                </div>
              </div>
              <button
                @click="sendFriendRequest(u.id)"
                class="add-friend-btn"
              >
                Add Friend
              </button>
            </li>
          </ul>
        </div>
      </div>

    </AuthCard>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { getToken, getUserId } from '@/lib/auth';
import { User, fetchNonFriends, createFriendRequest } from '@/lib/api';
import AuthLayout from '@/components/AuthLayout.vue'
import AuthCard from '@/components/AuthCard.vue';

const router = useRouter();
const users = ref<User[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const searchQuery = ref("")

function initials(f: User) {
  const fn = (f.first_name || '').trim();
  const ln = (f.last_name || '').trim();
  const a = fn ? fn[0] : '';
  const b = ln ? ln[0] : '';
  return (a + b || (f.user_name?.[0] ?? 'U')).toUpperCase();
}

const filteredUsers = computed(() => {
  if (!searchQuery.value.trim()) return users.value
  const q = searchQuery.value.toLowerCase()
  return users.value.filter(u => {
    const name = `${u.first_name || ''} ${u.last_name || ''}`.toLowerCase()
    return name.includes(q)
  })
})

onMounted(async () => {
  if (!getToken()) {
    router.push("/login");
    return;
  }

  try {
    const res = await fetchNonFriends(getUserId());
    res.users.sort((a: User, b: User) => a.user_name.localeCompare(b.user_name));
    users.value = res.users;
  } catch (e: any) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
});

async function sendFriendRequest(userId: string) {
  try {
    await createFriendRequest(getUserId(), userId);
    users.value = users.value.filter((u: User) => u.id !== userId);
  } catch (e: any) {
    error.value = `Could not send request: ${e.message}`;
  }
}
</script>

<style scoped>
.find-friends-card {
  height: 85vh; 
  display: flex;
  flex-direction: column;
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  overflow: hidden;
}

.header {
  padding-bottom: 1rem;
  border-bottom: 1px solid #dee2e6;
  margin-bottom: 1rem;
}

.user-list-container {
  flex-grow: 1;
  overflow-y: auto;
  margin-right: -1rem;
  padding-right: 1rem;
}

.list-group-item {
  background-color: transparent;
  padding: 1rem 0;
}

.list-group-item:first-child {
  padding-top: 0;
}

.user-avatar {
  width: 44px;
  height: 44px;
  background-color: #d1e7dd;
  color: #0f5132;
  font-weight: 700;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 1rem;
}

.user-name {
  color: #343a40; 
  font-weight: 600;
}

.user-handle {
  color: #6c757d;
  font-size: 0.9rem;
  margin: 0;
}

.add-friend-btn {
  background-color: #198754;
  color: white;
  font-weight: 600;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  cursor: pointer;
  box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  transition: background-color 0.2s ease-in-out;
}
.add-friend-btn:hover {
  background-color: #157347;
}

.user-list-container::-webkit-scrollbar {
  width: 8px;
}
.user-list-container::-webkit-scrollbar-track {
  background: transparent;
}
.user-list-container::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 10px;
}
.user-list-container::-webkit-scrollbar-thumb:hover {
  background: #aaa;
}
</style>

