<template>
  <div>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索菜品" clearable style="width:200px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openEdit()">新增菜品</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="category_id" label="分类ID" width="90" />
      <el-table-column prop="station_id" label="档口ID" width="90" />
      <el-table-column prop="base_price" label="基础价" width="100" />
      <el-table-column label="上架" width="80">
        <template #default="{ row }">{{ row.is_on_sale ? '是' : '否' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="openMaterials(row)">配方</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />

    <el-dialog v-model="visible" :title="form.id ? '编辑菜品' : '新增菜品'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="分类">
          <el-select v-model="form.category_id" style="width:100%">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="档口">
          <el-select v-model="form.station_id" clearable style="width:100%">
            <el-option v-for="s in stations" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="基础价"><el-input v-model="form.base_price" placeholder="如 28.00" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort_no" :min="0" /></el-form-item>
        <el-form-item label="上架"><el-switch v-model="form.is_on_sale" /></el-form-item>
        <el-form-item label="图片URL"><el-input v-model="form.image_url" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="matVisible" title="菜品配方（物料）" width="640px">
      <div style="margin-bottom:8px">
        <el-button size="small" type="primary" @click="matRows.push({ material_id: undefined, qty: '1', remark: '' })">添加行</el-button>
      </div>
      <el-table :data="matRows" border size="small">
        <el-table-column label="物料">
          <template #default="{ row }">
            <el-select v-model="row.material_id" filterable style="width:100%">
              <el-option v-for="m in materials" :key="m.id" :label="`${m.name}(${m.code})`" :value="m.id" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="用量" width="120">
          <template #default="{ row }"><el-input v-model="row.qty" /></template>
        </el-table-column>
        <el-table-column label="备注" width="140">
          <template #default="{ row }"><el-input v-model="row.remark" /></template>
        </el-table-column>
        <el-table-column label="" width="70">
          <template #default="{ $index }">
            <el-button link type="danger" @click="matRows.splice($index, 1)">删</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="matVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveMaterials">保存配方</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiPage, apiPost, apiPut, apiDelete, apiGet } from '@openbows/shared'

const list = ref<any[]>([])
const categories = ref<any[]>([])
const stations = ref<any[]>([])
const materials = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const visible = ref(false)
const matVisible = ref(false)
const currentDishId = ref(0)
const matRows = ref<any[]>([])
const form = reactive<any>({
  id: 0, category_id: undefined, station_id: undefined, name: '', description: '',
  base_price: '0', is_on_sale: true, sort_no: 0, image_url: '',
})

async function loadRefs() {
  const [c, s, m] = await Promise.all([
    apiPage<any>('/menu-categories', { page: 1, page_size: 200 }),
    apiPage<any>('/stations', { page: 1, page_size: 200 }),
    apiPage<any>('/materials', { page: 1, page_size: 500 }),
  ])
  categories.value = c.list || []
  stations.value = s.list || []
  materials.value = m.list || []
}

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/dishes', { page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined })
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
      id: row.id, category_id: row.category_id, station_id: row.station_id || undefined,
      name: row.name, description: row.description || '', base_price: String(row.base_price ?? '0'),
      is_on_sale: !!row.is_on_sale, sort_no: row.sort_no || 0, image_url: row.image_url || '',
    })
  } else {
    Object.assign(form, {
      id: 0, category_id: categories.value[0]?.id, station_id: undefined, name: '', description: '',
      base_price: '0', is_on_sale: true, sort_no: 0, image_url: '',
    })
  }
  visible.value = true
}

async function save() {
  saving.value = true
  try {
    const body = {
      category_id: form.category_id,
      station_id: form.station_id || null,
      name: form.name,
      description: form.description,
      base_price: form.base_price,
      is_on_sale: form.is_on_sale,
      sort_no: form.sort_no,
      image_url: form.image_url,
    }
    if (form.id) await apiPut(`/dishes/${form.id}`, body)
    else await apiPost('/dishes', body)
    ElMessage.success('保存成功')
    visible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function openMaterials(row: any) {
  currentDishId.value = row.id
  try {
    const detail = await apiGet<any>(`/dishes/${row.id}`)
    matRows.value = (detail.materials || []).map((x: any) => ({
      material_id: x.material_id,
      qty: String(x.qty),
      remark: x.remark || '',
    }))
    if (!matRows.value.length) matRows.value = [{ material_id: undefined, qty: '1', remark: '' }]
    matVisible.value = true
  } catch (e: any) {
    ElMessage.error(e.message || '加载配方失败')
  }
}

async function saveMaterials() {
  saving.value = true
  try {
    const materialsPayload = matRows.value
      .filter((r) => r.material_id)
      .map((r) => ({ material_id: r.material_id, qty: String(r.qty), remark: r.remark || '' }))
    await apiPut(`/dishes/${currentDishId.value}/materials`, { materials: materialsPayload })
    ElMessage.success('配方已保存')
    matVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function onDelete(row: any) {
  await ElMessageBox.confirm(`确认删除菜品「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await apiDelete(`/dishes/${row.id}`)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(async () => {
  await loadRefs()
  await load()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
