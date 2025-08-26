<template>
  <div>
    <div class="row row-cols-1 row-cols-md-3 g-4 mb-4">
      <div class="col animated-fade-in">
        <KpiCard title="Ventas Netas" :kpi="summary.totalNetSalesComparative" />
      </div>
      <div class="col animated-fade-in" style="animation-delay: 100ms;">
        <KpiCard title="Utilidad Bruta" :kpi="summary.grossProfitComparative" />
      </div>
      <div class="col animated-fade-in" style="animation-delay: 200ms;">
        <KpiCard title="Ticket Promedio" :kpi="summary.averageTicketComparative" />
      </div>
    </div>

    <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4 mt-1">

      <DashboardCard title="Ventas y Utilidad (Período Actual)">
        <li class="list-group-item d-flex justify-content-between">
          <span>Total Ventas Netas:</span> 
          <strong>{{ formatCurrency(summary.currentPeriod.TotalNetSales) }}</strong>
        </li>
        <li 
          class="list-group-item d-flex justify-content-between" 
          :class="{ 'drilldown-row': !isConsolidatedView }"
          @click="navigateToDrilldown('invoices-cash', 'Ventas de Contado del Período')">
          <span>Ventas de Contado:</span> 
          <strong>{{ formatCurrency(summary.currentPeriod.TotalNetSalesCash) }}</strong>
        </li>
        <li 
          class="list-group-item d-flex justify-content-between" 
          :class="{ 'drilldown-row': !isConsolidatedView }"
          @click="navigateToDrilldown('invoices-credit', 'Ventas a Crédito del Período')">
          <span>Ventas a Crédito:</span> 
          <strong>{{ formatCurrency(summary.currentPeriod.TotalNetSalesCredit) }}</strong>
        </li>
        <li class="list-group-item d-flex justify-content-between"><span>Costo de Ventas:</span> <strong>{{ formatCurrency(summary.currentPeriod.CostOfGoodsSold) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Utilidad Bruta:</span> <strong>{{ formatCurrency(summary.currentPeriod.GrossProfit) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Margen Bruto:</span> <strong>{{ summary.currentPeriod.GrossProfitMargin.toFixed(2) }}%</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Ticket Promedio:</span> <strong>{{ formatCurrency(summary.currentPeriod.AverageTicket) }}</strong></li>
      </DashboardCard>

      <div class="col" :class="{ 'drilldown-card': !isConsolidatedView }" @click="navigateToDrilldown('receivables', 'Cuentas por Cobrar del Período')">
        <DashboardCard title="Cuentas por Cobrar">
          <li class="list-group-item d-flex justify-content-between"><span>Total por Cobrar:</span> <strong>{{ formatCurrency(summary.currentPeriod.TotalReceivables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Monto Vencido:</span> <strong>{{ formatCurrency(summary.currentPeriod.OverdueReceivables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Porcentaje Vencido:</span> <strong>{{ summary.currentPeriod.ReceivablePercentage.toFixed(2) }}%</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Días en la Calle:</span> <strong>{{ summary.currentPeriod.ReceivablesTurnoverDays.toFixed(0) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Clientes Activos con Deuda:</span> <strong>{{ summary.currentPeriod.ActiveClientsWithDebt }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Clientes con Deuda Vencida:</span> <strong>{{ summary.currentPeriod.TotalClientsWithOverdue }}</strong></li>
        </DashboardCard>
      </div>

      <div class="col" :class="{ 'drilldown-card': !isConsolidatedView }" @click="navigateToDrilldown('payables', 'Cuentas por Pagar del Período')">
        <DashboardCard title="Cuentas por Pagar">
           <li class="list-group-item d-flex justify-content-between"><span>Total por Pagar:</span> <strong>{{ formatCurrency(summary.currentPeriod.TotalPayables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Monto Vencido:</span> <strong>{{ formatCurrency(summary.currentPeriod.OverduePayables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Días Compras por Pagar:</span> <strong>{{ summary.currentPeriod.PayablesTurnoverDays.toFixed(0) }}</strong></li>
        </DashboardCard>
      </div>
      
      <DashboardCard title="Totales Generales">
        <li class="list-group-item d-flex justify-content-between"><span>Total Facturas Emitidas:</span> <strong>{{ summary.currentPeriod.TotalInvoices }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Clientes Activos:</span> <strong>{{ summary.currentPeriod.TotalActiveClients }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Productos Activos:</span> <strong>{{ summary.currentPeriod.TotalActiveProducts }}</strong></li>
      </DashboardCard>

      <DashboardCard title="Impuestos y Retenciones">
        <li class="list-group-item d-flex justify-content-between"><span>IVA Débito Fiscal (Ventas):</span> <strong>{{ formatCurrency(summary.currentPeriod.SalesVAT) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>IVA Crédito Fiscal (Compras):</span> <strong>{{ formatCurrency(summary.currentPeriod.PurchasesVAT) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Total IVA por Pagar:</span> <strong>{{ formatCurrency(summary.currentPeriod.VATPayable) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>IVA Retenido por Clientes:</span> <strong>{{ formatCurrency(summary.currentPeriod.SalesIVAWithheld) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>IVA Retenido a Proveedores:</span> <strong>{{ formatCurrency(summary.currentPeriod.PurchasesIVAWithheld) }}</strong></li>
      </DashboardCard>

      </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useDashboardStore } from '../store/dashboard';
import DashboardCard from './DashboardCard.vue';
import RankList from './RankList.vue'; // <-- Esta importación ya no es necesaria aquí, pero no causa daño.
import KpiCard from './KpiCard.vue';
import { formatCurrency } from '../utils/formatters';

const props = defineProps({
  summary: Object,
  startDate: String,
  endDate: String,
});

const router = useRouter();
const dashboardStore = useDashboardStore();

const isConsolidatedView = computed(() => dashboardStore.selectedConnectionId === 0);

const navigateToDrilldown = (docType, title) => {
  if (isConsolidatedView.value) {
    return;
  }
  
  router.push({
    name: 'TransactionList',
    params: { type: docType },
    query: {
        startDate: props.startDate,
        endDate: props.endDate,
        title: title,
    }
  });
};
</script>

<style scoped>
.drilldown-card {
  cursor: pointer;
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
}
.drilldown-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 24px rgba(37, 54, 119, 0.3);
}
.drilldown-row {
  cursor: pointer;
  transition: background-color 0.2s ease;
}
.drilldown-row:hover {
  background-color: var(--bs-tertiary-bg);
}
</style>