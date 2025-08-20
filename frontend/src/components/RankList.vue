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
import { formatCurrency } from '../utils/formatters'

defineProps({
  title: String,
  items: Array,
});
</script>