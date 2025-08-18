<template>
  <div>
    <Navbar v-if="authStore.isAuthenticated" />
    <main class="container mt-4">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { useAuthStore } from './store/auth';
import Navbar from './components/Navbar.vue';

// Obtenemos una instancia de nuestro store de autenticación
const authStore = useAuthStore();
</script>

<style>
/* Puedes añadir estilos globales aquí si lo necesitas */
body {
  background-color: #f8f9fa; /* Un color de fondo suave */
}
</style>