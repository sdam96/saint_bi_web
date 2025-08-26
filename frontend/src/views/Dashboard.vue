<template>
  <div class="dashboard">
    <div class="text-center mb-4">
      <h1 class="animated-fade-in">Dashboard Gerencial</h1>
    </div>

    <DashboardSelector 
      :connections="store.connections"
      :selected-id="store.selectedConnectionId"
      @connection-selected="handleConnectionSelect"
      class="animated-fade-in"
    />

    <DateRangePicker
      v-if="store.selectedConnectionId !== null && store.selectedConnectionId !== undefined"
      v-model:start-date="startDate"
      v-model:end-date="endDate"
      class="animated-fade-in"
    />
    
    <div v-if="store.error" class="alert alert-danger mt-4 text-center">{{ store.error }}</div>

    <Spinner v-if="store.isLoading" />

    <div v-if="store.summaryData && !store.isLoading" class="mt-4">
      <ul class="nav nav-tabs nav-fill mb-3">
        <li class="nav-item">
          <a class="nav-link" :class="{ active: activeTab === 'summary' }" @click.prevent="activeTab = 'summary'" href="#">
            Resumen Principal
          </a>
        </li>
        <li class="nav-item">
          <a class="nav-link" :class="{ active: activeTab === 'rankings' }" @click.prevent="activeTab = 'rankings'" href="#">
            Rankings Top 5
          </a>
        </li>
      </ul>

      <div class="tab-content">
        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'summary' }">
          <SummaryDisplay 
            :summary="store.summaryData"
            :start-date="startDate"
            :end-date="endDate"
          />
        </div>
        
        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'rankings' }">
          <div v-if="store.summaryData.currentPeriod" class="row row-cols-1 row-cols-md-2 g-4">
            <RankList title="Top 5 Productos por Venta" :items="store.summaryData.currentPeriod.Top5ProductsBySales" />
            <RankList title="Top 5 Productos por Utilidad" :items="store.summaryData.currentPeriod.Top5ProductsByProfit" />
            <RankList title="Top 5 Clientes por Venta" :items="store.summaryData.currentPeriod.Top5ClientsBySales" />
            <RankList title="Top 5 Vendedores por Venta" :items="store.summaryData.currentPeriod.Top5SellersBySales" />
          </div>
        </div>
      </div>
    </div>
    
    <p v-if="!store.selectedConnectionId && store.connections.length > 1" class="text-center mt-4">
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
import DateRangePicker from '../components/DateRangePicker.vue';
import RankList from '../components/RankList.vue'; // <-- Importar RankList aquí

const store = useDashboardStore();
const activeTab = ref('summary'); // Estado para la pestaña activa

const formatDate = (date) => {
  const d = new Date(date);
  const year = d.getFullYear();
  let month = '' + (d.getMonth() + 1);
  let day = '' + d.getDate();
  if (month.length < 2) month = '0' + month;
  if (day.length < 2) day = '0' + day;
  return [year, month, day].join('-');
}

const endDate = ref(formatDate(new Date()));
const startDate = ref(formatDate(new Date(new Date().setDate(new Date().getDate() - 30))));

onMounted(() => {
  if (store.connections.length === 0) {
    store.fetchConnections();
  }
});

const handleConnectionSelect = (connectionId) => {
  store.selectConnection(connectionId, { 
    startDate: startDate.value, 
    endDate: endDate.value 
  });
};

watch([startDate, endDate], () => {
  if (store.selectedConnectionId !== null && store.selectedConnectionId !== undefined) {
    store.fetchDashboardData({ 
      startDate: startDate.value, 
      endDate: endDate.value 
    });
  }
});
</script>

<style scoped>
.nav-tabs {
  border-bottom: 1px solid var(--bs-border-color);
}
.nav-link {
  cursor: pointer;
  color: var(--bs-secondary);
  border: none;
}
.nav-link.active {
  color: var(--bs-body-color);
  background-color: var(--bs-secondary-bg) !important;
  border-bottom: 2px solid var(--saint-blue);
  font-weight: bold;
}
</style>