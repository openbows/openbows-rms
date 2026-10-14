<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="brand" role="link" tabindex="0" title="返回首页" @click="goHome" @keydown.enter="goHome">
        Openbows RMS
      </div>
      <el-menu
        :default-active="route.path"
        :default-openeds="openedGroups"
        router
        background-color="#1f2d3d"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-sub-menu v-for="g in menus" :key="g.key" :index="g.key">
          <template #title>
            <span>{{ g.title }}</span>
          </template>
          <el-menu-item v-for="m in g.children" :key="m.path" :index="m.path">
            <span>{{ m.title }}</span>
          </el-menu-item>
        </el-sub-menu>
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
  {
    key: 'system',
    title: '系统管理',
    children: [
      { path: '/stores', title: '门店' },
      { path: '/users', title: '用户' },
      { path: '/roles', title: '角色' },
    ],
  },
  {
    key: 'menu',
    title: '菜单管理',
    children: [
      { path: '/stations', title: '档口' },
      { path: '/categories', title: '分类' },
      { path: '/dishes', title: '菜品' },
    ],
  },
  {
    key: 'table',
    title: '桌台服务',
    children: [
      { path: '/areas', title: '区域' },
      { path: '/tables', title: '桌台' },
      { path: '/reservations', title: '预订' },
      { path: '/queues', title: '排队' },
    ],
  },
  {
    key: 'member',
    title: '会员管理',
    children: [
      { path: '/members', title: '会员' },
    ],
  },
  {
    key: 'stock',
    title: '库存管理',
    children: [
      { path: '/materials', title: '物料' },
      { path: '/stock-docs', title: '库存单据' },
      { path: '/stock-balances', title: '库存余额' },
    ],
  },
  {
    key: 'purchase',
    title: '采购管理',
    children: [
      { path: '/suppliers', title: '供应商' },
      { path: '/purchase-orders', title: '采购单' },
      { path: '/purchase-receipts', title: '采购入库' },
    ],
  },
  {
    key: 'finance',
    title: '财务管理',
    children: [
      { path: '/settlements', title: '日结' },
      { path: '/reconciliations', title: '对账' },
    ],
  },
  {
    key: 'hr',
    title: '人力资源',
    children: [
      { path: '/employees', title: '员工' },
      { path: '/shifts', title: '排班' },
      { path: '/attendances', title: '考勤' },
    ],
  },
]

const openedGroups: string[] = []

onMounted(() => {
  auth.loadMe().catch(() => {
    auth.logout()
    router.push('/login')
  })
})

function goHome() {
  router.push('/home')
}

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
  cursor: pointer;
  user-select: none;
  transition: background 0.2s ease, color 0.2s ease;
}
.brand:hover {
  background: rgba(64, 158, 255, 0.12);
  color: #9ecbff;
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
