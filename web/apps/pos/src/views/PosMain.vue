<template>
  <div class="pos">
    <header class="topbar">
      <div class="brand">Openbows 收银</div>
      <div class="actions">
        <el-button @click="queueDrawer = true">排队</el-button>
        <el-button @click="resDrawer = true">预订</el-button>
        <el-button @click="refreshAll">刷新</el-button>
        <span class="user">{{ auth.displayName }}</span>
        <el-button type="danger" link @click="onLogout">退出</el-button>
      </div>
    </header>

    <div class="body">
      <!-- Left: tables -->
      <aside class="panel left">
        <div class="panel-title">桌台</div>
        <div class="table-grid">
          <div
            v-for="t in tables"
            :key="t.id"
            class="table-card"
            :class="[t.status, { active: selectedTable?.id === t.id }]"
            @click="onSelectTable(t)"
          >
            <div class="code">{{ t.code }}</div>
            <div class="name">{{ t.name || '桌台' }}</div>
            <div class="meta">{{ statusLabel(t.status) }} · {{ t.capacity }}人</div>
          </div>
        </div>
      </aside>

      <!-- Center: order -->
      <section class="panel center">
        <div class="panel-title">
          <span>当前订单</span>
          <span v-if="order" class="order-meta">{{ order.order_no }} · {{ orderStatusLabel(order.status) }}</span>
        </div>
        <div v-if="!selectedTable" class="empty">请选择左侧桌台</div>
        <template v-else>
          <div class="table-info">
            桌台 {{ selectedTable.code }}
            <el-tag size="small" style="margin-left:8px">{{ statusLabel(selectedTable.status) }}</el-tag>
            <el-button v-if="!order" size="small" type="primary" style="margin-left:12px" @click="createOrderDialog = true">开台下单</el-button>
          </div>
          <el-table v-if="order" :data="orderItems" size="small" height="360" border>
            <el-table-column prop="item.dish_name" label="菜品" />
            <el-table-column prop="item.spec_name" label="规格" width="90" />
            <el-table-column prop="item.qty" label="数量" width="70" />
            <el-table-column prop="item.amount" label="金额" width="90" />
            <el-table-column prop="item.status" label="状态" width="80" />
            <el-table-column label="操作" width="90">
              <template #default="{ row }">
                <el-button
                  v-if="row.item.status === 'normal' && order.status !== 'completed' && order.status !== 'cancelled'"
                  link type="danger" @click="returnItem(row.item)"
                >退菜</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="order" class="order-footer">
            <div class="amounts">
              小计 {{ order.subtotal_amount }} · 优惠 {{ order.discount_amount }} · 应付 <b>{{ order.payable_amount }}</b>
            </div>
            <div class="btns">
              <el-button v-if="canAdvance" type="warning" @click="advanceOrder">推进状态</el-button>
              <el-button v-if="order.status === 'pending_pay'" type="success" @click="openPay">结账</el-button>
              <el-button v-if="order.status !== 'completed' && order.status !== 'cancelled'" type="danger" plain @click="cancelOrder">作废</el-button>
            </div>
          </div>
        </template>
      </section>

      <!-- Right: menu -->
      <aside class="panel right">
        <div class="panel-title">点菜</div>
        <el-tabs v-model="activeCategory" @tab-change="() => {}">
          <el-tab-pane label="全部" name="all" />
          <el-tab-pane v-for="c in categories" :key="c.id" :label="c.name" :name="String(c.id)" />
        </el-tabs>
        <div class="dish-grid">
          <div
            v-for="d in filteredDishes"
            :key="d.id"
            class="dish-card"
            :class="{ disabled: !order || order.status === 'completed' || order.status === 'cancelled' || order.status === 'pending_pay' }"
            @click="addDish(d)"
          >
            <div class="dname">{{ d.name }}</div>
            <div class="dprice">¥{{ d.base_price }}</div>
          </div>
        </div>
      </aside>
    </div>

    <!-- Create order -->
    <el-dialog v-model="createOrderDialog" title="开台下单" width="420px">
      <el-form label-width="90px">
        <el-form-item label="人数"><el-input-number v-model="createForm.guest_count" :min="1" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="createForm.remark" /></el-form-item>
        <el-form-item label="首道菜">
          <el-select v-model="createForm.dish_id" filterable style="width:100%">
            <el-option v-for="d in dishes" :key="d.id" :label="`${d.name} ¥${d.base_price}`" :value="d.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createOrderDialog=false">取消</el-button>
        <el-button type="primary" :loading="busy" @click="createOrder">确认开台</el-button>
      </template>
    </el-dialog>

    <!-- Pay -->
    <el-dialog v-model="payVisible" title="结账收款" width="440px">
      <el-form label-width="90px">
        <el-form-item label="应付"><el-input :model-value="order?.payable_amount" disabled /></el-form-item>
        <el-form-item label="渠道">
          <el-select v-model="payForm.channel" style="width:100%">
            <el-option label="现金" value="cash" />
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="其他" value="other" />
            <el-option label="会员余额" value="member" />
          </el-select>
        </el-form-item>
        <el-form-item label="金额"><el-input v-model="payForm.amount" /></el-form-item>
        <el-form-item v-if="payForm.channel === 'member'" label="会员ID">
          <el-input-number v-model="payForm.member_id" :min="1" style="width:100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="payForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="payVisible=false">取消</el-button>
        <el-button type="success" :loading="busy" @click="doPay">确认收款</el-button>
      </template>
    </el-dialog>

    <!-- Queue drawer -->
    <el-drawer v-model="queueDrawer" title="排队叫号" size="420px">
      <div style="margin-bottom:12px;display:flex;gap:8px">
        <el-button type="success" size="small" @click="takeQueue">取号</el-button>
        <el-button size="small" @click="loadQueues">刷新</el-button>
      </div>
      <el-table :data="queues" size="small" border>
        <el-table-column prop="ticket_no" label="号" width="60" />
        <el-table-column prop="guest_count" label="人" width="50" />
        <el-table-column prop="status" label="状态" width="80" />
        <el-table-column label="操作" min-width="160">
          <template #default="{ row }">
            <el-button v-if="row.status==='waiting'" link type="primary" @click="callQueue(row)">叫号</el-button>
            <el-button v-if="row.status==='waiting'||row.status==='called'" link type="success" @click="seatQueue(row)">入座</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>

    <!-- Reservation drawer -->
    <el-drawer v-model="resDrawer" title="预订管理" size="480px">
      <div style="margin-bottom:12px">
        <el-button size="small" @click="loadReservations">刷新</el-button>
      </div>
      <el-table :data="reservations" size="small" border>
        <el-table-column prop="customer_name" label="客户" />
        <el-table-column prop="guest_count" label="人" width="50" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button v-if="row.status==='pending'" link type="primary" @click="confirmRes(row)">确认</el-button>
            <el-button v-if="row.status==='pending'||row.status==='confirmed'" link type="success" @click="seatRes(row)">入座</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiGet, apiPost } from '@openbows/shared'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const tables = ref<any[]>([])
