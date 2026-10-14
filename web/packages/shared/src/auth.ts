import { apiGet, apiPost, clearToken, setToken } from './http'

export type User = {
  id: number
  store_id: number
  username: string
  display_name: string
}

export type Me = {
  user: User
  roles: { id: number; code: string; name: string }[]
  permission_codes: string[]
}

export async function login(username: string, password: string) {
  const data = await apiPost<{ token: string; user: User }>('/auth/login', { username, password })
  setToken(data.token)
  return data
}

export async function fetchMe() {
  return apiGet<Me>('/auth/me')
}

export function logout() {
  clearToken()
}
