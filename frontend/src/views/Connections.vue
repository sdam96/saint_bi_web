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
              <tr v-if="filteredConnections.length === 0">
                <td colspan="4" class="text-center">No hay conexiones configuradas.</td>
              </tr>
              <tr v-for="conn in filteredConnections" :key="conn.ID">
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
            <label for="alias" class="form-label">Alias</label>
            <input type="text" id="alias" v-model="newConnection.Alias" class="form-control" required>
          </div>
          <div class="mb-3">
            <label for="api_url" class="form-label">URL de la API</label>
            <input type="url" id="api_url" v-model="newConnection.ApiURL" class="form-control" required>
          </div>
          <div class="mb-3">
            <label for="api_user" class="form-label">Usuario API</label>
            <input type="text" id="api_user" v-model="newConnection.ApiUser" class="form-control" required>
          </div>
          <div class="mb-3">
            <label for="api_password" class="form-label">Clave API</label>
            <input type="password" id="api_password" v-model="newConnection.ApiPassword" class="form-control" required>
          </div>
          <div class="mb-3">
            <label for="config_id" class="form-label">ID de Configuración</label>
            <input type="number" id="config_id" v-model.number="newConnection.ConfigID" class="form-control" required>
          </div>
          <div class="mb-3">
            <label for="refresh_seconds" class="form-label">Tiempo de Refresco (segundos)</label>
            <input type="number" id="refresh_seconds" v-model.number="newConnection.RefreshSeconds" class="form-control" required>
          </div>
          <button type="submit" class="btn btn-primary">Agregar Conexión</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue';
import { useAdminStore } from '../store/admin';

const store = useAdminStore();

const newConnection = ref({
  Alias: '',
  ApiURL: '',
  ApiUser: '',
  ApiPassword: '',
  ConfigID: 1,
  RefreshSeconds: 300,
});

const filteredConnections = computed(() => {
  return store.connections.filter(c => c.ID !== 0); 
});

onMounted(() => {
  store.fetchConnections();
});

const handleAddConnection = async () => {
  await store.addConnection(newConnection.value);
  newConnection.value = {
    Alias: '', ApiURL: '', ApiUser: '', ApiPassword: '',
    ConfigID: 1, RefreshSeconds: 300,
  };
};
</script>