// frontend/src/store/dashboard.js
import { defineStore } from 'pinia';
import axios from 'axios';

export const useDashboardStore = defineStore('dashboard', {
    state: () => ({
        connections: [],
        selectedConnectionId: null,
        summaryData: null,
        isLoading: false,
        error: null,
    }),

    actions: {
        // Acción para obtener la lista de conexiones disponibles
        async fetchConnections() {
            try {
                const response = await axios.get('/api/connections');
                this.connections = response.data;
            } catch (err) {
                this.error = 'Error al cargar las conexiones.';
                console.error(err);
            }
        },

        // Acción que se ejecuta cuando el usuario selecciona una conexión
        async selectConnection(connectionId) {
            this.selectedConnectionId = connectionId;
            this.summaryData = null; // Limpiar datos anteriores
            this.error = null;
            
            if (!connectionId) {
                return; // Si el usuario deselecciona, no hacemos nada más
            }

            try {
                // 1. Le decimos al backend qué conexión usar
                await axios.post('/api/dashboard/select-connection', {
                    connection_id: connectionId,
                });

                // 2. Pedimos los datos del dashboard para esa conexión
                await this.fetchDashboardData();

            } catch (err) {
                this.error = 'No se pudo seleccionar la conexión o cargar los datos.';
                console.error(err);
            }
        },

        // Acción para obtener los datos del resumen gerencial
        async fetchDashboardData() {
            if (!this.selectedConnectionId) return;

            this.isLoading = true;
            this.error = null;
            try {
                const response = await axios.get('/api/dashboard/data');
                this.summaryData = response.data;
            } catch (err) {
                this.error = 'Error al cargar los datos del dashboard.';
                this.summaryData = null;
                console.error(err);
            } finally {
                this.isLoading = false;
            }
        },
    },
});