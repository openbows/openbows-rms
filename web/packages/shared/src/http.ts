import axios, { type AxiosInstance } from 'axios'

export type ApiBody<T = unknown> = {
  code: number
  msg: string
  data: T
}

export type PageData<T> = {
  list: T[]
  total: number
  page: number
  page_size: number
}

const TOKEN_KEY = 'ob_rms_token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function createHttp(baseURL = '/apis/v1'): AxiosInstance {
  const http = axios.create({ baseURL, timeout: 30000 })
  http.interceptors.request.use((cfg) => {
    const t = getToken()
    if (t) cfg.headers.Authorization = `Bearer ${t}`
    return cfg
  })
  http.interceptors.response.use(
    (res) => {
      const body = res.data as ApiBody
      if (body && typeof body.code === 'number' && body.code !== 200) {
        return Promise.reject(new Error(body.msg || '请求失败'))
      }
      return res
    },
    (err) => Promise.reject(err)
  )
  return http
}

export const http = createHttp()

export async function apiGet<T>(url: string, params?: Record<string, unknown>) {
  const res = await http.get<ApiBody<T>>(url, { params })
  return res.data.data
}

export async function apiPost<T>(url: string, data?: unknown) {
  const res = await http.post<ApiBody<T>>(url, data)
  return res.data.data
}

export async function apiPut<T>(url: string, data?: unknown) {
  const res = await http.put<ApiBody<T>>(url, data)
  return res.data.data
}

export async function apiDelete<T = null>(url: string) {
  const res = await http.delete<ApiBody<T>>(url)
  return res.data.data
}

export async function apiPage<T>(url: string, params?: Record<string, unknown>) {
  const res = await http.get<ApiBody<PageData<T>>>(url, { params })
  return res.data.data
}
