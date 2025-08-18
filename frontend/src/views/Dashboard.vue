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
import { onMounted } from 'vue';
import { useDashboardStore } from '../store/dashboard';
import DashboardSelector from '../components/DashboardSelector.vue';
import SummaryDisplay from '../components/SummaryDisplay.vue';
import Spinner from '../components/Spinner.vue';

const store = useDashboardStore();

// Cuando el componente se monta por primera vez, cargamos las conexiones.
onMounted(() => {
  store.fetchConnections();
});

// Esta función se llama cuando el selector emite un evento.
const handleConnectionSelect = (connectionId) => {
  store.selectConnection(connectionId);
};
</script>