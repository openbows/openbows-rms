<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openCreate">新建入库单</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="receipt_no" label="入库单号" />
      <el-table-column prop="purchase_order_id" label="采购单ID" width="100" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button v-if="row.status === 'draft'" link type="primary" @click="post(row)">过账入库</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" title="新建采购入库" width="720px">
      <el-form label-width="100px">
        <el-form-item label="采购单">
          <el-select v-model="form.purchase_order_id" filterable style="width:100%" @change="onOrderChange">
            <el-option v-for="o in orders" :key="o.id" :label="`${o.order_no || o.id} (#${o.id})`" :value="o.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <el-table :data="form.items" border size="small">
        <el-table-column prop="purchase_order_item_id" label="采购明细ID" width="120" />
        <el-table-column label="入库数量" width="140">
          <template #default="{ row }"><el-input v-model="row.qty" /></template>
        </el-table-column>
        <el-table-column label="备注">
          <template #default="{ row }"><el-input v-model="row.remark" /></template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="create">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost, apiGet } from '@openbows/shared'

const list = ref<any[]>([])
const orders = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const form = reactive<any>({ purchase_order_id: undefined, remark: '', items: [] as any[] })

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/purchase-receipts', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  const data = await apiPage<any>('/purchase-orders', { page: 1, page_size: 200 })
  orders.value = (data.list || []).filter((o: any) => o.status === 'submitted' || o.status === 'partial' || o.status === 'draft')
  form.purchase_order_id = orders.value[0]?.id
  form.remark = ''
  form.items = []
  if (form.purchase_order_id) await onOrderChange(form.purchase_order_id)
  visible.value = true
}

async function onOrderChange(id: number) {
  if (!id) return
  const detail = await apiGet<any>(`/purchase-orders/${id}`)
  form.items = (detail.items || []).map((it: any) => ({
    purchase_order_item_id: it.id,
    qty: String(it.qty),
    remark: '',
  }))
}

async function create() {
  saving.value = true
  try {
    await apiPost('/purchase-receipts', {
      purchase_order_id: form.purchase_order_id,
      remark: form.remark,
      items: form.items.map((i: any) => ({
        purchase_order_item_id: i.purchase_order_item_id,
        qty: i.qty,
        remark: i.remark,
      })),
    })
    ElMessage.success('已创建')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  } finally {
    saving.value = false
  }
}

async function post(row: any) {
  await ElMessageBox.confirm('确认过账入库？将同步增加库存。', '提示', { type: 'warning' })
  try {
    await apiPost(`/purchase-receipts/${row.id}/post`, {})
    ElMessage.success('已过账')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '过账失败')
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
