<template>
  <div>
    <h1>Gestionar Usuarios</h1>
    <div class="row g-5">
      <div class="col-md-8">
        <h2>Usuarios Existentes</h2>
        <div v-if="store.isLoading" class="text-center">Cargando...</div>
        <div v-else-if="store.error" class="alert alert-danger">{{ store.error }}</div>
        <div v-else class="list-group">
          <div v-if="store.users.length === 0" class="list-group-item text-center">
            No hay usuarios registrados.
          </div>
          <div v-for="user in store.users" :key="user.ID" class="list-group-item">
            {{ user.Username }}
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <h2>Agregar Nuevo Usuario</h2>
        <form @submit.prevent="handleAddUser">
          <div class="mb-3">
            <label for="username" class="form-label">Usuario:</label>
            <input type="text" id="username" v-model="newUser.username" class="form-control" required>
          </div>
          <div class="mb-3">
            <label for="password" class="form-label">Clave:</label>
            <input type="password" id="password" v-model="newUser.password" class="form-control" required>
          </div>
          <button type="submit" class="btn btn-primary">Crear Usuario</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useAdminStore } from '../store/admin';

const store = useAdminStore();

const newUser = ref({
  username: '',
  password: '',
});

onMounted(() => {
  store.fetchUsers();
});

const handleAddUser = async () => {
  await store.addUser(newUser.value);
  // Limpiamos el formulario
  newUser.value.username = '';
  newUser.value.password = '';
};
</script>