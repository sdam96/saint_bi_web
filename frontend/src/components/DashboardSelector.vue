<template>
  <div class="row justify-content-center">
    <div class="col-md-6">
      <label for="connection_id_select" class="form-label">Seleccionar Conexión:</label>
      <select 
        id="connection_id_select" 
        class="form-select"
        :value="selectedId"
        @change="$emit('connection-selected', parseInt($event.target.value))"
      >
        <option value="">-- Seleccione una conexión --</option>
        <option v-for="conn in connections" :key="conn.ID" :value="conn.ID">
          {{ conn.Alias }}
        </option>
      </select>
    </div>
  </div>
</template>

<script setup>
// Definimos las propiedades (props) y eventos (emits) que el componente acepta.
defineProps({
  connections: Array,
  selectedId: Number,
});
defineEmits(['connection-selected']);
</script>