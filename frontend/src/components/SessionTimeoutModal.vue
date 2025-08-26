<template>
  <div class="modal fade" id="sessionTimeoutModal" tabindex="-1" aria-labelledby="sessionTimeoutModalLabel" aria-hidden="true" data-bs-backdrop="static" data-bs-keyboard="false">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="sessionTimeoutModalLabel">Su sesión está a punto de expirar</h5>
          </div>
        <div class="modal-body">
          <p>Tu sesión se cerrará automáticamente en <strong class="text-danger">{{ countdown }}</strong> segundos por inactividad.</p>
          <p>¿Deseas extender tu sesión?</p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="logout">Cerrar Sesión</button>
          <button type="button" class="btn btn-primary" @click="extendSession">Extender Sesión</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// 'defineProps' y 'defineEmits' son compiladores de macros de Vue para definir
// las propiedades que el componente recibe y los eventos que puede emitir.
defineProps({
  countdown: Number, // Recibe el número de segundos restantes
});

const emit = defineEmits(['extend', 'logout']); // Define los eventos que el padre puede escuchar

// Función que emite el evento 'extend' cuando el usuario hace clic en el botón.
const extendSession = () => {
  emit('extend');
};

// Función que emite el evento 'logout' cuando el usuario hace clic.
const logout = () => {
  emit('logout');
};
</script>