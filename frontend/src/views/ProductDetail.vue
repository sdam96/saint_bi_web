<template>
  <div>
    <div class="mb-4">
      <a @click="$router.back()" class="btn btn-outline-secondary" style="cursor: pointer;">
        &larr; Volver
      </a>
    </div>

    <Spinner v-if="store.isLoading" />
    <div v-if="store.error" class="alert alert-danger">{{ store.error }}</div>

    <div v-if="detail && !store.isLoading" class="card">
      <div class="card-header bg-dark text-white">
        <h2 class="mb-0">Detalle del Producto</h2>
      </div>
      <div class="card-body">
        <p><strong>Código:</strong> {{ detail.codprod }}</p>
        <p><strong>Descripción:</strong> {{ detail.descrip }}</p>
        <p><strong>Marca:</strong> {{ detail.marca }}</p>
        <p><strong>Existencia:</strong> {{ detail.existen }}</p>
        <p><strong>Precio 1:</strong> {{ formatCurrency(detail.precio1) }}</p>
        <p><strong>Costo Actual:</strong> {{ formatCurrency(detail.costact) }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useDrilldownStore } from '../store/drilldown';
import Spinner from '../components/Spinner.vue';
import { formatCurrency } from '../utils/formatters';

const store = useDrilldownStore();
const route = useRoute();
const entityId = route.params.id;

const detail = computed(() => store.entityDetail);

onMounted(() => {
  store.fetchEntityDetail('product', entityId);
});

</script>