// frontend/src/store/admin.js
import { defineStore } from 'pinia';
import axios from 'axios';

export const useAdminStore = defineStore('admin', {
    state: () => ({
        connections: [],
        users: [],
        isLoading: false,
        error: null,
    }),

    actions: {
        // --- Acciones para Conexiones ---
        async fetchConnections() {
            this.isLoading = true;
            this.error = null;
            try {
                const response = await axios.get('/api/connections');
                this.connections = response.data || [];
            } catch (err) {
                this.error = 'Error al cargar las conexiones.';
                console.error(err);
            } finally {
                this.isLoading = false;
            }
        },

        async addConnection(connectionData) {
            this.error = null;
            try {
                const response = await axios.post('/api/connections', connectionData);
                this.connections = response.data; // La API devuelve la lista actualizada
            } catch (err) {
                this.error = 'Error al agregar la conexión.';
                console.error(err);
            }
        },

        async deleteConnection(id) {
            this.error = null;
            try {
                await axios.delete(`/api/connections/${id}`);
                // Eliminamos la conexión de la lista local para una actualización instantánea
                this.connections = this.connections.filter(c => c.ID !== id);
            } catch (err) {
                this.error = 'Error al eliminar la conexión.';
                console.error(err);
            }
        },

        // --- Acciones para Usuarios ---
        async fetchUsers() {
            this.isLoading = true;
            this.error = null;
            try {
                const response = await axios.get('/api/users');
                this.users = response.data || [];
            } catch (err) {
                this.error = 'Error al cargar los usuarios.';
                console.error(err);
            } finally {
                this.isLoading = false;
            }
        },

        async addUser(userData) {
            this.error = null;
            try {
                const response = await axios.post('/api/users', userData);
                this.users = response.data; // La API devuelve la lista actualizada
            } catch (err) {
                this.error = 'Error al agregar el usuario.';
                console.error(err);
            }
        },
    },
});