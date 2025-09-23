export function saveAuth(token: string, userId: string | number) {
  localStorage.setItem('auth_token', token);      
  localStorage.setItem('user_id', String(userId));
}

export function getToken(): string | null {
  return localStorage.getItem('auth_token');
}

export function getUserId(): string | null {
  return localStorage.getItem('user_id')
}

export function clearAuth() {
  localStorage.removeItem('auth_token');
  localStorage.removeItem('user_id');
}

export function authHeader(): Record<string, string> {
  const t = getToken();
  return t ? { Authorization: `Bearer ${t}` } : {};
}