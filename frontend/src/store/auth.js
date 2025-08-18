// frontend/src/store/auth.js
import { defineStore } from 'pinia';
import axios from 'axios'; // Usaremos axios para las llamadas a la API

export const useAuthStore = defineStore('auth', {
    state: () => ({
        // Inicializamos el estado desde localStorage para mantener la sesión si el usuario recarga la página.
        user: JSON.parse(localStorage.getItem('user')),
    }),

    getters: {
        // Un getter para saber fácilmente si el usuario está autenticado.
        isAuthenticated: (state) => !!state.user,
        // Getter para obtener los datos del usuario.
        currentUser: (state) => state.user,
    },

    actions: {
        // Acción para manejar el inicio de sesión.
        async login(username, password) {
            try {
                // Hacemos la llamada a nuestro endpoint de la API de Go.
                const response = await axios.post('/api/login', {
                    username,
                    password,
                });

                // Si la llamada es exitosa, guardamos los datos del usuario.
                this.user = response.data;
                // Guardamos en localStorage para persistir la sesión.
                localStorage.setItem('user', JSON.stringify(response.data));

                return true; // Indicamos que el login fue exitoso
            } catch (error) {
                console.error('Error en el inicio de sesión:', error);
                // Limpiamos cualquier dato de usuario que pudiera haber.
                this.user = null;
                localStorage.removeItem('user');
                return false; // Indicamos que el login falló
            }
        },

        // Acción para manejar el cierre de sesión.
        async logout() {
            try {
                // Llama al endpoint de logout de la API.
                await axios.post('/api/logout');
            } catch (error) {
                console.error('Error durante el logout en el servidor, cerrando sesión localmente de todas formas.', error);
            } finally {
                // Se limpia el estado local sin importar el resultado de la API.
                this.user = null;
                localStorage.removeItem('user');
                // Redirigimos al usuario a la página de login.
                // Importamos el router aquí para evitar dependencias circulares.
                const router = (await import('../router')).default;
                router.push('/login');
            }
        },
    },
});