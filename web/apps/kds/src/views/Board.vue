<template>
  <div class="board">
    <header class="topbar">
      <div class="brand">厨房显示 KDS</div>
      <div class="filters">
        <span>档口</span>
        <el-select v-model="stationId" clearable placeholder="全部档口" style="width:180px" @change="loadTickets">
          <el-option v-for="s in stations" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-tag type="info">每 3 秒自动刷新</el-tag>
        <el-button @click="loadTickets">立即刷新</el-button>
        <span class="user">{{ auth.displayName }}</span>
        <el-button type="danger" link @click="onLogout">退出</el-button>
      </div>
    </header>

    <div class="legend">
      <span class="dot pending"></span>待制作
      <span class="dot cooking"></span>制作中
      <span class="dot done"></span>已完成
    </div>

    <div class="cards" v-loading="loading">
      <div v-for="v in tickets" :key="v.ticket.id" class="card" :class="v.ticket.status">
        <div class="card-head">
          <div>
            <div class="order-no">{{ v.ticket.order_no }}</div>
            <div class="sub">桌台 #{{ v.ticket.table_id }} · 档口 #{{ v.ticket.station_id }}</div>
          </div>
          <el-tag :type="statusType(v.ticket.status)" effect="dark">{{ statusLabel(v.ticket.status) }}</el-tag>
        </div>
        <ul class="items">
          <li v-for="it in v.items" :key="it.id">
            <span class="iname">{{ it.dish_name }}</span>
            <span v-if="it.spec_name" class="ispec">{{ it.spec_name }}</span>
            <span class="iqty">×{{ it.qty }}</span>
          </li>
        </ul>
        <div class="card-actions">
          <el-button
            v-if="v.ticket.status === 'pending'"
            type="warning"
            @click="startTicket(v.ticket.id)"
          >开始制作</el-button>
          <el-button
            v-if="v.ticket.status === 'cooking'"
            type="success"
            @click="doneTicket(v.ticket.id)"
          >出餐完成</el-button>
        </div>
      </div>
      <div v-if="!loading && !tickets.length" class="empty">暂无厨单</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { apiGet, apiPage, apiPost } from '@openbows/shared'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const stations = ref<any[]>([])
const stationId = ref<number | undefined>(undefined)
const tickets = ref<any[]>([])
const loading = ref(false)
let timer: number | undefined

function statusLabel(s: string) {
  return ({ pending: '待制作', cooking: '制作中', done: '已完成' } as any)[s] || s
}
function statusType(s: string) {
  return ({ pending: 'info', cooking: 'warning', done: 'success' } as any)[s] || 'info'
}

async function loadStations() {
  const data = await apiPage<any>('/stations', { page: 1, page_size: 100 })
  stations.value = data.list || []
}

async function loadTickets() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {}
    if (stationId.value) params.station_id = stationId.value
    tickets.value = (await apiGet<any[]>('/kds/tickets', params)) || []
  } catch (e: any) {
    ElMessage.error(e.message || '加载厨单失败')
  } finally {
    loading.value = false
  }
}

async function startTicket(id: number) {
  try {
    await apiPost(`/kds/tickets/${id}/start`, {})
    ElMessage.success('已开始制作')
    loadTickets()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

async function doneTicket(id: number) {
  try {
    await apiPost(`/kds/tickets/${id}/done`, {})
    ElMessage.success('出餐完成')
    loadTickets()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

function onLogout() {
  auth.logout()
  router.push('/login')
}

onMounted(async () => {
  try {
    await auth.loadMe()
  } catch {
    auth.logout()
    router.push('/login')
    return
  }
  await loadStations()
  await loadTickets()
  timer = window.setInterval(loadTickets, 3000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.board { min-height: 100vh; display: flex; flex-direction: column; }
.topbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 20px; background: #0b1220; border-bottom: 1px solid #1f2937;
}
.brand { font-size: 20px; font-weight: 700; }
.filters { display: flex; align-items: center; gap: 10px; }
.user { color: #9ca3af; }
.legend {
  display: flex; gap: 16px; align-items: center; padding: 10px 20px;
  color: #9ca3af; font-size: 13px;
}
.dot {
  width: 10px; height: 10px; border-radius: 50%; display: inline-block; margin-right: 6px;
}
.dot.pending { background: #60a5fa; }
.dot.cooking { background: #f59e0b; }
.dot.done { background: #22c55e; }
.cards {
  flex: 1; padding: 12px 20px 24px;
  display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 14px;
  align-content: start;
}
.card {
  background: #1f2937; border-radius: 12px; padding: 14px;
  border-left: 5px solid #60a5fa;
}
.card.cooking { border-left-color: #f59e0b; }
.card.done { border-left-color: #22c55e; opacity: 0.75; }
.card-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
.order-no { font-size: 18px; font-weight: 700; }
.sub { font-size: 12px; color: #9ca3af; margin-top: 4px; }
.items { list-style: none; margin: 0; padding: 0; }
.items li {
  display: flex; gap: 8px; align-items: baseline;
  padding: 6px 0; border-bottom: 1px dashed #374151;
}
.iname { font-weight: 600; }
.ispec { color: #9ca3af; font-size: 12px; }
.iqty { margin-left: auto; color: #fbbf24; font-weight: 700; }
.card-actions { margin-top: 12px; display: flex; gap: 8px; }
.empty { grid-column: 1 / -1; text-align: center; color: #6b7280; padding: 60px 0; }
</style>
