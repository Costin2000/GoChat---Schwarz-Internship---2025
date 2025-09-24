<template>
  <AuthLayout>
    <AuthCard title="Friend Requests" maxWidth="900px">
      
      <div v-if="loading && requests.length === 0" class="text-muted text-center py-5">
        Loading...
      </div>
      <div v-else-if="error" class="text-danger text-center py-5">
        {{ error }}
      </div>
      <div v-else-if="requests.length === 0" class="text-muted text-center py-5">
        No pending friend requests.
      </div>

      <ul v-else class="list-group list-group-flush">
        <li
          v-for="r in requests"
          :key="r.id"
          class="list-group-item d-flex align-items-center justify-content-between"
        >
          <div class="d-flex align-items-center">
            <div class="user-avatar">
              {{ initials(usersMap[r.sender_id]) }}
            </div>
            <div>
              <div class="user-name">
                {{ usersMap[r.sender_id]?.first_name }} {{ usersMap[r.sender_id]?.last_name }}
              </div>
              <p class="user-handle">
                @{{ usersMap[r.sender_id]?.user_name }}
              </p>
              <p class="request-date">
                {{ formatDate(r.created_at) }}
              </p>
            </div>
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
        <div v-if="loading && requests.length > 0" class="text-muted mt-2">
          Loading more...
        </div>
      </div>
    </AuthCard>
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
const usersMap = ref<Record<string, User>>({});
const nextPageToken = ref<string>("");
const loading = ref(false);
const error = ref<string | null>(null);
const actionLoading = ref<string | null>(null);

function initials(user?: User) {
  if (!user) return '?';
  const fn = (user.first_name || '').trim();
  const ln = (user.last_name || '').trim();
  const a = fn ? fn[0] : '';
  const b = ln ? fn[0] : '';
  return (a + b || (user.user_name?.[0] ?? 'U')).toUpperCase();
}

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
    const res = await apiFetch<{ users: User[] }>("/v1/users:list", {
      method: "POST",
      body: {
        pageSize: userIds.length,
        filters: [{ userIds: { userId: userIds } }],
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
        friend_request: { id, status },
        field_mask: "status",
      },
    });
    const accReq = requests.value.find((r) => r.id === id);
    requests.value = requests.value.filter((r) => r.id !== id);
    if (status == "STATUS_ACCEPTED") {
      await createConversation(accReq?.receiver_id, accReq?.sender_id);
    }
  } catch (e: any) {
    error.value = e?.message || `Failed to ${status === "STATUS_ACCEPTED" ? "accept" : "reject"} request`;
  } finally {
    actionLoading.value = null;
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return d.toLocaleString();
}
</script>

<style scoped>
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

.request-date {
  color: #adb5bd;
  font-size: 0.8rem;
  margin-top: 0.25rem;
}

.btn-group {
  gap: 0.75rem;
}

.action-btn {
  background-color: #e9ecef;
  color: #212529;
  font-weight: 600;
  padding: 0.35rem 0.7rem;
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.accept-btn:hover:not(:disabled) {
  background-color: #198754;
  color: white;
}

.reject-btn:hover:not(:disabled) {
  background-color: #dc3545;
  color: white;
}

.load-more-btn {
  background-color: #f8f9fa;
  color: #212529;
  font-weight: 600;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  border: 1px solid #dee2e6;
  cursor: pointer;
  transition: all 0.2s ease;
}

.load-more-btn:hover:not(:disabled) {
  background-color: #e9ecef;
}
</style>

