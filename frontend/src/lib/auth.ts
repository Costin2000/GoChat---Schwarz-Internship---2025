export function saveAuth(token: string, userId: string | number) {
  localStorage.setItem('token', token);      
  localStorage.setItem('user_id', String(userId));
}

export function getToken(): string | null {
  return localStorage.getItem('token') || localStorage.getItem('auth_token');
}

export function clearAuth() {
  localStorage.removeItem('token');
  localStorage.removeItem('auth_token');
  localStorage.removeItem('user_id');
}

export function authHeader(): Record<string, string> {
  const t = getToken();
  return t ? { Authorization: `Bearer ${t}` } : {};
}