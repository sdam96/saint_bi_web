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
        <h2 class="mb-0">Detalle del Vendedor</h2>
      </div>
      <div class="card-body">
        <p><strong>Código:</strong> {{ detail.codvend }}</p>
        <p><strong>Nombre:</strong> {{ detail.descrip }}</p>
        <p><strong>Teléfono:</strong> {{ detail.telef }}</p>
        <p><strong>Email:</strong> {{ detail.email }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useDrilldownStore } from '../store/drilldown';
import Spinner from '../components/Spinner.vue';

const store = useDrilldownStore();
const route = useRoute();
const entityId = route.params.id;

const detail = computed(() => store.entityDetail);

onMounted(() => {
  store.fetchEntityDetail('seller', entityId);
});
</script>