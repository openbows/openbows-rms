<template>
  <div>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索手机/姓名" clearable style="width:200px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增会员</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="phone" label="手机" />
      <el-table-column prop="name" label="姓名" />
      <el-table-column prop="level" label="等级" width="100" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">{{ row.status ? '启用' : '停用' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="openRecharge(row)">充值</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" :title="form.id ? '编辑会员' : '新增会员'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="手机"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender" style="width:100%">
            <el-option label="未知" value="unknown" />
            <el-option label="男" value="male" />
            <el-option label="女" value="female" />
          </el-select>
        </el-form-item>
        <el-form-item label="生日"><el-input v-model="form.birthday" placeholder="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="等级"><el-input v-model="form.level" placeholder="normal" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.status" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="rechargeVisible" title="会员充值" width="400px">
      <el-form label-width="80px">
        <el-form-item label="金额"><el-input v-model="rechargeAmount" placeholder="100.00" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="rechargeRemark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="doRecharge">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost, apiPut, apiDelete } from '@openbows/shared'

const list = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const rechargeVisible = ref(false)
const rechargeId = ref(0)
const rechargeAmount = ref('100')
const rechargeRemark = ref('')
const form = reactive<any>({ id: 0, phone: '', name: '', gender: 'unknown', birthday: '', level: 'normal', status: true, remark: '' })

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/members', { page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined })
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
      id: row.id, phone: row.phone, name: row.name || '', gender: row.gender || 'unknown',
      birthday: row.birthday || '', level: row.level || 'normal', status: !!row.status, remark: row.remark || '',
    })
  } else {
    Object.assign(form, { id: 0, phone: '', name: '', gender: 'unknown', birthday: '', level: 'normal', status: true, remark: '' })
  }
  visible.value = true
}

async function save() {
  saving.value = true
  try {
    const body = {
      phone: form.phone, name: form.name, gender: form.gender, birthday: form.birthday,
      level: form.level, status: form.status, remark: form.remark,
    }
    if (form.id) await apiPut(`/members/${form.id}`, body)
    else await apiPost('/members', body)
    ElMessage.success('保存成功')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function openRecharge(row: any) {
  rechargeId.value = row.id
  rechargeAmount.value = '100'
  rechargeRemark.value = ''
  rechargeVisible.value = true
}

async function doRecharge() {
  saving.value = true
  try {
    await apiPost(`/members/${rechargeId.value}/recharge`, { amount: rechargeAmount.value, remark: rechargeRemark.value })
    ElMessage.success('充值成功')
    rechargeVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '充值失败')
  } finally {
    saving.value = false
  }
}

async function onDelete(row: any) {
  await ElMessageBox.confirm(`确认删除会员「${row.phone}」？`, '提示', { type: 'warning' })
  try {
    await apiDelete(`/members/${row.id}`)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
