<template>
  <div class="card h-100">
    <div class="card-body text-center">
      <h6 class="card-subtitle mb-2 text-muted">{{ title }}</h6>
      <h4 class="card-title">{{ formatNumber(kpi.value) }}</h4>
      <p class="card-text" :class="isPositive ? 'text-success' : 'text-danger'">
        <span v-if="kpi.percentageChange !== 0">
          <svg v-if="isPositive" width="1em" height="1em" viewBox="0 0 16 16" class="bi bi-arrow-up-short" fill="currentColor" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M8 12a.5.5 0 0 0 .5-.5V5.707l2.146 2.147a.5.5 0 0 0 .708-.708l-3-3a.5.5 0 0 0-.708 0l-3 3a.5.5 0 1 0 .708.708L7.5 5.707V11.5a.5.5 0 0 0 .5.5z"/></svg>
          <svg v-else width="1em" height="1em" viewBox="0 0 16 16" class="bi bi-arrow-down-short" fill="currentColor" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M8 4a.5.5 0 0 1 .5.5v5.793l2.146-2.147a.5.5 0 0 1 .708.708l-3 3a.5.5 0 0 1-.708 0l-3-3a.5.5 0 1 1 .708-.708L7.5 10.293V4.5A.5.5 0 0 1 8 4z"/></svg>
          {{ kpi.percentageChange.toFixed(2) }}%
        </span>
        <small class="d-block text-muted">vs {{ formatNumber(kpi.previousValue) }}</small>
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  title: String,
  kpi: Object, // Espera un objeto como { value, previousValue, percentageChange }
  higherIsBetter: {
    type: Boolean,
    default: true,
  },
});

const isPositive = computed(() => {
  if (props.higherIsBetter) {
    return props.kpi.percentageChange >= 0;
  }
  return props.kpi.percentageChange <= 0;
});

const formatNumber = (value) => {
  if (typeof value !== 'number') return '0,00';
  return new Intl.NumberFormat('es-VE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value);
};
</script>

<style scoped>
.card-title {
  font-size: 1.75rem;
  font-weight: 500;
}
.card-text svg {
  vertical-align: middle;
}
</style>