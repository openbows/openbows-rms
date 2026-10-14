<template>
  <div>
    <div class="toolbar">
      <el-button type="success" @click="openEdit()">新增角色</el-button>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="code" label="编码" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="openPerms(row)">权限</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
    <el-dialog v-model="visible" :title="form.id ? '编辑角色' : '新增角色'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="permVisible" title="分配权限" width="560px">
      <el-checkbox-group v-model="selectedPermIds">
        <el-checkbox v-for="p in allPerms" :key="p.id" :label="p.id" :value="p.id" style="width:45%;margin:4px 0">{{ p.name }} ({{ p.code }})</el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="permVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="savePerms">保存</el-button>
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
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const visible = ref(false)
const permVisible = ref(false)
const currentRoleId = ref(0)
const allPerms = ref<any[]>([])
const selectedPermIds = ref<number[]>([])
const form = reactive<any>({ id: 0, code: '', name: '', remark: '' })
async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/roles', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []; total.value = data.total || 0
  } catch (e: any) { ElMessage.error(e.message || '加载失败') }
  finally { loading.value = false }
}
function openEdit(row?: any) {
  Object.assign(form, row ? { id: row.id, code: row.code, name: row.name, remark: row.remark || '' } : { id: 0, code: '', name: '', remark: '' })
  visible.value = true
}
async function save() {
  saving.value = true
  try {
    const body = { code: form.code, name: form.name, remark: form.remark }
    if (form.id) await apiPut(`/roles/${form.id}`, body)
    else await apiPost('/roles', body)
    ElMessage.success('保存成功'); visible.value = false; load()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function openPerms(row: any) {
  currentRoleId.value = row.id
  const [perms, assigned] = await Promise.all([
    apiGet<any[]>('/permissions'),
    apiGet<any[]>(`/roles/${row.id}/permissions`),
  ])
  allPerms.value = perms || []
  selectedPermIds.value = (assigned || []).map((p: any) => p.id)
  permVisible.value = true
}
async function savePerms() {
  saving.value = true
  try {
    await apiPut(`/roles/${currentRoleId.value}/permissions`, { permission_ids: selectedPermIds.value })
    ElMessage.success('权限已更新'); permVisible.value = false
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' })
  try { await apiDelete(`/roles/${row.id}`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>
