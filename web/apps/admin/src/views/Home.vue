<template>
  <div class="home">
    <section class="hero">
      <div class="hero-glow" aria-hidden="true" />
      <p class="eyebrow">Openbows RMS</p>
      <h1>餐饮管理系统</h1>
      <p class="lead">
        欢迎回来{{ auth.displayName ? `，${auth.displayName}` : '' }}。从这里快速进入门店经营、菜单、桌台、库存与财务等模块。
      </p>
      <div class="hero-meta">
        <span>{{ today }}</span>
        <a href="http://www.openbows.com" target="_blank" rel="noopener noreferrer">www.openbows.com</a>
      </div>
    </section>

    <section class="section">
      <h2>快捷入口</h2>
      <div class="grid">
        <button
          v-for="item in shortcuts"
          :key="item.path"
          type="button"
          class="card"
          @click="go(item.path)"
        >
          <span class="card-icon" :style="{ background: item.tone }">{{ item.mark }}</span>
          <span class="card-body">
            <span class="card-title">{{ item.title }}</span>
            <span class="card-desc">{{ item.desc }}</span>
          </span>
          <span class="card-arrow">→</span>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const today = computed(() =>
  new Date().toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }),
)

const shortcuts = [
  { path: '/stores', title: '门店', desc: '组织与门店基础信息', mark: '店', tone: 'linear-gradient(135deg,#2f6fed,#5b9dff)' },
  { path: '/dishes', title: '菜品', desc: '菜单、规格与档口出品', mark: '菜', tone: 'linear-gradient(135deg,#0f8a6b,#3ecf9a)' },
  { path: '/tables', title: '桌台', desc: '区域桌台与营业状态', mark: '台', tone: 'linear-gradient(135deg,#c27803,#f0b429)' },
  { path: '/queues', title: '排队', desc: '叫号与入座流转', mark: '排', tone: 'linear-gradient(135deg,#c2410c,#fb923c)' },
  { path: '/members', title: '会员', desc: '会员档案与消费关联', mark: '会', tone: 'linear-gradient(135deg,#0369a1,#38bdf8)' },
  { path: '/stock-balances', title: '库存', desc: '结存与出入库单据', mark: '库', tone: 'linear-gradient(135deg,#0e7490,#22d3ee)' },
  { path: '/purchase-orders', title: '采购', desc: '采购单与入库协同', mark: '采', tone: 'linear-gradient(135deg,#334155,#64748b)' },
  { path: '/settlements', title: '日结', desc: '营业日结与渠道对账', mark: '财', tone: 'linear-gradient(135deg,#be123c,#fb7185)' },
]

function go(path: string) {
  router.push(path)
}
</script>

<style scoped>
.home {
  min-height: calc(100vh - 88px);
  margin: -16px;
  padding: 28px 28px 40px;
  background:
    radial-gradient(1200px 420px at 12% -10%, rgba(64, 158, 255, 0.18), transparent 55%),
    radial-gradient(900px 380px at 90% 0%, rgba(94, 114, 140, 0.14), transparent 50%),
    linear-gradient(180deg, #f3f6fa 0%, #eef2f7 45%, #f7f9fc 100%);
}

.hero {
  position: relative;
  overflow: hidden;
  border-radius: 20px;
  padding: 36px 40px;
  color: #fff;
  background: linear-gradient(125deg, #1f2d3d 0%, #2d4158 48%, #3d5a78 100%);
  box-shadow: 0 18px 40px rgba(31, 45, 61, 0.22);
  animation: rise 0.55s ease-out both;
}

.hero-glow {
  position: absolute;
  right: -80px;
  top: -90px;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(64, 158, 255, 0.35), transparent 68%);
  pointer-events: none;
  animation: pulse 4.5s ease-in-out infinite;
}

.eyebrow {
  margin: 0 0 10px;
  font-size: 13px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.72);
}

.hero h1 {
  margin: 0;
  font-size: 34px;
  font-weight: 720;
  letter-spacing: 0.02em;
}

.lead {
  max-width: 640px;
  margin: 14px 0 0;
  font-size: 15px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.86);
}

.hero-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 28px;
  margin-top: 22px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
}

.hero-meta a {
  color: #9ecbff;
  text-decoration: none;
}

.hero-meta a:hover {
  color: #fff;
  text-decoration: underline;
}

.section {
  margin-top: 28px;
  animation: rise 0.65s ease-out 0.08s both;
}

.section h2 {
  margin: 0 0 14px;
  font-size: 18px;
  font-weight: 650;
  color: #1f2d3d;
}

.grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.card {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 16px 14px;
  border: 1px solid rgba(31, 45, 61, 0.08);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 8px 20px rgba(31, 45, 61, 0.06);
  text-align: left;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.card:hover {
  transform: translateY(-3px);
  border-color: rgba(64, 158, 255, 0.35);
  box-shadow: 0 14px 28px rgba(31, 45, 61, 0.12);
}

.card-icon {
  flex: 0 0 42px;
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 15px;
  font-weight: 700;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.card-title {
  font-size: 15px;
  font-weight: 650;
  color: #1f2d3d;
}

.card-desc {
  font-size: 12px;
  color: #748194;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-arrow {
  margin-left: auto;
  color: #94a3b8;
  font-size: 16px;
  transition: transform 0.2s ease, color 0.2s ease;
}

.card:hover .card-arrow {
  color: #409eff;
  transform: translateX(3px);
}

@keyframes rise {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% { opacity: 0.7; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.06); }
}

@media (max-width: 1200px) {
  .grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 720px) {
  .home { padding: 16px; }
  .hero { padding: 28px 22px; }
  .hero h1 { font-size: 26px; }
  .grid { grid-template-columns: 1fr; }
}
</style>
