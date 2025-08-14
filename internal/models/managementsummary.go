package models

import "time"

// ManagementSummary contiene todos los KPIs calculados para el dashboard gerencial.
type ManagementSummary struct {
	// Ventas y Utilidad
	TotalNetSales       float64
	TotalNetSalesCash   float64
	TotalNetSalesCredit float64
	CostOfGoodsSold     float64
	GrossProfit         float64
	GrossProfitMargin   float64
	AverageTicket       float64
	TotalInvoices       int
	TotalActiveProducts int
	TotalActiveClients  int

	// Cuentas por Cobrar
	TotalReceivables        float64
	OverdueReceivables      float64
	ReceivablesTurnoverDays float64
	ReceivablePercentage    float64
	ActiveClientsWithDebt   int
	TotalClientsWithOverdue int

	// Cuentas por Pagar
	TotalPayables        float64
	OverduePayables      float64
	PayablesTurnoverDays float64

	// Impuestos y Retenciones
	SalesVAT             float64
	PurchasesVAT         float64
	VATPayable           float64 // Diferencia entre IVA de ventas y compras
	SalesIVAWithheld     float64 // IVA Retenido por clientes
	PurchasesIVAWithheld float64 // IVA Retenido a proveedores

	// Ranking y Top 5
	Top5ClientsBySales   []RankedItem
	Top5ProductsBySales  []RankedItem
	Top5ProductsByProfit []RankedItem
	Top5SellersBySales   []RankedItem
}

// RankedItem es una estructura genérica para los rankings del dashboard.
type RankedItem struct {
	Name  string
	Value float64
}

// Helper para el cálculo de fechas
var thirtyDaysAgo = time.Now().AddDate(0, 0, -30)
