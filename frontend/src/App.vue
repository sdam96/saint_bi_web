<template>
  <div>
    <Navbar v-if="authStore.isAuthenticated" />
    <main class="container mt-4 mb-5">
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
import SessionTimeoutModal from './components/SessionTimeoutModal.vue';
import { useSessionManager } from './composables/useSessionManager';

const authStore = useAuthStore();
const { countdown, handleExtend, logout } = useSessionManager();
</script>