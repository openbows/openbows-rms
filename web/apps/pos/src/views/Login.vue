<template>
  <div class="login-wrap">
    <el-card class="login-card" shadow="hover">
      <h2>Openbows 收银 POS</h2>
      <p class="hint">登录后进入收银台</p>
      <el-form :model="form" @submit.prevent="onSubmit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password autocomplete="current-password" />
        </el-form-item>
        <el-button type="primary" style="width:100%" :loading="auth.loading" native-type="submit">登录</el-button>
      </el-form>
    </el-card>
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
const form = reactive({ username: 'admin', password: 'Admin@123' })

async function onSubmit() {
  try {
    await auth.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.replace((route.query.redirect as string) || '/')
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  }
}
</script>

<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(120deg, #0f766e 0%, #134e4a 50%, #1e3a5f 100%);
}
.login-card { width: 400px; }
h2 { margin: 0 0 8px; text-align: center; }
.hint { text-align: center; color: #909399; margin: 0 0 20px; }
</style>
