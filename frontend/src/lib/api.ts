import { authHeader } from './auth'
const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

type Opts = RequestInit & { json?: any }

export async function apiFetch<T = any>(path: string, opts: Opts = {}) {
  const headers = new Headers(opts.headers || {})
  Object.entries(authHeader()).forEach(([k, v]) => headers.set(k, v))

  let body = opts.body

  if (opts.json !== undefined) {
    headers.set('Content-Type', 'application/json')
    body = JSON.stringify(opts.json)
  } else if (body && typeof body === 'object' && !(body instanceof FormData) && !(body instanceof Blob)) {
    headers.set('Content-Type', 'application/json')
    body = JSON.stringify(body)
  }

  const res = await fetch(`${API_BASE}${path}`, { ...opts, headers, body })
  const text = await res.text()

  let data: any = {}
  try { data = text ? JSON.parse(text) : {} } catch {}
  if (!res.ok) throw new Error(data?.message || `HTTP ${res.status}`)

  return data as T
}

export interface User {
  id: string;
  first_name: string;
  last_name: string;
  user_name: string,
}

export function fetchNonFriends(userId: string) {
  return apiFetch<{ users: User[] }>(
    '/v1/friends',
    {
      method: "POST",
      json: {
        user_id: userId,
        show_friends: false
      }
    }
  );
}

export function createFriendRequest(senderId: string, receiverId: string) {
  return apiFetch("/v1/friend-request", {
    method: "POST",
    json: {
      sender_id: senderId,
      receiver_id: receiverId
    }
  });
}