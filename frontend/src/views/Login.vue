<script setup lang="ts">import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
const router = useRouter();
const authStore = useAuthStore();
const email = ref('');
const password = ref('');
const error = ref('');
const loading = ref(false);
const handleLogin = async () => {
 if (!email.value || !password.value) {
 error.value = '请填写所有字段';
 return;
 }
 loading.value = true;
 try {
 await authStore.login({ email: email.value, password: password.value });
 await router.push('/');
 }
 catch (err) {
 error.value = '登录失败，请检查邮箱和密码';
 }
 finally {
 loading.value = false;
 }
};
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <div class="w-full max-w-md">
      <div class="bg-white rounded-xl shadow-lg p-8">
        <div class="text-center mb-8">
          <h1 class="text-2xl font-bold text-gray-800">Reqman AI</h1>
          <p class="text-gray-500 mt-2">登录您的账户</p>
        </div>
        
        <form @submit.prevent="handleLogin">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">邮箱</label>
            <input
              v-model="email"
              type="email"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
              placeholder="请输入邮箱"
            />
          </div>
          
          <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 mb-2">密码</label>
            <input
              v-model="password"
              type="password"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
              placeholder="请输入密码"
            />
          </div>
          
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition"
          >
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>
        
        <div v-if="error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {{ error }}
        </div>
        
        <p class="mt-6 text-center text-gray-500">
          还没有账户？
          <router-link to="/register" class="text-blue-600 hover:text-blue-700 font-medium">注册</router-link>
        </p>
      </div>
    </div>
  </div>
</template>