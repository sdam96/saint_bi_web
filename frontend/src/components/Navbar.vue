<template>
  <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
    <div class="container">
      <router-link to="/dashboard" class="navbar-brand">Saint BI</router-link>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
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
.router-link-active {
  font-weight: bold;
}
</style>