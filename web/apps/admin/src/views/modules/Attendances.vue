<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openPunch">打卡</el-button>
      <el-button type="warning" @click="openMakeup">补卡</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="employee_id" label="员工ID" width="90" />
      <el-table-column prop="work_date" label="工作日" />
      <el-table-column prop="clock_in_at" label="上班" min-width="160" />
      <el-table-column prop="clock_out_at" label="下班" min-width="160" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="remark" label="备注" />
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="punchVisible" title="员工打卡" width="420px">
      <el-form label-width="90px">
        <el-form-item label="员工">
          <el-select v-model="punch.employee_id" filterable style="width:100%">
            <el-option v-for="e in employees" :key="e.id" :label="`${e.name}(${e.emp_no})`" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="punch.type">
            <el-radio value="in">上班</el-radio>
            <el-radio value="out">下班</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="punch.work_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="punch.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="punchVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="doPunch">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="makeupVisible" title="补卡" width="480px">
      <el-form label-width="90px">
        <el-form-item label="员工">
          <el-select v-model="makeup.employee_id" filterable style="width:100%">
            <el-option v-for="e in employees" :key="e.id" :label="`${e.name}(${e.emp_no})`" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="工作日"><el-date-picker v-model="makeup.work_date" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-form-item label="上班时间"><el-input v-model="makeup.clock_in_at" placeholder="2026-10-14 09:00:00" /></el-form-item>
        <el-form-item label="下班时间"><el-input v-model="makeup.clock_out_at" placeholder="2026-10-14 18:00:00" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="makeup.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="makeupVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="doMakeup">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiPage, apiPost } from '@openbows/shared'

const list = ref<any[]>([])
const employees = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const punchVisible = ref(false)
const makeupVisible = ref(false)
const punch = reactive<any>({ employee_id: undefined, type: 'in', work_date: new Date().toISOString().slice(0, 10), remark: '' })
const makeup = reactive<any>({
  employee_id: undefined,
  work_date: new Date().toISOString().slice(0, 10),
  clock_in_at: '',
  clock_out_at: '',
  remark: '',
})

async function loadEmployees() {
  const data = await apiPage<any>('/employees', { page: 1, page_size: 200 })
  employees.value = data.list || []
}

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/attendances', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openPunch() {
  punch.employee_id = employees.value[0]?.id
  punch.type = 'in'
  punch.work_date = new Date().toISOString().slice(0, 10)
  punch.remark = ''
  punchVisible.value = true
}

async function doPunch() {
  saving.value = true
  try {
    await apiPost('/attendances/punch', { ...punch })
    ElMessage.success('打卡成功')
    punchVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '打卡失败')
  } finally {
    saving.value = false
  }
}

function openMakeup() {
  makeup.employee_id = employees.value[0]?.id
  makeup.work_date = new Date().toISOString().slice(0, 10)
  makeup.clock_in_at = `${makeup.work_date} 09:00:00`
  makeup.clock_out_at = `${makeup.work_date} 18:00:00`
  makeup.remark = ''
  makeupVisible.value = true
}

async function doMakeup() {
  saving.value = true
  try {
    await apiPost('/attendances/makeup', { ...makeup })
    ElMessage.success('补卡成功')
    makeupVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '补卡失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadEmployees()
  await load()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
