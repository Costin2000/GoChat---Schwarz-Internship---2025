<template>
  <AuthLayout>
      <div class="p-4">
        <h1 class="text-xl font-bold mb-4" style="color: white">Friend Requests</h1>

        <div v-if="loading && requests.length === 0" class="text-green-200">
          Loading...
        </div>
        <div v-else-if="error" class="text-red-300">
          {{ error }}
        </div>
        <div v-else-if="requests.length === 0" style="color: white">
          No pending friend requests.
        </div>

        <ul v-else class="space-y-2">
          <li
            v-for="r in requests"
            :key="r.id"
            class="flex justify-between items-center border rounded p-2 request-item"
          >
            <div>
              <!-- Primul rând: nume complet -->
              <p class="font-semibold text-lg" style="color: white">
                {{ usersMap[r.sender_id]?.first_name }} {{ usersMap[r.sender_id]?.last_name }}
              </p>
              <!-- Al doilea rând: username -->
              <p class="text-sm" style="color: #08ecb7">
                @{{ usersMap[r.sender_id]?.user_name }}
              </p>
              <!-- Al treilea rând: data -->
              <p class="text-xs" style="color: #bcded0">
                {{ formatDate(r.created_at) }}
              </p>
            </div>

            <div class="flex btn-group">
              <button
                @click="updateRequest(r.id, 'STATUS_ACCEPTED')"
                class="action-btn accept-btn"
                :disabled="actionLoading === r.id"
                aria-label="Accept friend request"
              >
                <span v-if="actionLoading === r.id">...</span>
                <span v-else>Accept</span>
              </button>

              <button
                @click="updateRequest(r.id, 'STATUS_REJECTED')"
                class="action-btn reject-btn"
                :disabled="actionLoading === r.id"
                aria-label="Reject friend request"
              >
                <span v-if="actionLoading === r.id">...</span>
                <span v-else>Reject</span>
              </button>
            </div>
          </li>
        </ul>

        <div class="mt-4 text-center">
          <button
            v-if="nextPageToken && !loading"
            @click="loadMore"
            class="load-more-btn"
          >
            Load more
          </button>
          <div v-if="loading && requests.length > 0" class="text-green-200 mt-2">
            Loading more...
          </div>
        </div>
      </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { apiFetch, createConversation } from "@/lib/api";
import { getToken, getUserId } from "@/lib/auth";
import { Conversation } from '@/proto/services/conversation-base/proto/conversation'
import AuthLayout from '@/components/AuthLayout.vue'
import AuthCard from '@/components/AuthCard.vue'

const router = useRouter();

type FriendRequest = {
  id: string;
  sender_id: string;
  receiver_id: string;
  status: string;
  created_at: string;
};

type User = {
  id: string;
  first_name: string;
  last_name: string;
  user_name: string;
};

const requests = ref<FriendRequest[]>([]);
const usersMap = ref<Record<string, User>>({}); // mapăm userId -> user
const nextPageToken = ref<string>("");
const loading = ref(false);
const error = ref<string | null>(null);
const actionLoading = ref<string | null>(null);

onMounted(() => {
  if (!getToken()) {
    router.push("/login");
    return;
  }
  fetchRequests("");
});

async function fetchRequests(token: string) {
  loading.value = true;
  error.value = null;

  try {
    const res = await apiFetch<{
      nextPageToken: string;
      requests: FriendRequest[];
    }>("/v1/friend-requests", {
      method: "POST",
      body: {
        pageSize: 10,
        filters: [
          { receiver_id: getUserId() },
          { status: "pending" },
        ],
        nextPageToken: token,
      },
    });

    requests.value = [...requests.value, ...res.requests];
    nextPageToken.value = res.nextPageToken;

    // Preluăm user details pentru fiecare sender_id
    await fetchUsers(res.requests.map(r => r.sender_id));
  } catch (e: any) {
    error.value = e?.message || "Failed to fetch friend requests";
  } finally {
    loading.value = false;
  }
}

async function fetchUsers(userIds: string[]) {
  if (!userIds.length) return;

  try {
    const res = await apiFetch<{
      users: User[];
    }>("/v1/users:list", {
      method: "POST",
      body: {
        pageSize: userIds.length,
        filters: [
          {
            userIds: { userId: userIds },
          },
        ],
      },
    });

    res.users.forEach(u => {
      usersMap.value[u.id] = u;
    });
  } catch (e: any) {
    console.error("Failed to fetch user details", e);
  }
}

function loadMore() {
  if (nextPageToken.value) {
    fetchRequests(nextPageToken.value);
  }
}

async function updateRequest(id: string, status: "STATUS_ACCEPTED" | "STATUS_REJECTED") {
  actionLoading.value = id;
  error.value = null;

  try {
    await apiFetch(`/v1/friend-request/${id}`, {
      method: "PATCH",
      body: {
        friend_request: {
          id,
          status,
        },
        field_mask: "status",
      },
    });
    const accReq = requests.value.find((r) => r.id === id)
    console.log(accReq)
    requests.value = requests.value.filter((r) => r.id !== id);
    if (status == "STATUS_ACCEPTED") {
      await createConversation(accReq?.receiver_id, accReq?.sender_id)
    }
  } catch (e: any) {
    error.value = e?.message || `Failed to ${status === "STATUS_ACCEPTED" ? "accept" : "reject"} request`;
  } finally {
    actionLoading.value = null;
  }
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString();
}
</script>

<style scoped>
/* Grup de butoane - adaugam spatiu fix intre ele */
.btn-group {
  gap: 0.75rem; /* ~space-x-3 */
}

/* Butoane Accept / Reject */
.action-btn {
  background-color: #bbf7d0 !important; /* verde deschis */
  color: #000 !important;               /* text negru */
  font-weight: 600;
  padding: 0.35rem 0.7rem;
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  transition: background-color 0.15s ease-in-out, transform 0.08s;
  box-shadow: 0 1px 2px rgba(0,0,0,0.08);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.action-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  transform: none;
}

.accept-btn:hover:not(:disabled) {
  background-color: #34d399 !important;
  transform: translateY(-1px);
}

.reject-btn:hover:not(:disabled) {
  background-color: #f87171 !important; /* rosu */
  transform: translateY(-1px);
}

.load-more-btn {
  background-color: #bbf7d0 !important;
  color: #000 !important;
  font-weight: 600;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  transition: background-color 0.15s ease-in-out, transform 0.08s;
  box-shadow: 0 2px 4px rgba(0,0,0,0.08);
}

.load-more-btn:hover:not(:disabled) {
  background-color: #34d399 !important;
  transform: translateY(-1px);
}

.load-more-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

/* Card request */
.request-item {
  background: rgba(255,255,255,0.02);
}
</style>