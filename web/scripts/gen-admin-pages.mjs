import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const dir = path.join(__dirname, '../apps/admin/src/views/modules')
fs.mkdirSync(dir, { recursive: true })

function write(name, content) {
  fs.writeFileSync(path.join(dir, name), content.trim() + '\n', 'utf8')
  console.log('wrote', name)
}

// --- Stores ---
write('Stores.vue', `<template>
  <div>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索编码/名称" clearable style="width:220px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="code" label="编码" />
      <el-table-column prop="name" label="名称" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">{{ row.status ? '启用' : '停用' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
    <el-dialog v-model="visible" :title="form.id ? '编辑门店' : '新增门店'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.status" /></el-form-item>
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
const form = reactive<any>({ id: 0, code: '', name: '', status: true })
async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/stores', { page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) { ElMessage.error(e.message || '加载失败') }
  finally { loading.value = false }
}
function openEdit(row?: any) {
  Object.assign(form, row ? { id: row.id, code: row.code, name: row.name, status: !!row.status } : { id: 0, code: '', name: '', status: true })
  visible.value = true
}
async function save() {
  saving.value = true
  try {
    const body = { code: form.code, name: form.name, status: form.status }
    if (form.id) await apiPut(\`/stores/\${form.id}\`, body)
    else await apiPost('/stores', body)
    ElMessage.success('保存成功')
    visible.value = false
    load()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm(\`确认删除门店「\${row.name}」？\`, '提示', { type: 'warning' })
  try { await apiDelete(\`/stores/\${row.id}\`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>`)

// --- Users ---
write('Users.vue', `<template>
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
      await apiPut(\`/users/\${form.id}\`, { display_name: form.display_name, phone: form.phone || null, password: form.password || undefined, status: form.status })
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
    apiGet<any[]>(\`/users/\${row.id}/roles\`),
  ])
  allRoles.value = rolesPage.list || []
  selectedRoleIds.value = (assigned || []).map((r: any) => r.id)
  roleVisible.value = true
}
async function saveRoles() {
  saving.value = true
  try {
    await apiPut(\`/users/\${currentUserId.value}/roles\`, { role_ids: selectedRoleIds.value })
    ElMessage.success('角色已更新'); roleVisible.value = false
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm(\`确认删除用户「\${row.username}」？\`, '提示', { type: 'warning' })
  try { await apiDelete(\`/users/\${row.id}\`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>`)

// --- Roles ---
write('Roles.vue', `<template>
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
    if (form.id) await apiPut(\`/roles/\${form.id}\`, body)
    else await apiPost('/roles', body)
    ElMessage.success('保存成功'); visible.value = false; load()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function openPerms(row: any) {
  currentRoleId.value = row.id
  const [perms, assigned] = await Promise.all([
    apiGet<any[]>('/permissions'),
    apiGet<any[]>(\`/roles/\${row.id}/permissions\`),
  ])
  allPerms.value = perms || []
  selectedPermIds.value = (assigned || []).map((p: any) => p.id)
  permVisible.value = true
}
async function savePerms() {
  saving.value = true
  try {
    await apiPut(\`/roles/\${currentRoleId.value}/permissions\`, { permission_ids: selectedPermIds.value })
    ElMessage.success('权限已更新'); permVisible.value = false
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm(\`确认删除角色「\${row.name}」？\`, '提示', { type: 'warning' })
  try { await apiDelete(\`/roles/\${row.id}\`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>`)

function simpleCrud({ file, title, api, fields, columns }) {
  const formFields = fields.map(f => {
    if (f.type === 'switch') return `<el-form-item label="${f.label}"><el-switch v-model="form.${f.key}" /></el-form-item>`
    if (f.type === 'number') return `<el-form-item label="${f.label}"><el-input-number v-model="form.${f.key}" :min="0" /></el-form-item>`
    return `<el-form-item label="${f.label}"><el-input v-model="form.${f.key}" /></el-form-item>`
  }).join('\n        ')
  const colTpl = columns.map(c => {
    if (c.bool) return `<el-table-column label="${c.label}" width="${c.width || 90}"><template #default="{ row }">{{ row.${c.key} ? '是' : '否' }}</template></el-table-column>`
    return `<el-table-column prop="${c.key}" label="${c.label}" ${c.width ? `width="${c.width}"` : ''} />`
  }).join('\n      ')
  const defaults = fields.map(f => `${f.key}: ${JSON.stringify(f.default ?? (f.type === 'switch' ? true : f.type === 'number' ? 0 : ''))}`).join(', ')
  const fromRow = fields.map(f => {
    if (f.type === 'switch') return `${f.key}: !!row.${f.key}`
    return `${f.key}: row.${f.key} ?? ${JSON.stringify(f.default ?? (f.type === 'number' ? 0 : ''))}`
  }).join(', ')
  const bodyKeys = fields.map(f => f.key).join(', ')

  write(file, `<template>
  <div>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索" clearable style="width:200px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增${title}</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      ${colTpl}
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
    <el-dialog v-model="visible" :title="form.id ? '编辑${title}' : '新增${title}'" width="480px">
      <el-form :model="form" label-width="90px">
        ${formFields}
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
const form = reactive<any>({ id: 0, ${defaults} })
async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('${api}', { page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined })
    list.value = data.list || []; total.value = data.total || 0
  } catch (e: any) { ElMessage.error(e.message || '加载失败') }
  finally { loading.value = false }
}
function openEdit(row?: any) {
  if (row) Object.assign(form, { id: row.id, ${fromRow} })
  else Object.assign(form, { id: 0, ${defaults} })
  visible.value = true
}
async function save() {
  saving.value = true
  try {
    const { id, ${bodyKeys} } = form
    const body = { ${bodyKeys} }
    if (id) await apiPut(\`${api}/\${id}\`, body)
    else await apiPost('${api}', body)
    ElMessage.success('保存成功'); visible.value = false; load()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}
async function onDelete(row: any) {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  try { await apiDelete(\`${api}/\${row.id}\`); ElMessage.success('已删除'); load() }
  catch (e: any) { ElMessage.error(e.message || '删除失败') }
}
onMounted(load)
</script>
<style scoped>
.toolbar { display:flex; gap:8px; margin-bottom:12px; }
.pager { margin-top:12px; justify-content:flex-end; }
</style>`)
}

