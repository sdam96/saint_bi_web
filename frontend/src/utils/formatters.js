import { useDashboardStore } from '../store/dashboard';

/**
 * Formatea un valor numérico como moneda, utilizando la configuración
 * del store de dashboard.
 * @param {number} value - El número a formatear.
 * @returns {string} El valor formateado como string.
 */
export const formatCurrency = (value) => {
  const store = useDashboardStore();
  
  // Valores por defecto si la configuración no está cargada
  let locale = 'en-US';
  let currency = 'USD';

  // Si tenemos la configuración, la usamos.
  if (store.connectionSettings) {
    locale = store.connectionSettings.LocaleFormat || locale;
    currency = store.connectionSettings.CurrencyISO || currency;
  }
  
  if (typeof value !== 'number') {
    value = 0;
  }

  // Usamos la API de Internacionalización del navegador para formatear.
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: currency,
  }).format(value);
};