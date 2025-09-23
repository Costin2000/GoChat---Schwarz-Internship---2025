<template>
  <div class="page-container">
    <h1 class="title">Conversations</h1>

    <div v-if="loading" class="loading-text">Loading conversations...</div>
    <div v-else-if="error" class="error-text">{{ error }}</div>

    <div v-else class="chat-container">
      <div class="conversations-panel">
        <ul v-if="conversations.length > 0" class="conversations-list">
          <li v-for="convo in conversations" :key="convo.id" class="conversation-item"
            :class="{ 'selected': convo.id === selectedConversationId }" @click="selectConversation(convo)">
            <div class="user-info">
              <span class="full-name">
                {{ userCache.get(getOtherParticipantId(convo))?.first_name || 'Loading...' }}
                {{ userCache.get(getOtherParticipantId(convo))?.last_name }}
              </span>
              <span class="username">@{{ userCache.get(getOtherParticipantId(convo))?.user_name || '...' }}</span>
            </div>
          </li>
        </ul>
        <div v-else class="info-text">No conversations found. Make some friends to start chatting! </div>
      </div>

      <div class="messages-panel">
        <div v-if="!selectedConversationId" class="info-text ">
          Select a conversation to view messages
        </div>
        <div v-else class="message-view-container">
          <div v-if="messagesLoading" class="loading-text">Loading messages...</div>
        <div v-else-if="messages.length === 0" class="info-text initial-message">
            No messages yet. Say hello!
          </div>
          <ul v-else ref="messageListEl" class="messages-list">
            <li v-for="msg in messages" :key="msg.id" class="message-item"
              :class="isMyMessage(msg.sender_id) ? 'sent' : 'received'">
              <div class="message-bubble">
                {{ msg.content }}
              </div>
            </li>
          </ul>
          <form @submit.prevent="sendMessage" class="message-input-form">
            <input v-model="newMessageContent" type="text" class="message-input" placeholder="Type a message..."
              autocomplete="off" />
            <button type="submit" class="send-btn">Send</button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  padding: 1rem;
  height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
}

.title {
  font-size: 1.5rem;
  font-weight: bold;
  margin-bottom: 1rem;
  color: white;
}

.loading-text,
.error-text,
.info-text {
  color: #ccc;
  text-align: center;
  padding: 2rem;
}

.initial-message {
  height: 100%;
}

.error-text {
  color: #ff8a8a;
}

.chat-container {
  display: flex;
  flex-grow: 1;
  gap: 1rem;
  border: 1px solid #35e3ef40;
  border-radius: 1.5rem;
  overflow: hidden;
}

.conversations-panel {
  flex: 0 0 350px;
  background-color: rgba(0, 0, 0, 0.2);
  border-right: 1px solid #35e3ef40;
  overflow-y: auto;
}

.conversations-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.conversation-item {
  padding: 1rem;
  cursor: pointer;
  border-bottom: 1px solid #35e3ef20;
  transition: background-color 0.2s ease;
}

.conversation-item:hover {
  background-color: #35e3ef20;
}

.conversation-item.selected {
  background-color: #08ecb7;
}

.conversation-item.selected .full-name,
.conversation-item.selected .username,
.conversation-item.selected .message-preview {
  color: #0c3329;
}

.user-info {
  display: flex;
  flex-direction: column;
  margin-bottom: 0.25rem;
}

.full-name {
  color: #08ecb7;
  font-weight: bold;
}

.username {
  font-size: 0.8rem;
  color: #35e3ef;
}

.message-preview {
  font-size: 0.9rem;
  color: #a0a0a0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.messages-panel {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
}

.messages-panel .placeholder {
  margin: auto;
}

