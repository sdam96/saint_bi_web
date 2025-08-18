<template>
  <div class="row justify-content-center mt-5">
    <div class="col-md-4">
      <div class="card">
        <div class="card-body">
          <h1 class="card-title text-center mb-4">Iniciar Sesión</h1>
          <form @submit.prevent="handleLogin">
            <div class="mb-3">
              <label for="username" class="form-label">Usuario:</label>
              <input type="text" class="form-control" id="username" v-model="username" placeholder="admin" required />
            </div>

            <div class="mb-3">
              <label for="password" class="form-label">Clave:</label>
              <input type="password" class="form-control" id="password" v-model="password" placeholder="*******" required />
            </div>

            <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

            <div class="d-grid">
              <button type="submit" class="btn btn-primary" :disabled="isLoading">
                {{ isLoading ? 'Entrando...' : 'Entrar' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../store/auth';

// 'ref' se usa para crear variables reactivas
const username = ref('');
const password = ref('');
const errorMessage = ref(null);
const isLoading = ref(false);

const router = useRouter();
const authStore = useAuthStore();

const handleLogin = async () => {
  isLoading.value = true;
  errorMessage.value = null;

  const success = await authStore.login(username.value, password.value);

  isLoading.value = false;

  if (success) {
    // Si el login es exitoso, el 'navigation guard' del router
    // nos redirigirá automáticamente al dashboard.
    router.push('/dashboard');
  } else {
    errorMessage.value = 'Credenciales inválidas. Por favor, intente de nuevo.';
  }
};
</script>