import { defineStore } from "pinia";
import axios from 'axios';

export const useAnalyticsStore = defineStore('analytics', {
    state: () => ({
        salesForecast: null,
        marketBasket: [],
        isLoading: false,
        error: null,
    }),
    actions: {
        async fetchSalesForecast(dates) {
            this.isLoading = true;
            this.error = null;
            try {
                const response = await axios.get('/api/analytics/sales-forecast', {
                    params: dates
                });

                this.salesForecast = response.data;
            } catch (err) {
                this.error = 'Error al cargar la proyeccion de ventas.';
                console.error(err);
            } finally {
                this.isLoading = false;
            }
        },

        async fetchMarketBasket(dates) {
            this.isLoading = true;
            this.error = null;
            try {
                const response = await axios.get('/api/analytics/market-basket', {
                    params: dates
                });
                this.marketBasket = response.data;
            } catch (error) {
                this.error = 'Error al cargar el analisis de canasta.';
                console.error(err);
                
            } finally {
                this.isLoading = false;
            }
        },
    },
});