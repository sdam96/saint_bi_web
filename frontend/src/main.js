// frontend/src/main.js
import { createApp } from 'vue';
import { createPinia } from 'pinia'; // Importar Pinia

// Importar Bootstrap CSS para el estilo
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap'; // Importar el JS de Bootstrap

import axios from 'axios';
import { useAuthStore } from './store/auth';

import App from './App.vue';
import router from './router'; 

// Configuración del interceptor de respuestas de Axios
axios.interceptors.response.use(
  // Si la respuesta es exitosa (código 2xx), simplemente la devolvemos.
  response => response,
  // Si hay un error...
  async (error) => {
    const authStore = useAuthStore();
    // Verificamos si el error es una respuesta 401 Unauthorized.
    if (error.response && error.response.status === 401) {
      // Si el usuario estaba autenticado en el frontend, significa que la sesión
      // del backend expiró. Forzamos el cierre de sesión.
      if (authStore.isAuthenticated) {
        console.warn('Sesión expirada detectada por el interceptor. Cerrando sesión.');
        await authStore.logout();
      }
    }
    // Devolvemos la promesa rechazada para que el código que originó la llamada pueda manejar el error.
    return Promise.reject(error);
  }
);

const app = createApp(App);

app.use(createPinia()); // Usar Pinia
app.use(router); // Usar el enrutador

app.mount('#app');