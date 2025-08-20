<template>
  <div>
    <div class="mb-4">
      <a @click="$router.back()" class="btn btn-outline-secondary" style="cursor: pointer;">
        &larr; Volver a la Lista
      </a>
    </div>

    <Spinner v-if="store.isLoading" />
    <div v-if="store.error" class="alert alert-danger">{{ store.error }}</div>

    <div v-if="detail && !store.isLoading">
      <div class="card mb-4">
        <div class="card-header bg-dark text-white">
          <h2 class="mb-0">Detalle de Factura #{{ detail.document.numerod }}</h2>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <p><strong>Fecha de Emisión:</strong> {{ formatDate(detail.document.fechae) }}</p>
              <p><strong>Monto Total:</strong> <span class="fw-bold fs-5 text-success">{{ formatCurrency(detail.document.mtototal) }}</span></p>
            </div>
            <div class="col-md-6">
              <p><strong>Crédito:</strong> {{ formatCurrency(detail.document.credito) }}</p>
              <p><strong>Contado:</strong> {{ formatCurrency(detail.document.contado) }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="row">
        <div class="col-md-6">
          <div class="card mb-4 drilldown-card" @click="viewEntityDetail('customer', detail.customer.codclie)">
            <div class="card-header">Información del Cliente</div>
            <div class="card-body" v-if="detail.customer">
              <p><strong>Código:</strong> {{ detail.customer.codclie }}</p>
              <p><strong>Nombre:</strong> {{ detail.customer.descrip }}</p>
              <p><strong>RIF:</strong> {{ detail.customer.id3 }}</p>
            </div>
            <div v-else class="card-body"><p class="text-muted">No se encontró información del cliente.</p></div>
          </div>
        </div>
        <div class="col-md-6">
          <div class="card mb-4 drilldown-card" @click="viewEntityDetail('seller', detail.seller.codvend)">
            <div class="card-header">Información del Vendedor</div>
            <div class="card-body" v-if="detail.seller">
              <p><strong>Código:</strong> {{ detail.seller.codvend }}</p>
              <p><strong>Nombre:</strong> {{ detail.seller.descrip }}</p>
            </div>
            <div v-else class="card-body"><p class="text-muted">No se encontró información del vendedor.</p></div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header"><h3 class="mb-0">Productos en la Factura</h3></div>
        <div class="table-responsive">
          <table class="table table-striped table-hover mb-0">
            <thead>
              <tr>
                <th>Código</th>
                <th>Descripción</th>
                <th class="text-end">Cantidad</th>
                <th class="text-end">Precio Unitario</th>
                <th class="text-end">Total Ítem</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in detail.items" :key="item.nrolinea" class="drilldown-card" @click="viewEntityDetail('product', item.coditem)">
                <td>{{ item.coditem }}</td>
                <td>{{ item.descrip1 }}</td>
                <td class="text-end">{{ item.cantidad.toFixed(2) }}</td>
                <td class="text-end">{{ formatCurrency(item.precio) }}</td>
                <td class="text-end fw-bold">{{ formatCurrency(item.totalitem) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// 'onMounted' es un hook del ciclo de vida de Vue que se ejecuta cuando el componente se ha montado en el DOM.
// 'computed' nos permite crear propiedades reactivas que se actualizan automáticamente cuando sus dependencias cambian.
import { onMounted, computed } from 'vue';
// 'useRoute' nos da acceso al objeto de la ruta actual (parámetros, query, etc.).
// 'useRouter' nos da acceso a la instancia del enrutador para navegar programáticamente.
import { useRoute, useRouter } from 'vue-router';
// Importamos nuestro store de Pinia para manejar el estado del drilldown.
import { useDrilldownStore } from '../store/drilldown';
// Importamos el componente Spinner para mostrar una indicación de carga.
import Spinner from '../components/Spinner.vue';
import { formatCurrency } from '../utils/formatters';

const store = useDrilldownStore();
const route = useRoute();
const router = useRouter();

// Leemos el ID de la factura desde los parámetros de la URL.
// Por ejemplo, para la ruta '/invoice/123', 'route.params.id' será '123'.
const invoiceId = route.params.id;

// Usamos una propiedad computada para acceder al detalle de la transacción desde el store.
// Esto es más limpio y eficiente que acceder a 'store.transactionDetail' directamente en el template.
const detail = computed(() => store.transactionDetail);

// Cuando el componente se monta en la página, llamamos a la acción del store para buscar los datos de esta factura.
onMounted(() => {
  store.fetchTransactionDetail('invoice', invoiceId);
});

// --- Funciones auxiliares de formato ---

// Formatea un string de fecha (ej. "2023-10-26 11:23:09") a un formato local legible.
const formatDate = (dateString) => {
  if (!dateString) return 'N/A';
  // Tomamos solo la parte de la fecha antes del espacio.
  const date = new Date(dateString.split(' ')[0]);
  return date.toLocaleDateString('es-VE');
};

// --- Navegación al siguiente nivel ---
const viewEntityDetail = (type, id) => {
  if(!id) return;

  let routeName = '';
  switch (type) {
    case 'customer':
      routeName = 'CustomerDetail';
      break;

    case 'seller':
      routeName = 'SellerDetail';
      break;
    
    case 'product':
      routeName = 'ProductDetail';
      break;
    default:
      break;
  }

  router.push({name: routeName, params: {id: id} });
};

</script>

<style scoped>
/* 'scoped' asegura que estos estilos solo se apliquen a este componente */
.drilldown-card {
  cursor: pointer;
}
.drilldown-card:hover {
  background-color: #f8f9fa; /* Un ligero resaltado al pasar el ratón */
}
</style>