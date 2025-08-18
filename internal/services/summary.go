// internal/services/summary.go
package services

import (
	"log"
	"sort"
	"sync"
	"time"

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/models"
)

// GetComparativeSummary es el nuevo orquestador principal del servicio.
func GetComparativeSummary(client *api.SaintClient, currentStart, currentEnd, prevStart, prevEnd time.Time) (*models.ComparativeSummary, error) {
	allData, err := fetchAllAPIData(client)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var currentSummary, previousSummary *models.ManagementSummary
	var errCurrent, errPrevious error

	wg.Add(2)
	go func() {
		defer wg.Done()
		currentSummary, errCurrent = calculateSummaryForPeriod(allData, currentStart, currentEnd)
	}()
	go func() {
		defer wg.Done()
		previousSummary, errPrevious = calculateSummaryForPeriod(allData, prevStart, prevEnd)
	}()
	wg.Wait()

	if errCurrent != nil {
		return nil, errCurrent
	}
	if errPrevious != nil {
		return nil, errPrevious
	}

	// CORRECCIÓN: Se cambió la etiqueta JSON incorrecta.
	finalSummary := &models.ComparativeSummary{
		CurrentPeriod:            *currentSummary,
		PreviousPeriod:           *previousSummary,
		TotalNetSalesComparative: calculateComparativeData(currentSummary.TotalNetSales, previousSummary.TotalNetSales),
		GrossProfitComparative:   calculateComparativeData(currentSummary.GrossProfit, previousSummary.GrossProfit),
		AverageTicketComparative: calculateComparativeData(currentSummary.AverageTicket, previousSummary.AverageTicket),
	}

	log.Println("Resumen gerencial comparativo calculado exitosamente.")
	return finalSummary, nil
}

// apiData es una struct interna que actúa como un contenedor para todos los datos.
type apiData struct {
	invoices     []models.Invoice
	invoiceItems []models.InvoiceItem
	purchases    []models.Purchase
	receivables  []models.AccReceivable
	payables     []models.AccPayable
	products     []models.Product
	customers    []models.Customer
	sellers      []models.Seller
}

