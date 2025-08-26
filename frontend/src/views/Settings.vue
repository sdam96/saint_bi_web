<template>
  <div>
    <h1>Configuración de Visualización</h1>
    <p class="text-muted">
      Define la moneda y el formato de números para la conexión 
      <strong>{{ currentConnectionName }}</strong>.
    </p>

    <div v-if="store.selectedConnectionId === null" class="alert alert-warning">
      Por favor, seleccione una conexión desde el Dashboard para configurar.
    </div>
    
    <Spinner v-else-if="store.isLoading" />
    <div v-else-if="store.error" class="alert alert-danger">{{ store.error }}</div>

    <div v-else-if="settings" class="row justify-content-center mt-4">
      <div class="col-md-6">
        <div class="card">
          <div class="card-body">
            <form @submit.prevent="handleSave">
              <div class="mb-3">
                <label for="locale" class="form-label">Formato Regional (Locale)</label>
                <select id="locale" class="form-select" v-model="settings.LocaleFormat">
                  <option v-for="locale in locales" :key="locale.code" :value="locale.code">
                    {{ locale.name }}
                  </option>
                </select>
                <div class="form-text">Define el separador de miles, decimales y el símbolo de moneda.</div>
              </div>
              <div class="mb-3">
                <label for="currency" class="form-label">Código de Moneda (ISO 4217)</label>
                <input type="text" id="currency" class="form-control" v-model="settings.CurrencyISO" placeholder="ej: VES, USD, EUR">
                <div class="form-text">Define la moneda a mostrar (ej: USD, EUR).</div>
              </div>
              <button type="submit" class="btn btn-primary" :disabled="isSaving">
                {{ isSaving ? 'Guardando...' : 'Guardar Cambios' }}
              </button>
            </form>
            <div v-if="saveSuccess" class="alert alert-success mt-3">¡Configuración guardada!</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue';
import { useDashboardStore } from '../store/dashboard';
import Spinner from '../components/Spinner.vue';

const locales = ref([
  { code: 'es-VE', name: 'Español (Venezuela)' },
  { code: 'es-ES', name: 'Español (España)' },
  { code: 'es-MX', name: 'Español (México)' },
  { code: 'es-CO', name: 'Español (Colombia)' },
  { code: 'en-US', name: 'Inglés (Estados Unidos)' },
  { code: 'en-GB', name: 'Inglés (Reino Unido)' },
  { code: 'pt-BR', name: 'Portugués (Brasil)' },
]);

const store = useDashboardStore();
const isSaving = ref(false);
const saveSuccess = ref(false);

const settings = computed(() => store.connectionSettings);

// **NUEVA PROPIEDAD COMPUTADA**
// Muestra el nombre de la conexión actual.
const currentConnectionName = computed(() => {
  if (store.selectedConnectionId === null) return '';
  const conn = store.connections.find(c => c.ID === store.selectedConnectionId);
  return conn ? conn.Alias : '';
});

// **LÓGICA DE onMounted MEJORADA**
onMounted(() => {
    // Si no hay conexiones, las buscamos primero.
    if(store.connections.length === 0){
        store.fetchConnections();
    }
  // Si no hay configuración cargada Y hay una conexión seleccionada, la cargamos.
  if (!store.connectionSettings && store.selectedConnectionId !== null) {
      if(store.selectedConnectionId === 0){
          store.loadConsolidatedSettings();
      } else {
          store.fetchConnectionSettingsAPI();
      }
  }
});

const handleSave = async () => {
  isSaving.value = true;
  saveSuccess.value = false;
  await store.saveConnectionSettings(settings.value);
  isSaving.value = false;
  saveSuccess.value = true;
  setTimeout(() => saveSuccess.value = false, 3000);
};
</script>