const categories = ref<any[]>([])
const dishes = ref<any[]>([])
const selectedTable = ref<any>(null)
const order = ref<any>(null)
const orderItems = ref<any[]>([])
const activeCategory = ref('all')
const busy = ref(false)
const createOrderDialog = ref(false)
const payVisible = ref(false)
const queueDrawer = ref(false)
const resDrawer = ref(false)
const queues = ref<any[]>([])
const reservations = ref<any[]>([])

const createForm = reactive({ guest_count: 2, remark: '', dish_id: undefined as number | undefined })
const payForm = reactive({ channel: 'cash', amount: '0', member_id: undefined as number | undefined, remark: '' })

const filteredDishes = computed(() => {
  if (activeCategory.value === 'all') return dishes.value.filter((d) => d.is_on_sale)
  return dishes.value.filter((d) => d.is_on_sale && String(d.category_id) === activeCategory.value)
})

const canAdvance = computed(() => {
  const s = order.value?.status
  return s === 'pending_confirm' || s === 'cooking'
})

function statusLabel(s: string) {
  return ({ idle: '空闲', occupied: '占用', reserved: '预留' } as any)[s] || s
}
function orderStatusLabel(s: string) {
  return ({
    pending_confirm: '待确认', cooking: '制作中', pending_pay: '待结账',
    completed: '已完成', cancelled: '已取消',
  } as any)[s] || s
}

async function refreshAll() {
  await Promise.all([loadTables(), loadMenu()])
  if (selectedTable.value) {
    const t = tables.value.find((x) => x.id === selectedTable.value.id)
    if (t) await onSelectTable(t)
  }
}

