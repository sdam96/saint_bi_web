<template>
  <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4 mt-3">
    <DashboardCard title="Ventas y Utilidad (Últimos 30 días)">
      <li class="list-group-item d-flex justify-content-between"><span>Total Ventas Netas:</span> <strong>{{ formatNumber(summary.TotalNetSales) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Ventas de Contado:</span> <strong>{{ formatNumber(summary.TotalNetSalesCash) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Ventas a Crédito:</span> <strong>{{ formatNumber(summary.TotalNetSalesCredit) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Costo de Ventas:</span> <strong>{{ formatNumber(summary.CostOfGoodsSold) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Utilidad Bruta:</span> <strong>{{ formatNumber(summary.GrossProfit) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Margen Bruto:</span> <strong>{{ summary.GrossProfitMargin.toFixed(2) }}%</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Ticket Promedio:</span> <strong>{{ formatNumber(summary.AverageTicket) }}</strong></li>
    </DashboardCard>

    <DashboardCard title="Cuentas por Cobrar">
      <li class="list-group-item d-flex justify-content-between"><span>Total por Cobrar:</span> <strong>{{ formatNumber(summary.TotalReceivables) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Monto Vencido:</span> <strong>{{ formatNumber(summary.OverdueReceivables) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Porcentaje Vencido:</span> <strong>{{ summary.ReceivablePercentage.toFixed(2) }}%</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Días en la Calle:</span> <strong>{{ summary.ReceivablesTurnoverDays.toFixed(0) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Clientes Activos con Deuda:</span> <strong>{{ summary.ActiveClientsWithDebt }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Clientes con Deuda Vencida:</span> <strong>{{ summary.TotalClientsWithOverdue }}</strong></li>
    </DashboardCard>

    <DashboardCard title="Cuentas por Pagar">
      <li class="list-group-item d-flex justify-content-between"><span>Total por Pagar:</span> <strong>{{ formatNumber(summary.TotalPayables) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Monto Vencido:</span> <strong>{{ formatNumber(summary.OverduePayables) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Días de Compras por Pagar:</span> <strong>{{ summary.PayablesTurnoverDays.toFixed(0) }}</strong></li>
    </DashboardCard>
    
    <DashboardCard title="Totales Generales">
      <li class="list-group-item d-flex justify-content-between"><span>Total Facturas Emitidas:</span> <strong>{{ summary.TotalInvoices }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Clientes Activos:</span> <strong>{{ summary.TotalActiveClients }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Productos Activos:</span> <strong>{{ summary.TotalActiveProducts }}</strong></li>
    </DashboardCard>

    <DashboardCard title="Impuestos y Retenciones">
      <li class="list-group-item d-flex justify-content-between"><span>IVA Débito Fiscal (Ventas):</span> <strong>{{ formatNumber(summary.SalesVAT) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>IVA Crédito Fiscal (Compras):</span> <strong>{{ formatNumber(summary.PurchasesVAT) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>Total IVA por Pagar:</span> <strong>{{ formatNumber(summary.VATPayable) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>IVA Retenido por Clientes:</span> <strong>{{ formatNumber(summary.SalesIVAWithheld) }}</strong></li>
      <li class="list-group-item d-flex justify-content-between"><span>IVA Retenido a Proveedores:</span> <strong>{{ formatNumber(summary.PurchasesIVAWithheld) }}</strong></li>
    </DashboardCard>

    <RankList title="Top 5 Productos por Venta" :items="summary.Top5ProductsBySales" />
    <RankList title="Top 5 Productos por Utilidad" :items="summary.Top5ProductsByProfit" />
    <RankList title="Top 5 Clientes por Venta" :items="summary.Top5ClientsBySales" />
    <RankList title="Top 5 Vendedores por Venta" :items="summary.Top5SellersBySales" />

  </div>
</template>

<script setup>
import DashboardCard from './DashboardCard.vue';
import RankList from './RankList.vue';

// Helper para formatear números con 2 decimales y separadores de miles
const formatNumber = (value) => {
  return new Intl.NumberFormat('es-VE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value);
};

defineProps({
  summary: Object,
});
</script>