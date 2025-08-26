<template>
  <div>
    <Navbar v-if="authStore.isAuthenticated" />
    <main class="container mt-4">
      <router-view />
    </main>

    <SessionTimeoutModal
      v-if="authStore.isAuthenticated"
      :countdown="countdown"
      @extend="handleExtend"
      @logout="logout"
    />
  </div>
</template>

<script setup>
import { useAuthStore } from './store/auth';
import Navbar from './components/Navbar.vue';
// --- AÑADIR ESTAS IMPORTACIONES ---
import SessionTimeoutModal from './components/SessionTimeoutModal.vue';
import { useSessionManager } from './composables/useSessionManager';
// --- FIN DE IMPORTACIONES ---

const authStore = useAuthStore();

// Inicializamos nuestro gestor de sesión.
// Esto nos da acceso a las variables y funciones que expusimos (countdown, handleExtend, etc.).
const { countdown, handleExtend, logout } = useSessionManager();
// --- FIN ---
</script>

<style>
body {
  background-color: #f8f9fa;
}
</style>