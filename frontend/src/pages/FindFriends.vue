<template>
  <div class="p-4">
    <h1 class="text-xl font-bold mb-4 text-white">Find Friends</h1>

    <div v-if="loading" class="text-green-200">Loading...</div>
    <div v-else-if="error" class="text-red-300">{{ error }}</div>
    <div v-else-if="users.length === 0" class="text-white">No users available.</div>

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
          class="add-friend-btn"
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

.add-friend-btn {
  background-color: #257654;
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
  background-color: #159f65
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { getToken, getUserId } from '@/lib/auth';
import { User, fetchNonFriends, createFriendRequest } from '@/lib/api';



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