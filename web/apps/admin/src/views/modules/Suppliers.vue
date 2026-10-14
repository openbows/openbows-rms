<template>
  <div>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索" clearable style="width:200px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增供应商</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="code" label="编码"  />
      <el-table-column prop="name" label="名称"  />
      <el-table-column prop="contact_name" label="联系人"  />
      <el-table-column prop="phone" label="电话"  />
      <el-table-column label="启用" width="90"><template #default="{ row }">{{ row.status ? '是' : '否' }}</template></el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
    <el-dialog v-model="visible" :title="form.id ? '编辑供应商' : '新增供应商'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="联系人"><el-input v-model="form.contact_name" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.status" /></el-form-item>
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
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const form = reactive<any>({ id: 0, code: "", name: "", contact_name: "", phone: "", address: "", status: true, remark: "" })
async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/suppliers', { page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined })
    list.value = data.list || []; total.value = data.total || 0
  } catch (e: any) { ElMessage.error(e.message || '加载失败') }
  finally { loading.value = false }
}
function openEdit(row?: any) {
  if (row) Object.assign(form, { id: row.id, code: row.code ?? "", name: row.name ?? "", contact_name: row.contact_name ?? "", phone: row.phone ?? "", address: row.address ?? "", status: !!row.status, remark: row.remark ?? "" })
  else Object.assign(form, { id: 0, code: "", name: "", contact_name: "", phone: "", address: "", status: true, remark: "" })
  visible.value = true
}
async function save() {
  saving.value = true
  try {
    const { id, code, name, contact_name, phone, address, status, remark } = form
    const body = { code, name, contact_name, phone, address, status, remark }
    if (id) await apiPut(`/suppliers/${id}`, body)
    else await apiPost('/suppliers', body)
    ElMessage.success('保存成功'); visible.value = false; load()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  try { await apiDelete(`/suppliers/${row.id}`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>
