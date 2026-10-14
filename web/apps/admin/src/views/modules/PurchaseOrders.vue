<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openCreate">新建采购单</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="order_no" label="单号" />
      <el-table-column prop="supplier_id" label="供应商ID" width="100" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button v-if="row.status === 'draft'" link type="primary" @click="submit(row)">提交</el-button>
          <el-button v-if="row.status === 'draft' || row.status === 'submitted'" link type="danger" @click="cancel(row)">取消</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" title="新建采购单" width="680px">
      <el-form label-width="90px">
        <el-form-item label="供应商">
          <el-select v-model="form.supplier_id" filterable style="width:100%">
            <el-option v-for="s in suppliers" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <div style="margin-bottom:8px">
        <el-button size="small" type="primary" @click="form.items.push({ material_id: undefined, qty: '1', unit_price: '0', remark: '' })">添加明细</el-button>
      </div>
      <el-table :data="form.items" border size="small">
        <el-table-column label="物料">
          <template #default="{ row }">
            <el-select v-model="row.material_id" filterable style="width:100%">
              <el-option v-for="m in materials" :key="m.id" :label="`${m.name}(${m.code})`" :value="m.id" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="数量" width="100"><template #default="{ row }"><el-input v-model="row.qty" /></template></el-table-column>
        <el-table-column label="单价" width="100"><template #default="{ row }"><el-input v-model="row.unit_price" /></template></el-table-column>
        <el-table-column width="60">
          <template #default="{ $index }"><el-button link type="danger" @click="form.items.splice($index,1)">删</el-button></template>
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
import { apiPage, apiPost } from '@openbows/shared'

const list = ref<any[]>([])
const suppliers = ref<any[]>([])
const materials = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const form = reactive<any>({ supplier_id: undefined, remark: '', items: [] as any[] })

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/purchase-orders', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  const [s, m] = await Promise.all([
    apiPage<any>('/suppliers', { page: 1, page_size: 200 }),
    apiPage<any>('/materials', { page: 1, page_size: 500 }),
  ])
  suppliers.value = s.list || []
  materials.value = m.list || []
  form.supplier_id = suppliers.value[0]?.id
  form.remark = ''
  form.items = [{ material_id: undefined, qty: '1', unit_price: '0', remark: '' }]
  visible.value = true
}

async function create() {
  saving.value = true
  try {
    await apiPost('/purchase-orders', {
      supplier_id: form.supplier_id,
      remark: form.remark,
      items: form.items.filter((i: any) => i.material_id).map((i: any) => ({
        material_id: i.material_id, qty: i.qty, unit_price: i.unit_price, remark: i.remark,
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

async function submit(row: any) {
  try {
    await apiPost(`/purchase-orders/${row.id}/submit`, {})
    ElMessage.success('已提交')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '提交失败')
  }
}

async function cancel(row: any) {
  await ElMessageBox.confirm('确认取消采购单？', '提示', { type: 'warning' })
  try {
    await apiPost(`/purchase-orders/${row.id}/cancel`, {})
    ElMessage.success('已取消')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '取消失败')
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
