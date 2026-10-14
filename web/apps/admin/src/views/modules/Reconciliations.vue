<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="visible = true">新建对账</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="business_day" label="营业日" />
      <el-table-column prop="channel" label="渠道" width="100" />
      <el-table-column prop="system_amount" label="系统金额" />
      <el-table-column prop="channel_amount" label="渠道金额" />
      <el-table-column prop="diff_amount" label="差异" />
      <el-table-column prop="status" label="状态" width="100" />
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" title="渠道对账" width="480px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="营业日">
          <el-date-picker v-model="form.business_day" type="date" value-format="YYYY-MM-DD" style="width:100%" />
        </el-form-item>
        <el-form-item label="渠道">
          <el-select v-model="form.channel" style="width:100%">
            <el-option label="现金" value="cash" />
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="其他" value="other" />
            <el-option label="会员" value="member" />
          </el-select>
        </el-form-item>
        <el-form-item label="渠道金额"><el-input v-model="form.channel_amount" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="create">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiPage, apiPost } from '@openbows/shared'

const list = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const form = reactive({
  business_day: new Date().toISOString().slice(0, 10),
  channel: 'wechat',
  channel_amount: '0',
  remark: '',
})

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/reconciliations', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function create() {
  saving.value = true
  try {
    await apiPost('/reconciliations', { ...form })
    ElMessage.success('对账完成')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '对账失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
