import { ref, onMounted, onUnmounted } from 'vue';
import { useAuthStore } from '../store/auth';
import { Modal } from 'bootstrap';

/**
 * Composable para gestionar la expiración de la sesión del usuario.
 * Muestra un modal de advertencia y maneja la extensión o el cierre de sesión.
 */
export function useSessionManager() {
  const authStore = useAuthStore();
  
  // -- CONSTANTES DE CONFIGURACIÓN --
  // Duración total de la sesión en milisegundos (20 minutos, igual que el backend).
  const SESSION_DURATION = 20 * 60 * 1000; 
  // Momento en que se mostrará el modal de advertencia (2 minutos antes de expirar).
  const WARNING_TIME = 2 * 60 * 1000; 
  
  // -- ESTADO REACTIVO --
  // Referencia al objeto del modal de Bootstrap.
  const sessionModal = ref(null);
  // Contador de segundos para mostrar en el modal.
  const countdown = ref(WARNING_TIME / 1000);
  
  // -- TEMPORIZADORES --
  // ID de los temporizadores para poder limpiarlos después.
  let warningTimer = null;
  let logoutTimer = null;
  let countdownInterval = null;

  // -- FUNCIONES DE LÓGICA --

  /** Inicia los temporizadores cuando el usuario inicia sesión o refresca la página. */
  const startTimers = () => {
    // Limpiamos cualquier temporizador anterior para evitar duplicados.
    clearTimers();

    // 1. Temporizador de Advertencia: Se activará cuando falte `WARNING_TIME` para expirar.
    warningTimer = setTimeout(() => {
      showModal();
    }, SESSION_DURATION - WARNING_TIME);

    // 2. Temporizador de Logout: Se activará cuando la sesión expire por completo.
    logoutTimer = setTimeout(() => {
      authStore.logout();
    }, SESSION_DURATION);
  };
  
  /** Limpia todos los temporizadores activos. */
  const clearTimers = () => {
    clearTimeout(warningTimer);
    clearTimeout(logoutTimer);
    clearInterval(countdownInterval);
  };

  /** Muestra el modal y comienza la cuenta regresiva. */
  const showModal = () => {
    countdown.value = WARNING_TIME / 1000; // Reinicia el contador
    if (sessionModal.value) {
      sessionModal.value.show();
      
      // Inicia un intervalo que descuenta un segundo cada segundo.
      countdownInterval = setInterval(() => {
        countdown.value--;
        if (countdown.value <= 0) {
          clearInterval(countdownInterval);
        }
      }, 1000);
    }
  };

  /** Oculta el modal. */
  const hideModal = () => {
    if (sessionModal.value) {
      sessionModal.value.hide();
    }
    clearInterval(countdownInterval);
  };

  /** Maneja la acción de extender la sesión. */
  const handleExtend = async () => {
    hideModal();
    const success = await authStore.extendSession();
    if (success) {
      // Si la sesión se extendió exitosamente, reiniciamos los temporizadores.
      startTimers(); 
    } else {
      // Si falla la extensión, cerramos la sesión para evitar inconsistencias.
      authStore.logout();
    }
  };

  // -- CICLO DE VIDA --
  
  // 'onMounted' se ejecuta cuando el componente que usa este composable se monta en el DOM.
  onMounted(() => {
    // Creamos una nueva instancia del modal de Bootstrap y la guardamos.
    const modalElement = document.getElementById('sessionTimeoutModal');
    if (modalElement) {
      sessionModal.value = new Modal(modalElement);
    }
    // Si el usuario ya está autenticado (refrescó la página), iniciamos los temporizadores.
    if (authStore.isAuthenticated) {
      startTimers();
    }
  });

  // 'onUnmounted' se ejecuta justo antes de que el componente se desmonte.
  // Es crucial limpiar los temporizadores para evitar fugas de memoria.
  onUnmounted(() => {
    clearTimers();
  });

  // Devolvemos las variables y funciones que el componente necesitará.
  return {
    countdown,
    handleExtend,
    logout: authStore.logout,
    startTimers, // Exponemos para llamarla después del login
  };
}