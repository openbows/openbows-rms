<template>
  <div class="login-wrap">
    <el-card class="login-card" shadow="hover">
      <h2>餐饮管理系统</h2>
      <el-form :model="form" @submit.prevent="onSubmit">
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="admin" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="Admin@123" autocomplete="current-password" />
        </el-form-item>
        <el-button type="primary" style="width:100%" :loading="auth.loading" native-type="submit">登录</el-button>
      </el-form>
    </el-card>
    <footer class="login-footer">
      <a href="http://www.openbows.com" target="_blank" rel="noopener noreferrer">http://www.openbows.com</a>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const form = reactive({
  username: 'admin',
  password: 'Admin@123',
})

async function onSubmit() {
  try {
    await auth.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  }
}
</script>

<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f2d3d 0%, #3a4a5c 50%, #5b7a99 100%);
  position: relative;
  padding-bottom: 48px;
  box-sizing: border-box;
}
.login-card { width: 320px; }
h2 { margin: 0 0 20px; text-align: center; }
.login-footer {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  text-align: center;
}
.login-footer a {
  color: rgba(255, 255, 255, 0.75);
  text-decoration: none;
  font-size: 13px;
}
.login-footer a:hover {
  color: #fff;
  text-decoration: underline;
}
</style>
