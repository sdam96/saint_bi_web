<template>
  <div>
    <h1>Gestionar Conexiones</h1>
    <div class="row g-5">
      <div class="col-md-8">
        <h2>Conexiones Existentes</h2>
        <div v-if="store.isLoading" class="text-center">Cargando...</div>
        <div v-else-if="store.error" class="alert alert-danger">{{ store.error }}</div>
        <div v-else class="table-responsive">
          <table class="table table-striped table-hover">
            <thead class="table-light">
              <tr>
                <th>Alias</th>
                <th>URL</th>
                <th>Usuario API</th>
                <th>Acción</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="store.connections.length === 0">
                <td colspan="4" class="text-center">No hay conexiones configuradas.</td>
              </tr>
              <tr v-for="conn in store.connections" :key="conn.ID">
                <td>{{ conn.Alias }}</td>
                <td>{{ conn.ApiURL }}</td>
                <td>{{ conn.ApiUser }}</td>
                <td>
                  <button @click="store.deleteConnection(conn.ID)" class="btn btn-sm btn-outline-danger">
                    Eliminar
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="col-md-4">
        <h2>Agregar Conexión</h2>
        <form @submit.prevent="handleAddConnection">
          <div class="mb-3">
            <label class="form-label">Alias:</label>
            <input type="text" v-model="newConnection.Alias" class="form-control" required>
          </div>
          <div class="mb-3">
            <label class="form-label">URL API:</label>
            <input type="text" v-model="newConnection.ApiURL" class="form-control" required placeholder="http://localhost:6163">
          </div>
          <div class="mb-3">
            <label class="form-label">Usuario API:</label>
            <input type="text" v-model="newConnection.ApiUser" class="form-control" required>
          </div>
          <div class="mb-3">
            <label class="form-label">Clave API:</label>
            <input type="password" v-model="newConnection.ApiPassword" class="form-control" required>
          </div>
          <div class="mb-3">
            <label class="form-label">ID Configuración:</label>
            <input type="number" v-model.number="newConnection.ConfigID" class="form-control" required>
          </div>
          <div class="mb-3">
            <label class="form-label">Refrescar cada (segundos):</label>
            <input type="number" v-model.number="newConnection.RefreshSeconds" class="form-control" required>
          </div>
          <button type="submit" class="btn btn-primary">Guardar Conexión</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useAdminStore } from '../store/admin';

const store = useAdminStore();

// Usamos 'ref' para el estado local del formulario
const newConnection = ref({
  Alias: '',
  ApiURL: '',
  ApiUser: '',
  ApiPassword: '',
  ConfigID: 1,
  RefreshSeconds: 60,
});

onMounted(() => {
  store.fetchConnections();
});

const handleAddConnection = async () => {
  await store.addConnection(newConnection.value);
  // Limpiamos el formulario después de enviar
  newConnection.value = {
    Alias: '',
    ApiURL: '',
    ApiUser: '',
    ApiPassword: '',
    ConfigID: 1,
    RefreshSeconds: 60,
  };
};
</script>