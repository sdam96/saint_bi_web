import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // Redirige cualquier solicitud que comience con /api
      // al servidor backend de Go que se ejecuta en el puerto 8080.
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true, // Necesario para los hosts virtuales
      }
    }
  }
})