async function loadTables() {
  const data = await apiPage<any>('/tables', { page: 1, page_size: 200 })
  tables.value = data.list || []
}

async function loadMenu() {
  const [c, d] = await Promise.all([
    apiPage<any>('/menu-categories', { page: 1, page_size: 100 }),
    apiPage<any>('/dishes', { page: 1, page_size: 500 }),
  ])
  categories.value = c.list || []
  dishes.value = d.list || []
  if (!createForm.dish_id && dishes.value.length) createForm.dish_id = dishes.value[0].id
}

async function onSelectTable(t: any) {
  selectedTable.value = t
  order.value = null
  orderItems.value = []
  if (t.current_order_id) {
    await loadOrder(t.current_order_id)
  }
}

async function loadOrder(id: number) {
  const detail = await apiGet<any>(`/orders/${id}`)
  order.value = detail.order
  orderItems.value = detail.items || []
}

async function createOrder() {
  if (!selectedTable.value || !createForm.dish_id) {
    ElMessage.warning('请选择桌台和首道菜')
    return
  }
  busy.value = true
  try {
    const detail = await apiPost<any>('/orders', {
      table_id: selectedTable.value.id,
      guest_count: createForm.guest_count,
      remark: createForm.remark,
      channel: 'dine_in',
      items: [{ dish_id: createForm.dish_id, qty: 1 }],
    })
    order.value = detail.order
    orderItems.value = detail.items || []
    createOrderDialog.value = false
    ElMessage.success('开台成功')
    await loadTables()
    selectedTable.value = tables.value.find((x) => x.id === selectedTable.value.id) || selectedTable.value
  } catch (e: any) {
    ElMessage.error(e.message || '开台失败')
  } finally {
    busy.value = false
  }
}

async function addDish(d: any) {
  if (!order.value || order.value.status === 'completed' || order.value.status === 'cancelled' || order.value.status === 'pending_pay') return
  busy.value = true
  try {
    const detail = await apiPost<any>(`/orders/${order.value.id}/items`, {
      items: [{ dish_id: d.id, qty: 1 }],
    })
    order.value = detail.order
    orderItems.value = detail.items || []
  } catch (e: any) {
    ElMessage.error(e.message || '加菜失败')
  } finally {
    busy.value = false
  }
}

async function returnItem(item: any) {
  await ElMessageBox.confirm(`确认退菜「${item.dish_name}」？`, '退菜', { type: 'warning' })
  try {
    const detail = await apiPost<any>(`/orders/${order.value.id}/items/${item.id}/return`, { remark: 'POS退菜' })
    order.value = detail.order
    orderItems.value = detail.items || []
    ElMessage.success('已退菜')
  } catch (e: any) {
    ElMessage.error(e.message || '退菜失败')
  }
}

async function advanceOrder() {
  try {
    const detail = await apiPost<any>(`/orders/${order.value.id}/advance`, { remark: 'POS推进' })
    order.value = detail.order
    orderItems.value = detail.items || []
    ElMessage.success('状态已更新')
  } catch (e: any) {
    ElMessage.error(e.message || '推进失败')
  }
}

async function cancelOrder() {
  await ElMessageBox.confirm('确认作废订单？', '提示', { type: 'warning' })
  try {
    await apiPost(`/orders/${order.value.id}/cancel`, { remark: 'POS作废' })
    ElMessage.success('订单已作废')
    order.value = null
    orderItems.value = []
    await loadTables()
  } catch (e: any) {
    ElMessage.error(e.message || '作废失败')
  }
}

function openPay() {
  payForm.channel = 'cash'
  payForm.amount = String(order.value?.payable_amount ?? '0')
  payForm.member_id = undefined
  payForm.remark = ''
  payVisible.value = true
}

async function doPay() {
  busy.value = true
  try {
    const body: any = {
      channel: payForm.channel,
      amount: payForm.amount,
      remark: payForm.remark,
    }
    if (payForm.channel === 'member') body.member_id = payForm.member_id
    await apiPost(`/orders/${order.value.id}/pay`, body)
    ElMessage.success('收款成功')
    payVisible.value = false
    order.value = null
    orderItems.value = []
    await loadTables()
    selectedTable.value = tables.value.find((x) => x.id === selectedTable.value?.id) || null
  } catch (e: any) {
    ElMessage.error(e.message || '收款失败')
  } finally {
    busy.value = false
  }
}

