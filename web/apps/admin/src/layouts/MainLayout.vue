<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="brand">Openbows RMS</div>
      <el-menu :default-active="route.path" router background-color="#1f2d3d" text-color="#bfcbd9" active-text-color="#409eff">
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <span>{{ m.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span class="title">{{ (route.meta.title as string) || '管理后台' }}</span>
        <div class="right">
          <span class="user">{{ auth.displayName }}</span>
          <el-button type="danger" link @click="onLogout">退出</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const menus = [
  { path: '/stores', title: '门店' },
  { path: '/users', title: '用户' },
  { path: '/roles', title: '角色' },
  { path: '/stations', title: '档口' },
  { path: '/categories', title: '分类' },
  { path: '/dishes', title: '菜品' },
  { path: '/areas', title: '区域' },
  { path: '/tables', title: '桌台' },
  { path: '/members', title: '会员' },
  { path: '/materials', title: '物料' },
  { path: '/stock-docs', title: '库存单据' },
  { path: '/stock-balances', title: '库存余额' },
  { path: '/suppliers', title: '供应商' },
  { path: '/purchase-orders', title: '采购单' },
  { path: '/purchase-receipts', title: '采购入库' },
  { path: '/settlements', title: '日结' },
  { path: '/reconciliations', title: '对账' },
  { path: '/employees', title: '员工' },
  { path: '/shifts', title: '排班' },
  { path: '/attendances', title: '考勤' },
  { path: '/reservations', title: '预订' },
  { path: '/queues', title: '排队' },
]

onMounted(() => {
  auth.loadMe().catch(() => {
    auth.logout()
    router.push('/login')
  })
})

function onLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout { height: 100vh; }
.aside { background: #1f2d3d; overflow-y: auto; }
.brand {
  height: 56px; line-height: 56px; text-align: center;
  color: #fff; font-weight: 700; font-size: 16px;
  border-bottom: 1px solid #2d3a4b;
}
.header {
  display: flex; align-items: center; justify-content: space-between;
  background: #fff; border-bottom: 1px solid #ebeef5;
}
.title { font-size: 16px; font-weight: 600; }
.right { display: flex; align-items: center; gap: 12px; }
.user { color: #606266; }
.main { padding: 16px; }
.el-menu { border-right: none; }
</style>
