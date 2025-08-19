// src/store/drilldown.js
import { defineStore } from 'pinia';
import axios from 'axios';

// 'defineStore' crea un nuevo store. El primer argumento ('drilldown') es un ID único.
export const useDrilldownStore = defineStore('drilldown', {
    // El 'state' es una función que devuelve el estado inicial del store.
    // Aquí es donde guardaremos los datos que obtengamos de la API.
    state: () => ({
        // Almacenará la lista de transacciones (facturas, cxc, etc.)
        transactions: [],
        // Almacenará el detalle completo de una sola transacción.
        transactionDetail: null,
        // Guardaremos el título para la vista, ej: "Detalle de Ventas".
        currentTitle: '',
        // Banderas para manejar el estado de carga y los errores.
        isLoading: false,
        error: null,
    }),

    // Las 'actions' son métodos que pueden modificar el estado.
    // Aquí definimos las funciones para comunicarnos con nuestra API de Go.
    actions: {
        /**
         * Busca una lista de transacciones por tipo y rango de fechas.
         * @param {string} docType - El tipo de documento (ej. 'invoices').
         * @param {string} title - Un título descriptivo para la vista (ej. 'Ventas del Período').
         * @param {object} dates - Un objeto con { startDate, endDate }.
         */
        async fetchTransactions(docType, title, dates) {
            // Ponemos el store en estado de "cargando".
            this.isLoading = true;
            this.error = null;
            this.currentTitle = title;
            this.transactions = []; // Limpiamos datos anteriores.

            try {
                // Hacemos la llamada a nuestro nuevo endpoint GET /api/transactions.
                // Usamos el objeto 'params' de axios para añadir los query parameters a la URL.
                // La URL resultante será: /api/transactions?type=...&startDate=...&endDate=...
                const response = await axios.get('/api/transactions', {
                    params: {
                        type: docType,
                        startDate: dates.startDate,
                        endDate: dates.endDate,
                    }
                });
                // Si la llamada es exitosa, guardamos los datos en el estado.
                this.transactions = response.data;
            } catch (err) {
                // Si ocurre un error, lo guardamos para mostrarlo en la UI.
                this.error = `Error al cargar las transacciones de ${docType}.`;
                console.error(err);
            } finally {
                // Se ejecuta siempre, al final del try/catch.
                // Dejamos de mostrar el spinner de carga.
                this.isLoading = false;
            }
        },

        /**
         * Busca el detalle completo de una única transacción.
         * @param {string} docType - El tipo de documento (ej. 'invoice').
         * @param {string} docId - El ID del documento (ej. '123').
         */
        async fetchTransactionDetail(docType, docId) {
            this.isLoading = true;
            this.error = null;
            this.transactionDetail = null;

            try {
                // Hacemos la llamada al endpoint de detalle: GET /api/transaction/{type}/{id}
                const response = await axios.get(`/api/transaction/${docType}/${docId}`);
                this.transactionDetail = response.data;
            } catch (err) {
                this.error = `Error al cargar el detalle para ${docType} con ID ${docId}.`;
                console.error(err);
            } finally {
                this.isLoading = false;
            }
        },
    },
});