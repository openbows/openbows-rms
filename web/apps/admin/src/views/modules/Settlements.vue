<template>
  <div>
    <div class="toolbar">
      <el-date-picker v-model="businessDay" type="date" value-format="YYYY-MM-DD" placeholder="营业日" />
      <el-button type="success" :loading="saving" @click="dailySettle">执行日结</el-button>
      <el-button type="primary" @click="load">刷新列表</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="business_day" label="营业日" />
      <el-table-column prop="total_amount" label="合计金额" />
      <el-table-column prop="order_count" label="订单数" width="100" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="remark" label="备注" />
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost } from '@openbows/shared'

const list = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const businessDay = ref(new Date().toISOString().slice(0, 10))

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/settlements', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function dailySettle() {
  await ElMessageBox.confirm(`确认对营业日 ${businessDay.value} 执行日结？`, '日结确认', { type: 'warning' })
  saving.value = true
  try {
    await apiPost('/settlements/daily', { business_day: businessDay.value, remark: '管理端日结' })
    ElMessage.success('日结成功')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '日结失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; align-items: center; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
