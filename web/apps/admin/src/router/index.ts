import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@openbows/shared'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/stores',
      children: [
        { path: 'stores', name: 'stores', component: () => import('@/views/modules/Stores.vue'), meta: { title: '门店' } },
        { path: 'users', name: 'users', component: () => import('@/views/modules/Users.vue'), meta: { title: '用户' } },
        { path: 'roles', name: 'roles', component: () => import('@/views/modules/Roles.vue'), meta: { title: '角色' } },
        { path: 'stations', name: 'stations', component: () => import('@/views/modules/Stations.vue'), meta: { title: '档口' } },
        { path: 'categories', name: 'categories', component: () => import('@/views/modules/Categories.vue'), meta: { title: '分类' } },
        { path: 'dishes', name: 'dishes', component: () => import('@/views/modules/Dishes.vue'), meta: { title: '菜品' } },
        { path: 'areas', name: 'areas', component: () => import('@/views/modules/Areas.vue'), meta: { title: '区域' } },
        { path: 'tables', name: 'tables', component: () => import('@/views/modules/Tables.vue'), meta: { title: '桌台' } },
        { path: 'members', name: 'members', component: () => import('@/views/modules/Members.vue'), meta: { title: '会员' } },
        { path: 'materials', name: 'materials', component: () => import('@/views/modules/Materials.vue'), meta: { title: '物料' } },
        { path: 'stock-docs', name: 'stock-docs', component: () => import('@/views/modules/StockDocs.vue'), meta: { title: '库存单据' } },
        { path: 'stock-balances', name: 'stock-balances', component: () => import('@/views/modules/StockBalances.vue'), meta: { title: '库存余额' } },
        { path: 'suppliers', name: 'suppliers', component: () => import('@/views/modules/Suppliers.vue'), meta: { title: '供应商' } },
        { path: 'purchase-orders', name: 'purchase-orders', component: () => import('@/views/modules/PurchaseOrders.vue'), meta: { title: '采购单' } },
        { path: 'purchase-receipts', name: 'purchase-receipts', component: () => import('@/views/modules/PurchaseReceipts.vue'), meta: { title: '采购入库' } },
        { path: 'settlements', name: 'settlements', component: () => import('@/views/modules/Settlements.vue'), meta: { title: '日结' } },
        { path: 'reconciliations', name: 'reconciliations', component: () => import('@/views/modules/Reconciliations.vue'), meta: { title: '对账' } },
        { path: 'employees', name: 'employees', component: () => import('@/views/modules/Employees.vue'), meta: { title: '员工' } },
        { path: 'shifts', name: 'shifts', component: () => import('@/views/modules/Shifts.vue'), meta: { title: '排班' } },
        { path: 'attendances', name: 'attendances', component: () => import('@/views/modules/Attendances.vue'), meta: { title: '考勤' } },
        { path: 'reservations', name: 'reservations', component: () => import('@/views/modules/Reservations.vue'), meta: { title: '预订' } },
        { path: 'queues', name: 'queues', component: () => import('@/views/modules/Queues.vue'), meta: { title: '排队' } },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) return true
  if (!getToken()) return { path: '/login', query: { redirect: to.fullPath } }
  return true
})

export default router
