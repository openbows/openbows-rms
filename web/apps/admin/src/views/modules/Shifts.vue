<template>
  <div>
    <div class="toolbar">
      <el-date-picker v-model="shiftDate" type="date" value-format="YYYY-MM-DD" placeholder="排班日期" @change="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增排班</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="employee_id" label="员工ID" width="90" />
      <el-table-column prop="shift_date" label="日期" />
      <el-table-column prop="shift_name" label="班次" />
      <el-table-column prop="start_time" label="开始" width="100" />
      <el-table-column prop="end_time" label="结束" width="100" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="row.status !== 'cancelled'" link type="warning" @click="cancel(row)">取消</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" :title="form.id ? '编辑排班' : '新增排班'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="员工">
          <el-select v-model="form.employee_id" filterable style="width:100%">
            <el-option v-for="e in employees" :key="e.id" :label="`${e.name}(${e.emp_no})`" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="form.shift_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="班次名"><el-input v-model="form.shift_name" placeholder="早班" /></el-form-item>
        <el-form-item label="开始"><el-input v-model="form.start_time" placeholder="09:00" /></el-form-item>
        <el-form-item label="结束"><el-input v-model="form.end_time" placeholder="17:00" /></el-form-item>
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
const employees = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const shiftDate = ref('')
const visible = ref(false)
const form = reactive<any>({
  id: 0, employee_id: undefined, shift_date: new Date().toISOString().slice(0, 10),
  start_time: '09:00', end_time: '17:00', shift_name: '早班', remark: '',
})

async function loadEmployees() {
  const data = await apiPage<any>('/employees', { page: 1, page_size: 200 })
  employees.value = data.list || []
}

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/shifts', {
      page: page.value, page_size: pageSize.value,
      shift_date: shiftDate.value || undefined,
    })
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
    Object.assign(form, {
      id: row.id, employee_id: row.employee_id, shift_date: String(row.shift_date).slice(0, 10),
      start_time: row.start_time, end_time: row.end_time, shift_name: row.shift_name || '', remark: row.remark || '',
    })
  } else {
    Object.assign(form, {
      id: 0, employee_id: employees.value[0]?.id, shift_date: new Date().toISOString().slice(0, 10),
      start_time: '09:00', end_time: '17:00', shift_name: '早班', remark: '',
    })
  }
  visible.value = true
}

async function save() {
  saving.value = true
  try {
    const body = {
      employee_id: form.employee_id, shift_date: form.shift_date,
      start_time: form.start_time, end_time: form.end_time,
      shift_name: form.shift_name, remark: form.remark,
    }
    if (form.id) await apiPut(`/shifts/${form.id}`, body)
    else await apiPost('/shifts', body)
    ElMessage.success('保存成功')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function cancel(row: any) {
  try {
    await apiPost(`/shifts/${row.id}/cancel`, {})
    ElMessage.success('已取消')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '取消失败')
  }
}

async function onDelete(row: any) {
  await ElMessageBox.confirm('确认删除排班？', '提示', { type: 'warning' })
  try {
    await apiDelete(`/shifts/${row.id}`)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(async () => {
  await loadEmployees()
  await load()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; align-items: center; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
