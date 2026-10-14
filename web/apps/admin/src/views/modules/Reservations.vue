<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openEdit()">新增预订</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="customer_name" label="客户" />
      <el-table-column prop="phone" label="电话" />
      <el-table-column prop="guest_count" label="人数" width="70" />
      <el-table-column prop="reserve_at" label="预订时间" min-width="160" />
      <el-table-column prop="channel" label="渠道" width="90" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column label="操作" width="260">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="row.status === 'pending'" link type="success" @click="confirm(row)">确认</el-button>
          <el-button v-if="row.status === 'pending' || row.status === 'confirmed'" link type="warning" @click="seat(row)">入座</el-button>
          <el-button link type="danger" @click="cancel(row)">取消</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" :title="form.id ? '编辑预订' : '新增预订'" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="客户名"><el-input v-model="form.customer_name" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="人数"><el-input-number v-model="form.guest_count" :min="1" /></el-form-item>
        <el-form-item label="预订时间">
          <el-date-picker v-model="form.reserve_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ssZ" style="width:100%" />
        </el-form-item>
        <el-form-item label="桌台">
          <el-select v-model="form.table_id" clearable style="width:100%">
            <el-option v-for="t in tables" :key="t.id" :label="`${t.code} ${t.name || ''}`" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="渠道">
          <el-select v-model="form.channel" style="width:100%">
            <el-option label="电话" value="phone" />
            <el-option label="线上" value="online" />
            <el-option label="到店" value="walk_in" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="定金"><el-input v-model="form.deposit_amount" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost, apiPut, apiDelete } from '@openbows/shared'

const list = ref<any[]>([])
const tables = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const form = reactive<any>({
  id: 0, customer_name: '', phone: '', guest_count: 2, reserve_at: '',
  table_id: undefined, channel: 'phone', deposit_amount: '0', remark: '',
})

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/reservations', { page: page.value, page_size: pageSize.value })
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

function openEdit(row?: any) {
  if (row) {
    Object.assign(form, {
      id: row.id, customer_name: row.customer_name, phone: row.phone || '', guest_count: row.guest_count,
      reserve_at: row.reserve_at, table_id: row.table_id || undefined, channel: row.channel || 'phone',
      deposit_amount: String(row.deposit_amount ?? '0'), remark: row.remark || '',
    })
  } else {
    const d = new Date()
    d.setHours(d.getHours() + 2)
    Object.assign(form, {
      id: 0, customer_name: '', phone: '', guest_count: 2, reserve_at: d.toISOString(),
      table_id: undefined, channel: 'phone', deposit_amount: '0', remark: '',
    })
  }
  visible.value = true
}

async function save() {
  saving.value = true
  try {
    const body = {
      customer_name: form.customer_name, phone: form.phone, guest_count: form.guest_count,
      reserve_at: form.reserve_at, table_id: form.table_id || null, channel: form.channel,
      deposit_amount: form.deposit_amount, remark: form.remark,
    }
    if (form.id) await apiPut(`/reservations/${form.id}`, body)
    else await apiPost('/reservations', body)
    ElMessage.success('保存成功')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function confirm(row: any) {
  try {
    await apiPost(`/reservations/${row.id}/confirm`, {})
    ElMessage.success('已确认')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

async function seat(row: any) {
  try {
    await apiPost(`/reservations/${row.id}/seat`, { table_id: row.table_id || null, create_order: true })
    ElMessage.success('已入座')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '入座失败')
  }
}

async function cancel(row: any) {
  await ElMessageBox.confirm('确认取消预订？', '提示', { type: 'warning' })
  try {
    await apiPost(`/reservations/${row.id}/cancel`, { reason: '管理端取消' })
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
