// src/store/dashboard.js
import { defineStore } from 'pinia';
import axios from 'axios';

export const useDashboardStore = defineStore('dashboard', {
    state: () => ({
        connections: [],
        selectedConnectionId: null,
        summaryData: null,
        connectionSettings: null,
        isLoading: false,
        error: null,
    }),

    actions: {
        async fetchConnections() {
            try {
                const response = await axios.get('/api/connections');
                this.connections = response.data;
            } catch (err) {
                this.error = 'Error al cargar las conexiones.';
                console.error(err);
            }
        },

        // La acción ahora acepta un objeto de fechas.
        async selectConnection(connectionId, dates) {
            this.selectedConnectionId = connectionId;
            this.summaryData = null;
            this.connectionSettings = null;
            this.error = null;
            
            if (!connectionId) {
                return;
            }

            try {
                await axios.post('/api/dashboard/select-connection', {
                    connection_id: connectionId,
                });

                await this.fetchConnectionSettings()

                // Pasamos las fechas al solicitar los datos por primera vez.
                await this.fetchDashboardData(dates);
            } catch (err) {
                this.error = 'No se pudo seleccionar la conexión o cargar los datos.';
                console.error(err);
            }
        },

        async fetchConnectionSettings(){
            this.isLoading = true;
            try {
                const response = await axios.get('/api/settings');
                this.connectionSettings = response.data;
            } catch (error) {
                this.error = 'Error al cargar la configuracion';
                console.log(error);
            } finally {
                this.isLoading = false;
            }
        },

        async saveConnectionSetting(settings) {
            try {
                await axios.post('/api/settings', {
                    CurrencyISO: settings.CurrencyISO,
                    Locale: settings.LocaleFormat,
                });

                this.connectionSettings.CurrencyISO = settings.CurrencyISO;
                this.connectionSettings.LocaleFormat = settings.LocaleFormat;
            } catch (error) {
                this.error = "Error al guardar la configuracion";
                console.error(error);
                
            }
        },

        // La acción de buscar datos ahora requiere las fechas.
        async fetchDashboardData(dates) {
            if (!this.selectedConnectionId || !dates.startDate || !dates.endDate) return;

            this.isLoading = true;
            this.error = null;
            try {
                // Se añaden las fechas como 'params' a la solicitud GET.
                // Axios los convertirá en /api/dashboard/data?startDate=...&endDate=...
                const response = await axios.get('/api/dashboard/data', {
                    params: dates 
                });
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