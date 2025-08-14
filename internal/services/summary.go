package services

import (
	"log"
	"sort"
	"sync"
	"time"

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/models"
)

// CalculateManagementSummary obtiene los datos y calcula los KPIs.
func CalculateManagementSummary(client *api.SaintClient) (*models.ManagementSummary, error) {
	// --- Paso 1: Obtener todos los datos de la API en paralelo ---
	var (
		invoices     []models.Invoice
		invoiceItems []models.InvoiceItem
		purchases    []models.Purchase
		receivables  []models.AccReceivable
		payables     []models.AccPayable
		products     []models.Product
		customers    []models.Customer
		sellers      []models.Seller
		wg           sync.WaitGroup
		errs         = make(chan error, 8)
	)

	apiCalls := []func() error{
		func() error { var err error; invoices, err = client.GetInvoices(); return err },
		func() error { var err error; invoiceItems, err = client.GetInvoiceItems(); return err },
		func() error { var err error; purchases, err = client.GetPurchases(); return err },
		func() error { var err error; receivables, err = client.GetAccReceivables(); return err },
		func() error { var err error; payables, err = client.GetAccPayables(); return err },
		func() error { var err error; products, err = client.GetProducts(); return err },
		func() error { var err error; customers, err = client.GetCustomers(); return err },
		func() error { var err error; sellers, err = client.GetSellers(); return err },
	}

	for _, call := range apiCalls {
		wg.Add(1)
		go func(apiCall func() error) {
			defer wg.Done()
			if err := apiCall(); err != nil {
				errs <- err
			}
		}(call)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			log.Printf("Error obteniendo datos de la API: %v", err)
			return nil, err
		}
	}

	// --- Paso 2: Realizar los cálculos ---
	summary := &models.ManagementSummary{}
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Crear mapas para facilitar la búsqueda de datos
	invoiceHeaderMap := make(map[string]models.Invoice)
	for _, inv := range invoices {
		if inv.NumeroD != nil {
			invoiceHeaderMap[*inv.NumeroD] = inv
		}
	}

	// Totales de ventas y costos
	for _, inv := range invoices {
		if inv.FechaE != nil {
			if date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE); err == nil && date.After(thirtyDaysAgo) {
				if inv.MtoTotal != nil {
					summary.TotalNetSales += *inv.MtoTotal
				}
				if inv.Credito != nil {
					summary.TotalNetSalesCredit += *inv.Credito
				}
				if inv.Contado != nil {
					summary.TotalNetSalesCash += *inv.Contado
				}
				if inv.CostoPrd != nil {
					summary.CostOfGoodsSold += *inv.CostoPrd
				}
			}
		}
	}
	summary.TotalInvoices = len(invoices)
	if summary.TotalInvoices > 0 {
		summary.AverageTicket = summary.TotalNetSales / float64(summary.TotalInvoices)
	}

	// Utilidad y margen
	summary.GrossProfit = summary.TotalNetSales - summary.CostOfGoodsSold
	if summary.TotalNetSales > 0 {
		summary.GrossProfitMargin = (summary.GrossProfit / summary.TotalNetSales) * 100
	}

	// Cuentas por Cobrar
	for _, r := range receivables {
		if r.Saldo != nil && *r.Saldo > 0 {
			summary.TotalReceivables += *r.Saldo
			if r.FechaV != nil {
				if vencimiento, err := time.Parse("2006-01-02 15:04:05", *r.FechaV); err == nil && vencimiento.Before(now) {
					summary.OverdueReceivables += *r.Saldo
				}
			}
		}
	}
	if summary.TotalNetSalesCredit > 0 {
		summary.ReceivablesTurnoverDays = (summary.TotalReceivables / summary.TotalNetSalesCredit) * 30
	}
	if summary.TotalReceivables > 0 {
		summary.ReceivablePercentage = (summary.OverdueReceivables / summary.TotalReceivables) * 100
	}

	// Cuentas por Pagar
	var totalPurchasesCredit float64
	for _, p := range purchases {
		if p.FechaE != nil {
			if date, err := time.Parse("2006-01-02 15:04:05", *p.FechaE); err == nil && date.After(thirtyDaysAgo) {
				if p.Credito != nil {
					totalPurchasesCredit += *p.Credito
				}
			}
		}
	}
	for _, p := range payables {
		if p.Saldo != nil && *p.Saldo > 0 {
			summary.TotalPayables += *p.Saldo
			if p.FechaV != nil {
				if vencimiento, err := time.Parse("2006-01-02 15:04:05", *p.FechaV); err == nil && vencimiento.Before(now) {
					summary.OverduePayables += *p.Saldo
				}
			}
		}
	}
	if totalPurchasesCredit > 0 {
		summary.PayablesTurnoverDays = (summary.TotalPayables / totalPurchasesCredit) * 30
	}

	// Clientes y productos activos
	for _, c := range customers {
		if c.Activo != nil && *c.Activo == 1 {
			summary.TotalActiveClients++
			if c.Saldo != nil && *c.Saldo > 0 {
				summary.ActiveClientsWithDebt++
			}
		}
	}
	for _, p := range products {
		if p.Activo != nil && *p.Activo == 1 {
			summary.TotalActiveProducts++
		}
	}

	// Impuestos y retenciones
	for _, inv := range invoices {
		if inv.MtoTax != nil {
			summary.SalesVAT += *inv.MtoTax
		}
		if inv.RetenIVA != nil {
			summary.SalesIVAWithheld += *inv.RetenIVA
		}
	}
	for _, p := range purchases {
		if p.MtoTax != nil {
			summary.PurchasesVAT += *p.MtoTax
		}
		if p.RetenIVA != nil {
			summary.PurchasesIVAWithheld += *p.RetenIVA
		}
	}
	summary.VATPayable = summary.SalesVAT - summary.PurchasesVAT

	// --- Rankings Top 5 ---
	summary.Top5ClientsBySales = rankItems(calculateSalesByClient(invoiceItems, invoiceHeaderMap, customers))
	summary.Top5ProductsBySales = rankItems(calculateSalesByProduct(invoiceItems, products))
	summary.Top5SellersBySales = rankItems(calculateSalesBySeller(invoiceItems, invoiceHeaderMap, sellers))
	summary.Top5ProductsByProfit = rankItems(calculateProfitByProduct(invoiceItems, products))

	log.Println("Resumen gerencial completo calculado exitosamente.")
	return summary, nil
}

