<template>
  <div>
    <div class="mb-4">
      <a @click="$router.back()" class="btn btn-outline-secondary" style="cursor: pointer;">
        &larr; Volver
      </a>
    </div>

    <h1 class="mb-4">{{ title }}</h1>

    <Spinner v-if="store.isLoading" />
    <div v-if="store.error" class="alert alert-danger">{{ store.error }}</div>

    <div v-if="!store.isLoading && !store.error && formattedTransactions.length > 0" class="card">
      <div class="table-responsive">
        <table class="table table-hover table-striped mb-0">
          <thead>
            <tr>
              <th>Número</th>
              <th>{{ descriptionHeader }}</th>
              <th>Fecha</th>
              <th class="text-end">Monto Total</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="tx in formattedTransactions" :key="tx.key" @click="viewDetail(tx.original)" class="transaction-row">
              <td>{{ tx.number }}</td>
              <td>{{ tx.description }}</td>
              <td>{{ tx.date }}</td>
              <td class="text-end">{{ tx.amount }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    
    <div v-if="!store.isLoading && !store.error && (!store.transactions || store.transactions.length === 0)" class="alert alert-info text-center">
      No se encontraron transacciones para el período seleccionado.
    </div>
  </div>
</template>

<script setup>
import { onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useDrilldownStore } from '../store/drilldown';
import Spinner from '../components/Spinner.vue';

const store = useDrilldownStore();
const route = useRoute();
const router = useRouter();

// Leemos los parámetros de la URL para saber qué buscar
const type = route.params.type;
const { startDate, endDate, title } = route.query;

// Cuando el componente se carga, le pedimos al store que busque los datos
onMounted(() => {
  store.fetchTransactions(type, title, { startDate, endDate });
});

// --- PROPIEDADES COMPUTADAS PARA LA VISTA ---

// Determina el título de la segunda columna según el tipo de transacción
const descriptionHeader = computed(() => {
  if (type.startsWith('invoices') || type === 'receivables') return 'Cliente';
  if (type === 'payables') return 'Proveedor';
  return 'Descripción';
});

// Transforma la data cruda de la API en una lista unificada para la tabla
const formattedTransactions = computed(() => {
  // Comprobación más robusta para asegurar que 'store.transactions' es un array.
  if (!Array.isArray(store.transactions)) return [];
  
  return store.transactions.map(tx => {
    // --- CORRECCIÓN CLAVE ---
    // Accedemos a las propiedades del objeto JSON usando los nombres en minúsculas
    // que envía el backend de Go (ej. 'mtototal' en lugar de 'MtoTotal').
    const amount = tx.mtototal !== undefined ? tx.mtototal : tx.monto;
    const description = tx.descrip;
    const number = tx.numerod;
    const date = tx.fechae;
    // --- FIN DE LA CORRECCIÓN ---

    return {
      key: tx.id || tx.nrounico || number, // Usamos claves en minúsculas
      number: number || 'N/A',
      description: description || 'No especificado',
      date: formatDate(date),
      amount: formatCurrency(amount),
      original: tx // Guardamos el objeto original para la navegación
    };
  });
});

// --- Funciones auxiliares ---
const formatCurrency = (value) => {
  if (typeof value !== 'number') return 'Bs. 0,00';
  return new Intl.NumberFormat('es-VE', { style: 'currency', currency: 'VES' }).format(value);
};

const formatDate = (dateString) => {
  if (!dateString) return 'N/A';
  // El formato de fecha de la API es "AAAA-MM-DD HH:mm:ss", lo acortamos a solo la fecha.
  const date = new Date(dateString.split(' ')[0]);
  return date.toLocaleDateString('es-VE');
};

// Navega al siguiente nivel de detalle
const viewDetail = (originalTransaction) => {
  // Solo navegamos si es una factura, ya que es la única vista de detalle que hemos creado
  if (type.startsWith('invoices')) {
    router.push({
        name: 'InvoiceDetail',
        params: { id: originalTransaction.numerod } // Usamos la propiedad en minúscula
    });
  }
};
</script>

<style scoped>
.transaction-row {
  cursor: pointer;
}
</style>