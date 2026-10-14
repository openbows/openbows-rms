import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, fetchMe, logout as apiLogout, getToken, type Me, type User } from '@openbows/shared'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(getToken())
  const user = ref<User | null>(null)
  const me = ref<Me | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value)
  const displayName = computed(() => user.value?.display_name || user.value?.username || '')

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const data = await apiLogin(username, password)
      token.value = data.token
      user.value = data.user
      await loadMe()
    } finally {
      loading.value = false
    }
  }

  async function loadMe() {
    if (!getToken()) return
    me.value = await fetchMe()
    user.value = me.value.user
    token.value = getToken()
  }

  function logout() {
    apiLogout()
    token.value = ''
    user.value = null
    me.value = null
  }

  return { token, user, me, loading, isLoggedIn, displayName, login, loadMe, logout }
})
