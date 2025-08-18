// src/views/Dashboard.vue
<template>
  <div class="dashboard">
    <div class="text-center mb-4">
      <h1>Dashboard Gerencial</h1>
    </div>

    <DashboardSelector 
      :connections="store.connections"
      :selected-id="store.selectedConnectionId"
      @connection-selected="handleConnectionSelect"
    />

    <DateRangePicker
      v-if="store.selectedConnectionId"
      v-model:start-date="startDate"
      v-model:end-date="endDate"
    />

    <div v-if="store.error" class="alert alert-danger mt-4 text-center">
      {{ store.error }}
    </div>

    <Spinner v-if="store.isLoading" />

    <SummaryDisplay v-if="store.summaryData && !store.isLoading" :summary="store.summaryData" />
    
    <p v-if="!store.selectedConnectionId && !store.isLoading" class="text-center mt-4">
      Seleccione una conexión para ver los datos.
    </p>

  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import { useDashboardStore } from '../store/dashboard';
import DashboardSelector from '../components/DashboardSelector.vue';
import SummaryDisplay from '../components/SummaryDisplay.vue';
import Spinner from '../components/Spinner.vue';
// Importamos el nuevo componente
import DateRangePicker from '../components/DateRangePicker.vue';

const store = useDashboardStore();

// --- Lógica para el manejo de Fechas ---

// Función para formatear fechas a AAAA-MM-DD
const formatDate = (date) => {
  const d = new Date(date);
  const year = d.getFullYear();
  let month = '' + (d.getMonth() + 1);
  let day = '' + d.getDate();
  if (month.length < 2) month = '0' + month;
  if (day.length < 2) day = '0' + day;
  return [year, month, day].join('-');
}

// Estado reactivo para las fechas, con valores por defecto
const endDate = ref(formatDate(new Date()));
const startDate = ref(formatDate(new Date(new Date().setDate(new Date().getDate() - 30))));

// Cuando el componente se monta, cargamos las conexiones.
onMounted(() => {
  store.fetchConnections();
});

// Cuando se selecciona una conexión, se cargan los datos con las fechas actuales.
const handleConnectionSelect = (connectionId) => {
  store.selectConnection(connectionId, { 
    startDate: startDate.value, 
    endDate: endDate.value 
  });
};

// Observador (watch): Cada vez que 'startDate' o 'endDate' cambian,
// se vuelven a pedir los datos del dashboard.
watch([startDate, endDate], () => {
  if (store.selectedConnectionId) {
    store.fetchDashboardData({ 
      startDate: startDate.value, 
      endDate: endDate.value 
    });
  }
});
</script>