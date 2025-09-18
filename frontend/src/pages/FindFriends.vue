<template>
  <div class="p-4">
    <h1 class="text-xl font-bold mb-4 text-white">Find Friends</h1>

    <div v-if="loading" class="text-green-200">Loading...</div>
    <div v-else-if="error" class="text-red-300">{{ error }}</div>
    <div v-else-if="users.length === 0" class="text-green-200">No users available.</div>

    <ul v-else class="space-y-2">
      <li
        v-for="u in users"
        :key="u.id"
        class="flex justify-between items-center border rounded p-2"
      >
        <div>
          <span class="user-name">{{ u.first_name }} {{ u.last_name }}</span>
          <p class="user-handle">@{{ u.user_name }}</p>
        </div>
        <button
          @click="sendFriendRequest(u.id)"
          class="bg-green-200 hover:bg-green-400 text-blue-300 px-4 py-1.5 rounded-lg shadow transition-colors"
        >
          Add Friend
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.user-name {
  color: #08ecb7;
  font-size: 2.1rem;
  font-weight: 600;
}

.user-handle {
  color: #35e3ef;
  font-size: 0.9rem;
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { getToken, getUserId } from '@/lib/auth';
import { apiFetch } from '@/lib/api';

interface User {
  id: string;
  first_name: string;
  last_name: string;
  user_name: string,
}

const router = useRouter();

const users = ref<User[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

onMounted(async () => {
  if (!getToken()) {
    router.push("/login");
    return;
  }

  try {
    const res = await fetchNonFriends(getUserId());
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
    users.value = users.value.filter((u) => u.id !== userId);
  } catch (e: any) {
    error.value = `Could not send request: ${e.message}`;
  }
}

function fetchNonFriends(userId: string) {
  return apiFetch<{ users: User[] }>(
    '/v1/friends',
    {
      method: "POST",
      json: {
        user_id: userId,
        show_friends: false}
    }
  );
}

function createFriendRequest(senderId: string, receiverId: string) {
  return apiFetch("/v1/friend-request", {
    method: "POST",
    json: {
      sender_id: senderId,
      receiver_id: receiverId
    }
  });
}

</script>