<template>
  <div>
    <h1>Configuración de Visualización</h1>
    <p class="text-muted">
      Define la moneda y el formato de números para la conexión seleccionada.
    </p>

    <Spinner v-if="store.isLoading" />
    <div v-if="store.error" class="alert alert-danger">{{ store.error }}</div>

    <div v-if="settings && !store.isLoading" class="row justify-content-center mt-4">
      <div class="col-md-6">
        <div class="card">
          <div class="card-body">
            <form @submit.prevent="handleSave">
              <div class="mb-3">
                <label for="locale" class="form-label">Formato Regional (Locale)</label>
                <input type="text" id="locale" class="form-control" v-model="settings.LocaleFormat" placeholder="ej: es-VE, en-US, pt-BR">
                <div class="form-text">Define el separador de miles y decimales.</div>
              </div>
              <div class="mb-3">
                <label for="currency" class="form-label">Código de Moneda (ISO 4217)</label>
                <input type="text" id="currency" class="form-control" v-model="settings.CurrencyISO" placeholder="ej: VES, USD, EUR">
                <div class="form-text">Define el símbolo de la moneda a mostrar.</div>
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

const store = useDashboardStore();
const isSaving = ref(false);
const saveSuccess = ref(false);

const settings = computed(() => store.connectionSettings);

onMounted(() => {
  if (!store.connectionSettings) {
    store.fetchConnectionSettings();
  }
});

const handleSave = async () => {
  isSaving.value = true;
  saveSuccess.value = false;
  await store.saveConnectionSettings(settings.value);
  isSaving.value = false;
  saveSuccess.value = true;
  setTimeout(() => saveSuccess.value = false, 3000); // Oculta el mensaje después de 3 segundos
};
</script>