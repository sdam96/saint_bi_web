import { createApp } from 'vue';
import { createPinia } from 'pinia';
import Particles from "@tsparticles/vue3";
import { loadSlim } from "tsparticles-slim";

import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap';

import axios from 'axios';
import { useAuthStore } from './store/auth';

import App from './App.vue';
import router from './router';

axios.interceptors.response.use(
  response => response,
  async (error) => {
    const authStore = useAuthStore();
    if (error.response && error.response.status === 401) {
      if (authStore.isAuthenticated) {
        console.warn('Sesión expirada detectada por el interceptor. Cerrando sesión.');
        await authStore.logout();
      }
    }
    return Promise.reject(error);
  }
);

const app = createApp(App);

app.use(createPinia());
app.use(router);

app.use(Particles, {
  init: async engine => {
    await loadSlim(engine);
  },
});

app.mount('#app');