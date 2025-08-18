<template>
  <DashboardCard :title="title">
    <li v-if="!items || items.length === 0" class="list-group-item">
      No hay datos
    </li>
    <li v-for="item in items" :key="item.Name" class="list-group-item d-flex justify-content-between">
      <span>{{ item.Name }}</span>
      <strong>{{ formatCurrency(item.Value) }}</strong>
    </li>
  </DashboardCard>
</template>

<script setup>
import DashboardCard from './DashboardCard.vue';

// Helper para formatear los números como moneda
const formatCurrency = (value) => {
  return new Intl.NumberFormat('es-VE', { style: 'currency', currency: 'VES' }).format(value);
};

defineProps({
  title: String,
  items: Array,
});
</script>