simpleCrud({
  file: 'Stations.vue', title: '档口', api: '/stations',
  fields: [
    { key: 'code', label: '编码' },
    { key: 'name', label: '名称' },
    { key: 'sort_no', label: '排序', type: 'number', default: 0 },
    { key: 'status', label: '启用', type: 'switch', default: true },
  ],
  columns: [
    { key: 'code', label: '编码' },
    { key: 'name', label: '名称' },
    { key: 'sort_no', label: '排序', width: 80 },
    { key: 'status', label: '启用', bool: true },
  ],
})

simpleCrud({
  file: 'Categories.vue', title: '分类', api: '/menu-categories',
  fields: [
    { key: 'name', label: '名称' },
    { key: 'sort_no', label: '排序', type: 'number', default: 0 },
    { key: 'status', label: '启用', type: 'switch', default: true },
  ],
  columns: [
    { key: 'name', label: '名称' },
    { key: 'sort_no', label: '排序', width: 80 },
    { key: 'status', label: '启用', bool: true },
  ],
})

simpleCrud({
  file: 'Areas.vue', title: '区域', api: '/areas',
  fields: [
    { key: 'name', label: '名称' },
    { key: 'sort_no', label: '排序', type: 'number', default: 0 },
    { key: 'status', label: '启用', type: 'switch', default: true },
  ],
  columns: [
    { key: 'name', label: '名称' },
    { key: 'sort_no', label: '排序', width: 80 },
    { key: 'status', label: '启用', bool: true },
  ],
})

simpleCrud({
  file: 'Materials.vue', title: '物料', api: '/materials',
  fields: [
    { key: 'code', label: '编码' },
    { key: 'name', label: '名称' },
    { key: 'unit', label: '单位', default: '个' },
    { key: 'category', label: '分类' },
    { key: 'status', label: '启用', type: 'switch', default: true },
    { key: 'remark', label: '备注' },
  ],
  columns: [
    { key: 'code', label: '编码' },
    { key: 'name', label: '名称' },
    { key: 'unit', label: '单位', width: 80 },
    { key: 'category', label: '分类' },
    { key: 'status', label: '启用', bool: true },
  ],
})

simpleCrud({
  file: 'Suppliers.vue', title: '供应商', api: '/suppliers',
  fields: [
    { key: 'code', label: '编码' },
    { key: 'name', label: '名称' },
    { key: 'contact_name', label: '联系人' },
    { key: 'phone', label: '电话' },
    { key: 'address', label: '地址' },
    { key: 'status', label: '启用', type: 'switch', default: true },
    { key: 'remark', label: '备注' },
  ],
  columns: [
    { key: 'code', label: '编码' },
    { key: 'name', label: '名称' },
    { key: 'contact_name', label: '联系人' },
    { key: 'phone', label: '电话' },
    { key: 'status', label: '启用', bool: true },
  ],
})

simpleCrud({
  file: 'Employees.vue', title: '员工', api: '/employees',
  fields: [
    { key: 'emp_no', label: '工号' },
    { key: 'name', label: '姓名' },
    { key: 'phone', label: '电话' },
    { key: 'job_title', label: '岗位' },
    { key: 'hired_at', label: '入职日' },
    { key: 'status', label: '启用', type: 'switch', default: true },
    { key: 'remark', label: '备注' },
  ],
  columns: [
    { key: 'emp_no', label: '工号' },
    { key: 'name', label: '姓名' },
    { key: 'phone', label: '电话' },
    { key: 'job_title', label: '岗位' },
    { key: 'status', label: '启用', bool: true },
  ],
})

console.log('basic pages done')