// fetchAllAPIData ejecuta todas las llamadas a la API de forma concurrente.
func fetchAllAPIData(client *api.SaintClient) (*apiData, error) {
	data := &apiData{}
	var wg sync.WaitGroup
	errs := make(chan error, 8)

	apiCalls := []func() error{
		func() error { var err error; data.invoices, err = client.GetInvoices(); return err },
		func() error { var err error; data.invoiceItems, err = client.GetInvoiceItems(); return err },
		func() error { var err error; data.purchases, err = client.GetPurchases(); return err },
		func() error { var err error; data.receivables, err = client.GetAccReceivables(); return err },
		func() error { var err error; data.payables, err = client.GetAccPayables(); return err },
		func() error { var err error; data.products, err = client.GetProducts(); return err },
		func() error { var err error; data.customers, err = client.GetCustomers(); return err },
		func() error { var err error; data.sellers, err = client.GetSellers(); return err },
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
	return data, nil
}

// calculateSummaryForPeriod contiene la lógica de cálculo completa y corregida.
func calculateSummaryForPeriod(data *apiData, startDate, endDate time.Time) (*models.ManagementSummary, error) {
	summary := &models.ManagementSummary{}
	now := time.Now()

	invoiceHeaderMap := make(map[string]models.Invoice)
	for _, inv := range data.invoices {
		if inv.NumeroD != nil {
			invoiceHeaderMap[*inv.NumeroD] = inv
		}
	}

	var invoicesInPeriod []models.Invoice
	for _, inv := range data.invoices {
		if inv.FechaE != nil {
			if date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE); err == nil && !date.Before(startDate) && !date.After(endDate) {
				invoicesInPeriod = append(invoicesInPeriod, inv)
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
				// CÁLCULO DE IMPUESTOS RESTAURADO
				if inv.MtoTax != nil {
					summary.SalesVAT += *inv.MtoTax
				}
				if inv.RetenIVA != nil {
					summary.SalesIVAWithheld += *inv.RetenIVA
				}
			}
		}
	}

	summary.TotalInvoices = len(invoicesInPeriod)
	if summary.TotalInvoices > 0 {
		summary.AverageTicket = summary.TotalNetSales / float64(summary.TotalInvoices)
	}

	summary.GrossProfit = summary.TotalNetSales - summary.CostOfGoodsSold
	if summary.TotalNetSales > 0 {
		summary.GrossProfitMargin = (summary.GrossProfit / summary.TotalNetSales) * 100
	}

	for _, r := range data.receivables {
		if r.Saldo != nil && *r.Saldo > 0 {
			summary.TotalReceivables += *r.Saldo
			if r.FechaV != nil {
				if vencimiento, err := time.Parse("2006-01-02 15:04:05", *r.FechaV); err == nil && vencimiento.Before(now) {
					summary.OverdueReceivables += *r.Saldo
				}
			}
		}
	}

	daysInRange := endDate.Sub(startDate).Hours() / 24
	if summary.TotalNetSalesCredit > 0 {
		summary.ReceivablesTurnoverDays = (summary.TotalReceivables / summary.TotalNetSalesCredit) * daysInRange
	}
	if summary.TotalReceivables > 0 {
		summary.ReceivablePercentage = (summary.OverdueReceivables / summary.TotalReceivables) * 100
	}

	// --- LÓGICA DE CUENTAS POR PAGAR RESTAURADA ---
	var totalPurchasesCredit float64
	for _, p := range data.purchases {
		if p.FechaE != nil {
			if date, err := time.Parse("2006-01-02 15:04:05", *p.FechaE); err == nil && !date.Before(startDate) && !date.After(endDate) {
				if p.Credito != nil {
					totalPurchasesCredit += *p.Credito
				}
				// CÁLCULO DE IMPUESTOS RESTAURADO
				if p.MtoTax != nil {
					summary.PurchasesVAT += *p.MtoTax
				}
				if p.RetenIVA != nil {
					summary.PurchasesIVAWithheld += *p.RetenIVA
				}
			}
		}
	}

	for _, p := range data.payables {
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
		summary.PayablesTurnoverDays = (summary.TotalPayables / totalPurchasesCredit) * daysInRange
	}
	// --- FIN DE LA LÓGICA RESTAURADA ---

	summary.VATPayable = summary.SalesVAT - summary.PurchasesVAT

	// Conteo de clientes y productos activos (no depende del rango de fechas)
	for _, c := range data.customers {
		if c.Activo != nil && *c.Activo == 1 {
			summary.TotalActiveClients++
			if c.Saldo != nil && *c.Saldo > 0 {
				summary.ActiveClientsWithDebt++
			}
		}
	}
	for _, p := range data.products {
		if p.Activo != nil && *p.Activo == 1 {
			summary.TotalActiveProducts++
		}
	}

	summary.Top5ClientsBySales = rankItems(calculateSalesByClient(data.invoiceItems, invoiceHeaderMap, data.customers))
	summary.Top5ProductsBySales = rankItems(calculateSalesByProduct(data.invoiceItems, data.products))
	summary.Top5SellersBySales = rankItems(calculateSalesBySeller(data.invoiceItems, invoiceHeaderMap, data.sellers))
	summary.Top5ProductsByProfit = rankItems(calculateProfitByProduct(data.invoiceItems, data.products))

	return summary, nil
}

func calculateComparativeData(current, previous float64) models.ComparativeData {
	data := models.ComparativeData{
		Value:         current,
		PreviousValue: previous,
	}
	if previous != 0 {
		data.PercentageChange = ((current - previous) / previous) * 100
	} else if current > 0 {
		data.PercentageChange = 100
	}
	return data
}

// ... (El resto de las funciones auxiliares para los rankings no necesitan cambios)
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
