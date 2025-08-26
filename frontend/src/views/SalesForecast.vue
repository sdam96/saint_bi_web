<template>
  <div>
    <h1 class="text-center mb-4">Proyección de Ventas</h1>
    <p class="text-center text-muted mb-4">
      Análisis de tendencia basado en el historial de ventas diarias mediante regresión lineal.
    </p>

    <DateRangePicker v-model:start-date="startDate" v-model:end-date="endDate" />

    <Spinner v-if="store.isLoading" />
    <div v-if="store.error" class="alert alert-danger mt-4">{{ store.error }}</div>

    <div v-if="!store.isLoading && !store.error" class="mt-4">
      <div class="card">
        <div class="card-body">
          <Line v-if="store.salesForecast && store.salesForecast.historicalData.length > 0" :data="chartData" />
          <p v-else class="text-center py-5">No hay suficientes datos en el período seleccionado para generar una proyección.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, watch, computed } from 'vue';
import { useAnalyticsStore } from '../store/analytics';
import { useDashboardStore } from '../store/dashboard';
import DateRangePicker from '../components/DateRangePicker.vue';
import Spinner from '../components/Spinner.vue';
import { Line } from 'vue-chartjs'; // El componente se importa como "Line"
import { Chart as ChartJS, Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement } from 'chart.js';

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement);

const store = useAnalyticsStore();
const dashboardStore = useDashboardStore();

const formatDate = (date) => date.toISOString().split('T')[0];
const endDate = ref(formatDate(new Date()));
const startDate = ref(formatDate(new Date(new Date().setDate(new Date().getDate() - 90))));

const fetchData = () => {
  store.fetchSalesForecast({
    startDate: startDate.value,
    endDate: endDate.value,
  });
};

onMounted(fetchData);
watch([startDate, endDate], fetchData);

const chartData = computed(() => {
  if (!store.salesForecast) return {};
  
  const locale = dashboardStore.connectionSettings?.LocaleFormat || 'es-VE';
  const labels = store.salesForecast.historicalData.map(d => new Date(d.date).toLocaleDateString(locale));
  
  return {
    labels,
    datasets: [
      {
        label: 'Ventas Históricas Diarias',
        backgroundColor: '#0d6efd',
        borderColor: '#0d6efd',
        data: store.salesForecast.historicalData.map(d => d.sales),
        tension: 0.1,
      },
      {
        label: 'Línea de Tendencia',
        backgroundColor: '#dc3545',
        borderColor: '#dc3545',
        data: store.salesForecast.trendLine.map(d => d.sales),
        borderDash: [5, 5],
        pointRadius: 0,
      }
    ]
  };
});
</script>