// frontend/src/main.js
import { createApp } from 'vue';
import { createPinia } from 'pinia'; // Importar Pinia

// Importar Bootstrap CSS para el estilo
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap'; // Importar el JS de Bootstrap

import App from './App.vue';
import router from './router'; // Importar nuestro enrutador

const app = createApp(App);

app.use(createPinia()); // Usar Pinia
app.use(router); // Usar el enrutador

app.mount('#app');