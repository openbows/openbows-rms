<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openEdit()">新增桌台</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="code" label="编号" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="area_id" label="区域ID" width="90" />
      <el-table-column prop="capacity" label="人数" width="80" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="current_order_id" label="当前订单" width="100" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
    <el-dialog v-model="visible" :title="form.id ? '编辑桌台' : '新增桌台'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="区域">
          <el-select v-model="form.area_id" style="width:100%">
            <el-option v-for="a in areas" :key="a.id" :label="a.name" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="编号"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="容纳人数"><el-input-number v-model="form.capacity" :min="1" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="空闲" value="idle" />
            <el-option label="占用" value="occupied" />
            <el-option label="预留" value="reserved" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort_no" :min="0" /></el-form-item>
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
const areas = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const visible = ref(false)
const form = reactive<any>({ id: 0, area_id: undefined, code: '', name: '', capacity: 4, status: 'idle', sort_no: 0 })

async function loadAreas() {
  const data = await apiPage<any>('/areas', { page: 1, page_size: 200 })
  areas.value = data.list || []
}

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/tables', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openEdit(row?: any) {
  if (row) {
    Object.assign(form, { id: row.id, area_id: row.area_id, code: row.code, name: row.name, capacity: row.capacity, status: row.status, sort_no: row.sort_no })
  } else {
    Object.assign(form, { id: 0, area_id: areas.value[0]?.id, code: '', name: '', capacity: 4, status: 'idle', sort_no: 0 })
  }
  visible.value = true
}

async function save() {
  saving.value = true
  try {
    const body = { area_id: form.area_id, code: form.code, name: form.name, capacity: form.capacity, status: form.status, sort_no: form.sort_no }
    if (form.id) await apiPut(`/tables/${form.id}`, body)
    else await apiPost('/tables', body)
    ElMessage.success('保存成功')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function onDelete(row: any) {
  await ElMessageBox.confirm(`确认删除桌台「${row.code}」？`, '提示', { type: 'warning' })
  try {
    await apiDelete(`/tables/${row.id}`)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(async () => {
  await loadAreas()
  await load()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
