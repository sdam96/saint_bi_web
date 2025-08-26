// src/store/dashboard.js
import { defineStore } from 'pinia';
import axios from 'axios';

// Clave para guardar la configuración consolidada en el localStorage.
const CONSOLIDATED_SETTINGS_KEY = 'consolidated-settings';

export const useDashboardStore = defineStore('dashboard', {
    state: () => ({
        connections: [],
        selectedConnectionId: 0, 
        summaryData: null,
        connectionSettings: null, // Este campo ahora contendrá la configuración activa, ya sea de una conexión real o la consolidada.
        isLoading: false,
        error: null,
    }),

    actions: {
        async fetchConnections() {
            try {
                const response = await axios.get('/api/connections');
                this.connections = response.data;
                if (this.connections.length > 0) {
                    // Por defecto, selecciona "Consolidado" (ID 0) y carga su configuración.
                    const today = new Date();
                    const yesterday = new Date(today);
                    yesterday.setDate(yesterday.getDate() - 30);
                    
                    const formatDate = (date) => date.toISOString().split('T')[0];

                    this.selectConnection(0, { startDate: formatDate(yesterday), endDate: formatDate(today) });
                }
            } catch (err) {
                this.error = 'Error al cargar las conexiones.';
                console.error(err);
            }
        },

        // **ACCIÓN MODIFICADA**
        async selectConnection(connectionId, dates) {
            this.selectedConnectionId = connectionId;
            this.summaryData = null;
            this.connectionSettings = null; // Limpiamos la configuración anterior
            this.error = null;
            
            if (connectionId === null || connectionId === undefined) {
                return;
            }

            try {
                // Notificamos al backend la selección para que el middleware se prepare
                await axios.post('/api/dashboard/select-connection', {
                    connection_id: connectionId,
                });

                // --- LÓGICA CLAVE ---
                // Si es la consolidada, cargamos desde localStorage.
                // Si es una conexión real, la pedimos a la API.
                if (connectionId === 0) {
                    this.loadConsolidatedSettings();
                } else {
                    await this.fetchConnectionSettingsAPI();
                }
                // --- FIN DE LÓGICA CLAVE ---

                await this.fetchDashboardData(dates);
            } catch (err) {
                this.error = 'No se pudo seleccionar la conexión o cargar los datos.';
                console.error(err);
            }
        },
        
        // **NUEVA ACCIÓN INTERNA**
        // Carga la configuración consolidada desde localStorage.
        loadConsolidatedSettings() {
            try {
                const storedSettings = localStorage.getItem(CONSOLIDATED_SETTINGS_KEY);
                if (storedSettings) {
                    this.connectionSettings = JSON.parse(storedSettings);
                } else {
                    // Si no hay nada guardado, usamos valores por defecto.
                    this.connectionSettings = {
                        LocaleFormat: 'es-VE', // Un default más localizado
                        CurrencyISO: 'USD',
                    };
                }
            } catch (e) {
                console.error("Error al leer la configuración consolidada:", e);
                // En caso de error, usamos valores por defecto.
                 this.connectionSettings = { LocaleFormat: 'es-VE', CurrencyISO: 'USD' };
            }
        },

        // **ACCIÓN RENOMBRADA** (antes fetchConnectionSettings)
        // Obtiene la configuración de una conexión específica desde la API.
        async fetchConnectionSettingsAPI() {
            this.isLoading = true;
            try {
                const response = await axios.get('/api/settings');
                this.connectionSettings = response.data;
            } catch (error) {
                this.error = 'Error al cargar la configuración de la conexión.';
                console.log(error);
            } finally {
                this.isLoading = false;
            }
        },

        // **ACCIÓN MODIFICADA**
        // Guarda la configuración.
        async saveConnectionSettings(settings) {
            try {
                // Si es la conexión consolidada, guardamos en localStorage.
                if (this.selectedConnectionId === 0) {
                    localStorage.setItem(CONSOLIDATED_SETTINGS_KEY, JSON.stringify(settings));
                    this.connectionSettings = settings; // Actualizamos el estado
                } else {
                // Si es una conexión normal, llamamos a la API.
                    await axios.post('/api/settings', {
                        CurrencyISO: settings.CurrencyISO,
                        LocaleFormat: settings.LocaleFormat,
                    });
                    this.connectionSettings = settings; // Actualizamos el estado
                }
            } catch (error) {
                this.error = "Error al guardar la configuración.";
                console.error(error);
            }
        },
        
        async fetchDashboardData(dates) {
            if ((this.selectedConnectionId === null || this.selectedConnectionId === undefined) || !dates.startDate || !dates.endDate) return;

            this.isLoading = true;
            this.error = null;
            try {
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