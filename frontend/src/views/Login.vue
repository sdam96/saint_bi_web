<template>
  <div class="login-container">
    <div class="card login-card animated-fade-in">
      <div class="card-body p-4 p-md-5">
        <div class="text-center mb-4">
          <img src="/src/assets/vue.svg" alt="Logo" width="72" height="57" class="mb-3">
          <h1 class="h3 fw-normal">Bienvenido a SAINT BI</h1>
          <p class="text-muted">Inicia sesión para continuar</p>
        </div>
        <form @submit.prevent="handleLogin">
          <div v-if="error" class="alert alert-danger">{{ error }}</div>
          <div class="form-floating mb-3">
            <input type="text" class="form-control" id="username" v-model="username" placeholder="Usuario" required>
            <label for="username">Usuario</label>
          </div>
          <div class="form-floating mb-3">
            <input type="password" class="form-control" id="password" v-model="password" placeholder="Clave" required>
            <label for="password">Clave</label>
          </div>
          <button class="w-100 btn btn-lg btn-primary" type="submit" :disabled="isLoading">
            {{ isLoading ? 'Ingresando...' : 'Ingresar' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useAuthStore } from '../store/auth';

const username = ref('');
const password = ref('');
const error = ref(null);
const isLoading = ref(false);
const authStore = useAuthStore();

const handleLogin = async () => {
  error.value = null;
  isLoading.value = true;
  const success = await authStore.login(username.value, password.value);
  if (!success) {
    error.value = authStore.error || 'Credenciales inválidas.';
  }
  isLoading.value = false;
};
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
}
.login-card {
  width: 100%;
  max-width: 420px;
}
</style>