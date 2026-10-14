<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="material_id" label="物料ID" width="100" />
      <el-table-column prop="qty" label="结存数量" />
      <el-table-column prop="avg_cost" label="平均成本" />
      <el-table-column prop="updated_at" label="更新时间" min-width="160" />
    </el-table>
    <el-pagination class="pager" v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="load" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { apiPage } from '@openbows/shared'

const list = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function load() {
  loading.value = true
  try {
    const data = await apiPage<any>('/stock-balances', { page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.pager { margin-top: 12px; justify-content: flex-end; }
</style>
