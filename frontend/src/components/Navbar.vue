<template>
  <nav class="navbar navbar-expand-lg navbar-dark sticky-top">
    <div class="container">
      <router-link to="/dashboard" class="navbar-brand fw-bold d-flex align-items-center">
        <img src="/src/assets/vue.svg" alt="Logo" width="30" height="24" class="me-2">
        <span>SAINT BI</span>
      </router-link>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav me-auto mb-2 mb-lg-0">
          <li class="nav-item">
            <router-link to="/dashboard" class="nav-link">Inicio</router-link>
          </li>
          <li class="nav-item dropdown">
            <a class="nav-link dropdown-toggle" href="#" id="navbarDropdownAnalytics" role="button" data-bs-toggle="dropdown" aria-expanded="false">
              Análisis
            </a>
            <ul class="dropdown-menu" aria-labelledby="navbarDropdownAnalytics">
              <li><router-link to="/analytics/sales-forecast" class="dropdown-item">Proyección de Ventas</router-link></li>
              <li><router-link to="/analytics/market-basket" class="dropdown-item">Análisis de Canasta</router-link></li>
            </ul>
          </li>
          <li class="nav-item">
            <router-link to="/connections" class="nav-link">Conexiones</router-link>
          </li>
          <li class="nav-item">
            <router-link to="/users" class="nav-link">Usuarios</router-link>
          </li>
          <li class="nav-item">
            <router-link to="/settings" class="nav-link">Configuración</router-link>
          </li>
        </ul>
        <div class="d-flex">
          <button @click="handleLogout" class="btn btn-outline-light">Salir</button>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../store/auth';
const authStore = useAuthStore();
const handleLogout = () => {
  authStore.logout();
};
</script>

<style scoped>
.navbar {
  background-color: rgba(30, 30, 36, 0.85);
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
.nav-link {
  transition: color 0.2s ease-in-out;
}
.nav-link:hover, .router-link-active {
  color: #fff !important;
  font-weight: 500;
}
.dropdown-menu {
  background-color: rgba(44, 44, 62, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
}
.dropdown-item {
  color: #dee2e6;
}
.dropdown-item:hover {
  background-color: var(--saint-blue);
  color: #fff;
}
</style>