func calculateSalesByClient(items []models.InvoiceItem, headerMap map[string]models.Invoice, customers []models.Customer) map[string]float64 {
	salesMap := make(map[string]float64)
	nameMap := make(map[string]string)
	for _, c := range customers {
		if c.CodClie != nil && c.Descrip != nil {
			nameMap[*c.CodClie] = *c.Descrip
		}
	}

	for _, item := range items {
		if item.NumeroD == nil || item.TotalItem == nil {
			continue
		}
		if header, ok := headerMap[*item.NumeroD]; ok {
			if header.CodClie != nil {
				if clientName, nameOk := nameMap[*header.CodClie]; nameOk {
					salesMap[clientName] += *item.TotalItem
				}
			}
		}
	}
	return salesMap
}

func calculateSalesBySeller(items []models.InvoiceItem, headerMap map[string]models.Invoice, sellers []models.Seller) map[string]float64 {
	salesMap := make(map[string]float64)
	nameMap := make(map[string]string)
	for _, s := range sellers {
		if s.CodVend != nil && s.Descrip != nil {
			nameMap[*s.CodVend] = *s.Descrip
		}
	}

	for _, item := range items {
		if item.NumeroD == nil || item.TotalItem == nil {
			continue
		}
		if header, ok := headerMap[*item.NumeroD]; ok {
			if header.CodVend != nil {
				if sellerName, nameOk := nameMap[*header.CodVend]; nameOk {
					salesMap[sellerName] += *item.TotalItem
				}
			}
		}
	}
	return salesMap
}

func calculateSalesByProduct(items []models.InvoiceItem, products []models.Product) map[string]float64 {
	salesMap := make(map[string]float64)
	nameMap := make(map[string]string)
	for _, p := range products {
		if p.CodProd != nil && p.Descrip != nil {
			nameMap[*p.CodProd] = *p.Descrip
		}
	}

	for _, item := range items {
		if item.CodItem == nil || item.TotalItem == nil {
			continue
		}
		if productName, ok := nameMap[*item.CodItem]; ok {
			salesMap[productName] += *item.TotalItem
		}
	}
	return salesMap
}

func calculateProfitByProduct(items []models.InvoiceItem, products []models.Product) map[string]float64 {
	profitMap := make(map[string]float64)
	productMap := make(map[string]models.Product)
	for _, p := range products {
		if p.CodProd != nil {
			productMap[*p.CodProd] = p
		}
	}

	for _, item := range items {
		if item.CodItem == nil {
			continue
		}
		if product, ok := productMap[*item.CodItem]; ok {
			if item.Precio != nil && product.CostAct != nil && item.Cantidad != nil && product.Descrip != nil {
				profit := (*item.Precio - *product.CostAct) * *item.Cantidad
				profitMap[*product.Descrip] += profit
			}
		}
	}
	return profitMap
}

func rankItems(itemsMap map[string]float64) []models.RankedItem {
	var ranked []models.RankedItem
	for name, value := range itemsMap {
		ranked = append(ranked, models.RankedItem{Name: name, Value: value})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Value > ranked[j].Value
	})

	if len(ranked) > 5 {
		return ranked[:5]
	}
	return ranked
}
