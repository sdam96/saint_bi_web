// src/store/auth.js
import { defineStore } from 'pinia';
import axios from 'axios';
import router from '../router'; // <-- Importar el enrutador

export const useAuthStore = defineStore('auth', {
    state: () => ({
        isAuthenticated: localStorage.getItem('isAuthenticated') === 'true',
        error: null,
    }),
    getters: {
        isUserAuthenticated: (state) => state.isAuthenticated,
    },
    actions: {
        async login(username, password) {
            try {
                const response = await axios.post('/api/login', { username, password });
                if (response.status === 200) {
                    // Se actualiza el estado correctamente.
                    this.isAuthenticated = true;
                    localStorage.setItem('isAuthenticated', 'true');
                    this.error = null;
                    
                    // Se fuerza la redirección al dashboard.
                    await router.push('/dashboard');
                    return true;
                }
            } catch (err) {
                this.error = 'Credenciales inválidas o error del servidor.';
                this.isAuthenticated = false;
                localStorage.removeItem('isAuthenticated');
                return false;
            }
        },
        async logout() {
            try {
                await axios.post('/api/logout');
            } catch (error) {
                console.error("Error durante el logout en el servidor:", error);
            } finally {
                this.isAuthenticated = false;
                localStorage.removeItem('isAuthenticated');
                await router.push('/login');
            }
        },
        async extendSession() {
            try {
                await axios.post('/api/session/extend');
                console.log('Sesión extendida exitosamente.');
                return true;
            } catch (error) {
                console.error('Fallo al extender la sesión:', error);
                // Si falla la extensión, cerramos la sesión para evitar inconsistencias.
                this.logout();
                return false;
            }
        },
    },
});