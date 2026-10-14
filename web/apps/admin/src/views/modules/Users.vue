<template>
  <div>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索用户名/姓名" clearable style="width:220px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="display_name" label="显示名" />
      <el-table-column prop="phone" label="手机" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">{{ row.status ? '启用' : '停用' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="openRoles(row)">角色</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
    <el-dialog v-model="visible" :title="form.id ? '编辑用户' : '新增用户'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="用户名" v-if="!form.id"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码" :required="!form.id"><el-input v-model="form.password" type="password" show-password :placeholder="form.id ? '留空不改' : ''" /></el-form-item>
        <el-form-item label="显示名"><el-input v-model="form.display_name" /></el-form-item>
        <el-form-item label="手机"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.status" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="roleVisible" title="分配角色" width="420px">
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox v-for="r in allRoles" :key="r.id" :label="r.id" :value="r.id">{{ r.name }} ({{ r.code }})</el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRoles">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost, apiPut, apiDelete, apiGet } from '@openbows/shared'
const list = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const roleVisible = ref(false)
const currentUserId = ref(0)
const allRoles = ref<any[]>([])
const selectedRoleIds = ref<number[]>([])
const form = reactive<any>({ id: 0, username: '', password: '', display_name: '', phone: '', status: true })
async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/users', { page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined })
    list.value = data.list || []; total.value = data.total || 0
  } catch (e: any) { ElMessage.error(e.message || '加载失败') }
  finally { loading.value = false }
}
function openEdit(row?: any) {
  Object.assign(form, row
    ? { id: row.id, username: row.username, password: '', display_name: row.display_name, phone: row.phone || '', status: !!row.status }
    : { id: 0, username: '', password: '', display_name: '', phone: '', status: true })
  visible.value = true
}
async function save() {
  saving.value = true
  try {
    if (form.id) {
      await apiPut(`/users/${form.id}`, { display_name: form.display_name, phone: form.phone || null, password: form.password || undefined, status: form.status })
    } else {
      await apiPost('/users', { username: form.username, password: form.password, display_name: form.display_name, phone: form.phone || null, status: form.status })
    }
    ElMessage.success('保存成功'); visible.value = false; load()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function openRoles(row: any) {
  currentUserId.value = row.id
  const [rolesPage, assigned] = await Promise.all([
    apiPage<any>('/roles', { page: 1, page_size: 200 }),
    apiGet<any[]>(`/users/${row.id}/roles`),
  ])
  allRoles.value = rolesPage.list || []
  selectedRoleIds.value = (assigned || []).map((r: any) => r.id)
  roleVisible.value = true
}
async function saveRoles() {
  saving.value = true
  try {
    await apiPut(`/users/${currentUserId.value}/roles`, { role_ids: selectedRoleIds.value })
    ElMessage.success('角色已更新'); roleVisible.value = false
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm(`确认删除用户「${row.username}」？`, '提示', { type: 'warning' })
  try { await apiDelete(`/users/${row.id}`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>