.message-view-container {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.messages-list {
  list-style: none;
  padding: 1rem;
  margin: 0;
  overflow-y: auto;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.message-item {
  display: flex;
  max-width: 70%;
}

.message-bubble {
  padding: 0.5rem 1rem;
  border-radius: 1.25rem;
  color: white;
  word-break: break-word;
}

.message-item.sent {
  justify-content: flex-end;
  align-self: flex-end;
}

.sent .message-bubble {
  background-color: #257654;
  border-bottom-right-radius: 0.25rem;
}

.message-item.received {
  justify-content: flex-start;
}

.received .message-bubble {
  background-color: #374151;
  border-bottom-left-radius: 0.25rem;
}

.message-input-form {
  display: flex;
  padding: 1rem;
  gap: 0.5rem;
  border-top: 1px solid #35e3ef40;
}

.message-input {
  flex-grow: 1;
  background-color: #1f2937;
  border: 1px solid #4b5563;
  border-radius: 0.5rem;
  padding: 0.5rem 1rem;
  color: white;
}

.message-input:focus {
  outline: none;
  border-color: #08ecb7;
}

.send-btn {
  background-color: #14B8A6;
  color: white;
  font-weight: 600;
  border: none;
  padding: 0.5rem 1.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.send-btn:hover {
  background-color: #0D9488;
}
</style>


<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue';
import { useRouter } from 'vue-router';
import { getToken, getUserId } from '@/lib/auth';
import {
  listMessages,
  listConversations,
  createMessage,
  getUser,
} from '@/lib/api';

import { User } from '@/proto/services/user-base/proto/userbase'
import { Conversation } from '@/proto/services/conversation-base/proto/conversation'
import { Message } from '@/proto/services/message-base/proto/messagebase'

const router = useRouter();
console.log('Component loaded. Current User ID:', getUserId());

const conversations = ref<Conversation[]>([]);
const userCache = ref(new Map<string, User>());
const loading = ref(true);
const error = ref<string | null>(null);

const selectedConversationId = ref<string | null>(null);
const messages = ref<Message[]>([]);
const messagesLoading = ref(false);
const newMessageContent = ref('');
const messageListEl = ref<HTMLElement | null>(null);


onMounted(async () => {
  if (!getToken()) {
    router.push("/login");
    return;
  }
  try {
    const res = await listConversations(getUserId());
    conversations.value = res.conversations || [];
    await fetchAllUsersForConversations(conversations.value);
  } catch (e: any) {
    error.value = `Failed to load conversations: ${e.message}`;
  } finally {
    loading.value = false;
  }
});



async function fetchAllUsersForConversations(convos: Conversation[]) {
  const userIds = new Set<string>();
  convos.forEach(c => {
    userIds.add(c.user1_id);
    userIds.add(c.user2_id);
  });

  for (const userId of userIds) {
    if (!userCache.value.has(userId)) {
      try {
        const user = await getUser(userId);
        userCache.value.set(userId, user);
      } catch (e) {
        console.error(`Failed to fetch user ${userId}`, e);
      }
    }
  }
}


async function scrollToBottom() {
  await nextTick();
  const el = messageListEl.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
}


async function selectConversation(conversation: Conversation) {
  selectedConversationId.value = conversation.id;
  messages.value = [];
  messagesLoading.value = true;
  try {
    const res = await listMessages(conversation.id);

    console.log('Received API response for messages:', res);

    res.messages.sort((a: Message, b: Message) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
    messages.value = res.messages || [];
    watch(messages, () => {
      scrollToBottom();
    }, { deep: true });
  } catch (e: any) {
    error.value = `Failed to load messages: ${e.message}`;
  } finally {
    messagesLoading.value = false;
  }
}

async function sendMessage() {
  const currentUserId = getUserId();
  const content = newMessageContent.value.trim();

  if (!content || !selectedConversationId.value || !currentUserId) {
    return;
  }

  try {
    await createMessage(
      selectedConversationId.value,
      currentUserId,
      content
    );
    const res = await listMessages(selectedConversationId.value);

    res.messages.sort((a: Message, b: Message) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
    messages.value = res.messages || [];

    newMessageContent.value = '';
    scrollToBottom();

  } catch (e: any) {
    error.value = `Failed to send message: ${e.message}`;
  }
}

function getOtherParticipantId(conversation: Conversation): string {
  return conversation.user1_id === getUserId() ? conversation.user2_id : conversation.user1_id;
}

function isMyMessage(senderId: number | undefined): boolean {
  const currentUserId = getUserId();
  console.log(`Comparing message senderId: "${senderId}" with currentUserId: "${currentUserId}"`);

  if (senderId === null || senderId === undefined) {
    return false;
  }

  return senderId.toString() === currentUserId;
}
</script>
