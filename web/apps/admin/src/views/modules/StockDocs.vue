<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openCreate">新建单据</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="doc_no" label="单号" />
      <el-table-column prop="doc_type" label="类型" width="100" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-if="row.status === 'draft'" link type="primary" @click="postDoc(row)">过账</el-button>
          <el-button v-if="row.status !== 'voided'" link type="danger" @click="voidDoc(row)">作废</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" title="新建库存单据" width="680px">
      <el-form label-width="90px">
        <el-form-item label="类型">
          <el-select v-model="form.doc_type" style="width:200px">
            <el-option label="入库" value="inbound" />
            <el-option label="出库" value="outbound" />
            <el-option label="调整" value="adjust" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <div style="margin-bottom:8px">
        <el-button size="small" type="primary" @click="form.items.push({ material_id: undefined, qty: '1', unit_cost: '0', remark: '' })">添加明细</el-button>
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
        <el-table-column label="单位成本" width="110"><template #default="{ row }"><el-input v-model="row.unit_cost" /></template></el-table-column>
        <el-table-column label="备注" width="120"><template #default="{ row }"><el-input v-model="row.remark" /></template></el-table-column>
        <el-table-column width="60">
          <template #default="{ $index }"><el-button link type="danger" @click="form.items.splice($index,1)">删</el-button></template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="createDoc">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost } from '@openbows/shared'

const list = ref<any[]>([])
const materials = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const form = reactive<any>({ doc_type: 'inbound', remark: '', items: [] as any[] })

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/stock-docs', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  const m = await apiPage<any>('/materials', { page: 1, page_size: 500 })
  materials.value = m.list || []
  form.doc_type = 'inbound'
  form.remark = ''
  form.items = [{ material_id: undefined, qty: '1', unit_cost: '0', remark: '' }]
  visible.value = true
}

async function createDoc() {
  saving.value = true
  try {
    await apiPost('/stock-docs', {
      doc_type: form.doc_type,
      remark: form.remark,
      items: form.items.filter((i: any) => i.material_id).map((i: any) => ({
        material_id: i.material_id, qty: i.qty, unit_cost: i.unit_cost, remark: i.remark,
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

async function postDoc(row: any) {
  await ElMessageBox.confirm('确认过账？过账后将更新库存结存。', '提示', { type: 'warning' })
  try {
    await apiPost(`/stock-docs/${row.id}/post`, {})
    ElMessage.success('已过账')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '过账失败')
  }
}

async function voidDoc(row: any) {
  await ElMessageBox.confirm('确认作废该单据？', '提示', { type: 'warning' })
  try {
    await apiPost(`/stock-docs/${row.id}/void`, { reason: '管理端作废' })
    ElMessage.success('已作废')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '作废失败')
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
