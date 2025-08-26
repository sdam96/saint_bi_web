<template>
  <div>
    <h1 class="text-center mb-4">Análisis de Canasta de Mercado</h1>
    <p class="text-center text-muted mb-4">
      Descubre qué productos se compran juntos con más frecuencia. Selecciona un rango de fechas para analizar tendencias estacionales.
    </p>

    <DateRangePicker v-model:start-date="startDate" v-model:end-date="endDate" />

    <Spinner v-if="store.isLoading" />
    <div v-if="store.error" class="alert alert-danger mt-4">{{ store.error }}</div>

    <div v-if="!store.isLoading && !store.error" class="card mt-4">
      <div v-if="store.marketBasket && store.marketBasket.length > 0" class="table-responsive">
        <table class="table table-striped table-hover mb-0">
          <thead class="table-light">
            <tr>
              <th>Si un cliente compra...</th>
              <th>Es probable que también compre...</th>
              <th class="text-end">Probabilidad (Confianza)</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in store.marketBasket" :key="index">
              <td><strong>{{ item.itemA }}</strong></td>
              <td><strong>{{ item.itemB }}</strong></td>
              <td class="text-end fw-bold text-success">{{ (item.confidence * 100).toFixed(2) }}%</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="card-body">
        <p class="text-center py-5">No se encontraron asociaciones significativas entre productos con los filtros actuales.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import { useAnalyticsStore } from '../store/analytics';
import Spinner from '../components/Spinner.vue';
import DateRangePicker from '../components/DateRangePicker.vue';

const store = useAnalyticsStore();

const formatDate = (date) => date.toISOString().split('T')[0];
const endDate = ref('');
const startDate = ref('');

const fetchData = () => {
    const dates = startDate.value && endDate.value ? { startDate: startDate.value, endDate: endDate.value } : {};
    store.fetchMarketBasket(dates);
};

onMounted(fetchData);
watch([startDate, endDate], fetchData);
</script>