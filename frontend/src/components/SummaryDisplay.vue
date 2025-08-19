<template>
  <div>
    <div class="row row-cols-1 row-cols-md-3 g-4 mt-3 mb-4">
      <div class="col">
        <KpiCard title="Ventas Netas" :kpi="summary.totalNetSalesComparative" />
      </div>
      <div class="col">
        <KpiCard title="Utilidad Bruta" :kpi="summary.grossProfitComparative" />
      </div>
      <div class="col">
        <KpiCard title="Ticket Promedio" :kpi="summary.averageTicketComparative" />
      </div>
    </div>

    <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4 mt-1">

      <DashboardCard title="Ventas y Utilidad (Período Actual)">
        <li class="list-group-item d-flex justify-content-between"><span>Total Ventas Netas:</span> <strong>{{ formatNumber(summary.currentPeriod.TotalNetSales) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between drilldown-row" @click="navigateToDrilldown('invoices-cash', 'Ventas de Contado del Período')">
          <span>Ventas de Contado:</span> <strong>{{ formatNumber(summary.currentPeriod.TotalNetSalesCash) }}</strong>
        </li>
        <li class="list-group-item d-flex justify-content-between drilldown-row" @click="navigateToDrilldown('invoices-credit', 'Ventas a Crédito del Período')">
          <span>Ventas a Crédito:</span> <strong>{{ formatNumber(summary.currentPeriod.TotalNetSalesCredit) }}</strong>
        </li>
        <li class="list-group-item d-flex justify-content-between"><span>Costo de Ventas:</span> <strong>{{ formatNumber(summary.currentPeriod.CostOfGoodsSold) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Utilidad Bruta:</span> <strong>{{ formatNumber(summary.currentPeriod.GrossProfit) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Margen Bruto:</span> <strong>{{ summary.currentPeriod.GrossProfitMargin.toFixed(2) }}%</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Ticket Promedio:</span> <strong>{{ formatNumber(summary.currentPeriod.AverageTicket) }}</strong></li>
      </DashboardCard>

      <div class="col drilldown-card" @click="navigateToDrilldown('receivables', 'Cuentas por Cobrar del Período')">
        <DashboardCard title="Cuentas por Cobrar">
          <li class="list-group-item d-flex justify-content-between"><span>Total por Cobrar:</span> <strong>{{ formatNumber(summary.currentPeriod.TotalReceivables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Monto Vencido:</span> <strong>{{ formatNumber(summary.currentPeriod.OverdueReceivables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Porcentaje Vencido:</span> <strong>{{ summary.currentPeriod.ReceivablePercentage.toFixed(2) }}%</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Días en la Calle:</span> <strong>{{ summary.currentPeriod.ReceivablesTurnoverDays.toFixed(0) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Clientes Activos con Deuda:</span> <strong>{{ summary.currentPeriod.ActiveClientsWithDebt }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Clientes con Deuda Vencida:</span> <strong>{{ summary.currentPeriod.TotalClientsWithOverdue }}</strong></li>
        </DashboardCard>
      </div>

      <div class="col drilldown-card" @click="navigateToDrilldown('payables', 'Cuentas por Pagar del Período')">
        <DashboardCard title="Cuentas por Pagar">
           <li class="list-group-item d-flex justify-content-between"><span>Total por Pagar:</span> <strong>{{ formatNumber(summary.currentPeriod.TotalPayables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Monto Vencido:</span> <strong>{{ formatNumber(summary.currentPeriod.OverduePayables) }}</strong></li>
          <li class="list-group-item d-flex justify-content-between"><span>Días Compras por Pagar:</span> <strong>{{ summary.currentPeriod.PayablesTurnoverDays.toFixed(0) }}</strong></li>
        </DashboardCard>
      </div>
      
      <DashboardCard title="Totales Generales">
        <li class="list-group-item d-flex justify-content-between"><span>Total Facturas Emitidas:</span> <strong>{{ summary.currentPeriod.TotalInvoices }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Clientes Activos:</span> <strong>{{ summary.currentPeriod.TotalActiveClients }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Productos Activos:</span> <strong>{{ summary.currentPeriod.TotalActiveProducts }}</strong></li>
      </DashboardCard>

      <DashboardCard title="Impuestos y Retenciones">
        <li class="list-group-item d-flex justify-content-between"><span>IVA Débito Fiscal (Ventas):</span> <strong>{{ formatNumber(summary.currentPeriod.SalesVAT) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>IVA Crédito Fiscal (Compras):</span> <strong>{{ formatNumber(summary.currentPeriod.PurchasesVAT) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>Total IVA por Pagar:</span> <strong>{{ formatNumber(summary.currentPeriod.VATPayable) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>IVA Retenido por Clientes:</span> <strong>{{ formatNumber(summary.currentPeriod.SalesIVAWithheld) }}</strong></li>
        <li class="list-group-item d-flex justify-content-between"><span>IVA Retenido a Proveedores:</span> <strong>{{ formatNumber(summary.currentPeriod.PurchasesIVAWithheld) }}</strong></li>
      </DashboardCard>

      <RankList title="Top 5 Productos por Venta" :items="summary.currentPeriod.Top5ProductsBySales" />
      <RankList title="Top 5 Productos por Utilidad" :items="summary.currentPeriod.Top5ProductsByProfit" />
      <RankList title="Top 5 Clientes por Venta" :items="summary.currentPeriod.Top5ClientsBySales" />
      <RankList title="Top 5 Vendedores por Venta" :items="summary.currentPeriod.Top5SellersBySales" />

    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import DashboardCard from './DashboardCard.vue';
import RankList from './RankList.vue';
import KpiCard from './KpiCard.vue';

const props = defineProps({
  summary: Object,
  startDate: String,
  endDate: String,
});

const router = useRouter();

const formatNumber = (value) => {
  if (typeof value !== 'number') return '0,00';
  return new Intl.NumberFormat('es-VE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value);
};

const navigateToDrilldown = (docType, title) => {
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
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.drilldown-row {
  cursor: pointer;
  transition: background-color 0.2s ease;
}
.drilldown-row:hover {
  background-color: #e9ecef;
}
</style>