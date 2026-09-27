import api from './index'

export interface PersonalAccessToken {
  id: number
  name: string
  token_prefix: string
  scopes?: string
  last_used_at?: string | null
  expires_at?: string | null
  revoked_at?: string | null
  created_at: string
}

export interface CreatedPAT extends PersonalAccessToken {
  token: string
}

export const patApi = {
  list: () => api.get<PersonalAccessToken[]>('/auth/tokens').then(r => r.data || []),
  create: (name: string, expiresAt?: string | null) =>
    api.post<CreatedPAT>('/auth/tokens', { name, expires_at: expiresAt || undefined }).then(r => r.data),
  revoke: (id: number) => api.delete(`/auth/tokens/${id}`),
}

/** Calls /auth/me with the PAT itself (not the session JWT) to prove the token works. */
export async function testPAT(token: string): Promise<{ ok: boolean; status: number; name?: string }> {
  const res = await fetch('/api/v1/auth/me', { headers: { Authorization: `Bearer ${token}` } })
  if (!res.ok) return { ok: false, status: res.status }
  const body = await res.json().catch(() => ({}))
  const user = body?.data ?? body
  return { ok: true, status: res.status, name: user?.display_name || user?.email }
}