async function loadQueues() {
  const data = await apiPage<any>('/queues', { page: 1, page_size: 50 })
  queues.value = data.list || []
}

async function takeQueue() {
  try {
    await apiPost('/queues', { guest_count: 2 })
    ElMessage.success('取号成功')
    loadQueues()
  } catch (e: any) {
    ElMessage.error(e.message || '取号失败')
  }
}

async function callQueue(row: any) {
  try {
    await apiPost(`/queues/${row.id}/call`, {})
    ElMessage.success('已叫号')
    loadQueues()
  } catch (e: any) {
    ElMessage.error(e.message || '叫号失败')
  }
}

async function seatQueue(row: any) {
  const idle = tables.value.filter((t) => t.status === 'idle')
  if (!idle.length) {
    ElMessage.warning('暂无空闲桌台')
    return
  }
  try {
    await apiPost(`/queues/${row.id}/seat`, { table_id: idle[0].id, create_order: true })
    ElMessage.success(`已入座桌台 ${idle[0].code}`)
    loadQueues()
    await refreshAll()
  } catch (e: any) {
    ElMessage.error(e.message || '入座失败')
  }
}

async function loadReservations() {
  const data = await apiPage<any>('/reservations', { page: 1, page_size: 50 })
  reservations.value = data.list || []
}

async function confirmRes(row: any) {
  try {
    await apiPost(`/reservations/${row.id}/confirm`, {})
    ElMessage.success('已确认')
    loadReservations()
  } catch (e: any) {
    ElMessage.error(e.message || '确认失败')
  }
}

async function seatRes(row: any) {
  try {
    await apiPost(`/reservations/${row.id}/seat`, { table_id: row.table_id || null, create_order: true })
    ElMessage.success('预订已入座')
    loadReservations()
    await refreshAll()
  } catch (e: any) {
    ElMessage.error(e.message || '入座失败')
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
  await refreshAll()
  await Promise.all([loadQueues(), loadReservations()])
})
</script>

<style scoped>
.pos { height: 100vh; display: flex; flex-direction: column; }
.topbar {
  height: 52px; display: flex; align-items: center; justify-content: space-between;
  padding: 0 16px; background: #0f766e; color: #fff;
}
.brand { font-weight: 700; font-size: 18px; }
.actions { display: flex; align-items: center; gap: 8px; }
.actions :deep(.el-button) { --el-button-text-color: #fff; }
.user { margin-left: 8px; opacity: 0.9; }
.body { flex: 1; display: grid; grid-template-columns: 260px 1fr 300px; gap: 8px; padding: 8px; overflow: hidden; }
.panel { background: #fff; border-radius: 8px; padding: 12px; overflow: auto; display: flex; flex-direction: column; }
.panel-title { font-weight: 600; margin-bottom: 10px; display: flex; justify-content: space-between; align-items: center; }
.order-meta { font-size: 12px; color: #909399; font-weight: 400; }
.table-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.table-card {
  border: 2px solid #e5e7eb; border-radius: 8px; padding: 10px; cursor: pointer;
  transition: all .15s;
}
.table-card.idle { background: #ecfdf5; border-color: #6ee7b7; }
.table-card.occupied { background: #fef2f2; border-color: #fca5a5; }
.table-card.reserved { background: #fff7ed; border-color: #fdba74; }
.table-card.active { box-shadow: 0 0 0 2px #0f766e; }
.table-card .code { font-weight: 700; font-size: 16px; }
.table-card .meta { font-size: 12px; color: #6b7280; margin-top: 4px; }
.empty { color: #9ca3af; text-align: center; margin-top: 80px; }
.table-info { margin-bottom: 8px; }
.order-footer { margin-top: 12px; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.amounts b { color: #dc2626; font-size: 18px; }
.dish-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; overflow: auto; }
.dish-card {
  border: 1px solid #e5e7eb; border-radius: 8px; padding: 10px; cursor: pointer;
  background: #f8fafc;
}
.dish-card:hover { border-color: #0f766e; }
.dish-card.disabled { opacity: 0.45; pointer-events: none; }
.dname { font-weight: 600; font-size: 13px; }
.dprice { color: #0f766e; margin-top: 4px; }
</style>
