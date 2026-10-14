<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="takeVisible = true">取号</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="ticket_no" label="排队号" width="100" />
      <el-table-column prop="guest_count" label="人数" width="70" />
      <el-table-column prop="phone" label="电话" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="business_day" label="营业日" />
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button v-if="row.status === 'waiting'" link type="primary" @click="call(row)">叫号</el-button>
          <el-button v-if="row.status === 'waiting' || row.status === 'called'" link type="success" @click="openSeat(row)">入座</el-button>
          <el-button v-if="row.status === 'waiting' || row.status === 'called'" link type="danger" @click="cancel(row)">取消</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="takeVisible" title="取号" width="400px">
      <el-form label-width="80px">
        <el-form-item label="人数"><el-input-number v-model="takeForm.guest_count" :min="1" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="takeForm.phone" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="takeForm.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="takeVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="take">确认取号</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="seatVisible" title="排队入座" width="400px">
      <el-form label-width="100px">
        <el-form-item label="桌台">
          <el-select v-model="seatForm.table_id" style="width:100%">
            <el-option v-for="t in idleTables" :key="t.id" :label="`${t.code} ${t.name || ''}`" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="创建订单"><el-switch v-model="seatForm.create_order" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="seatVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="doSeat">确认入座</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost } from '@openbows/shared'

const list = ref<any[]>([])
const tables = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const takeVisible = ref(false)
const seatVisible = ref(false)
const currentId = ref(0)
const takeForm = reactive({ guest_count: 2, phone: '', remark: '' })
const seatForm = reactive({ table_id: undefined as number | undefined, create_order: true })

const idleTables = computed(() => tables.value.filter((t) => t.status === 'idle'))

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/queues', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadTables() {
  const data = await apiPage<any>('/tables', { page: 1, page_size: 200 })
  tables.value = data.list || []
}

async function take() {
  saving.value = true
  try {
    await apiPost('/queues', { ...takeForm })
    ElMessage.success('取号成功')
    takeVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '取号失败')
  } finally {
    saving.value = false
  }
}

async function call(row: any) {
  try {
    await apiPost(`/queues/${row.id}/call`, {})
    ElMessage.success('已叫号')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '叫号失败')
  }
}

async function openSeat(row: any) {
  await loadTables()
  currentId.value = row.id
  seatForm.table_id = idleTables.value[0]?.id
  seatForm.create_order = true
  seatVisible.value = true
}

async function doSeat() {
  if (!seatForm.table_id) {
    ElMessage.warning('请选择桌台')
    return
  }
  saving.value = true
  try {
    await apiPost(`/queues/${currentId.value}/seat`, {
      table_id: seatForm.table_id,
      create_order: seatForm.create_order,
    })
    ElMessage.success('已入座')
    seatVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '入座失败')
  } finally {
    saving.value = false
  }
}

async function cancel(row: any) {
  await ElMessageBox.confirm('确认取消排队？', '提示', { type: 'warning' })
  try {
    await apiPost(`/queues/${row.id}/cancel`, { remark: '管理端取消' })
    ElMessage.success('已取消')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '取消失败')
  }
}

onMounted(async () => {
  await loadTables()
  await load()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
